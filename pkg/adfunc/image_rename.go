package adfunc

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mritd/goadmission/pkg/conf"
	"github.com/sirupsen/logrus"

	"github.com/mritd/goadmission/pkg/route"
	admissionv1 "k8s.io/api/admission/v1"
)

var renameOnce sync.Once
var renameMap map[string]string

func init() {
	// init rename rules map
	renameOnce.Do(func() {
		renameMap = make(map[string]string, 10)
		for _, s := range conf.ImageRename {
			ss := strings.Split(s, "=")
			if len(ss) != 2 {
				logrus.Fatalf("failed to parse image name rename rules: %s", s)
			}
			renameMap[ss[0]] = ss[1]
		}
	})

	route.Register(route.AdmissionFunc{
		Type: route.Mutating,
		Path: "/rename",
		Func: func(review *admissionv1.AdmissionReview) (*admissionv1.AdmissionResponse, error) {
			switch review.Request.Kind.Kind {
			case "Pod":
				var pod corev1.Pod
				err := jsoniter.Unmarshal(review.Request.Object.Raw, &pod)
				if err != nil {
					errMsg := fmt.Sprintf("[route.Mutating] /rename: failed to unmarshal object: %v", err)
					logrus.Error(errMsg)
					return &admissionv1.AdmissionResponse{
						Allowed: false,
						Result: &metav1.Status{
							Code:    http.StatusBadRequest,
							Message: errMsg,
						},
					}, nil
				}

				var patches []Patch
				for i, c := range pod.Spec.Containers {
					for s, t := range renameMap {
						if strings.HasPrefix(c.Image, s) {
							patches = append(patches, Patch{
								Option: PatchOptionReplace,
								Path:   fmt.Sprintf("/spec/containers/%d/image", i),
								Value:  strings.Replace(c.Image, s, t, 1),
							})

							patches = append(patches, Patch{
								Option: PatchOptionAdd,
								Path:   "/metadata/annotations/",
								Value: map[string]string{
									fmt.Sprintf("rename-mutatingwebhook-%d.mritd.me", time.Now().Unix()): fmt.Sprintf("%d-%s-%s", i, strings.ReplaceAll(s, "/", "_"), strings.ReplaceAll(t, "/", "_")),
								},
							})
							break
						}
					}
				}

				patch, err := jsoniter.Marshal(patches)
				if err != nil {
					errMsg := fmt.Sprintf("[route.Mutating] /rename: failed to marshal patch: %v", err)
					logrus.Error(errMsg)
					return &admissionv1.AdmissionResponse{
						Allowed: false,
						Result: &metav1.Status{
							Code:    http.StatusInternalServerError,
							Message: errMsg,
						},
					}, nil
				}

				logrus.Infof("[route.Mutating] /rename: patches: %v", patches)
				return &admissionv1.AdmissionResponse{
					Allowed:   false,
					Patch:     patch,
					PatchType: JSONPatch(),
					Result: &metav1.Status{
						Code:    http.StatusOK,
						Message: "success",
					},
				}, nil
			default:
				errMsg := fmt.Sprintf("[route.Mutating] /rename: received wrong kind request: %s, Only support Kind: Pod", review.Request.Kind.Kind)
				logrus.Error(errMsg)
				return &admissionv1.AdmissionResponse{
					Allowed: false,
					Result: &metav1.Status{
						Code:    http.StatusForbidden,
						Message: errMsg,
					},
				}, nil
			}
		},
	})
}
