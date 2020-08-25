package adfunc

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mritd/goadmission/pkg/conf"

	jsoniter "github.com/json-iterator/go"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	admissionv1 "k8s.io/api/admission/v1"
)

var renameOnce sync.Once
var renameMap map[string]string

func init() {
	register(AdmissionFunc{
		Type: AdmissionTypeMutating,
		Path: "/rename",
		Func: rename,
	})
}

// rename auto modify the image name of the pod
func rename(request *admissionv1.AdmissionRequest) (*admissionv1.AdmissionResponse, error) {
	// init rename rules map
	renameOnce.Do(func() {
		renameMap = make(map[string]string, 10)
		for _, s := range conf.ImageRename {
			ss := strings.Split(s, "=")
			if len(ss) != 2 {
				logger.Fatalf("failed to parse image name rename rules: %s", s)
			}
			renameMap[ss[0]] = ss[1]
		}
	})

	switch request.Kind.Kind {
	case "Pod":
		var pod corev1.Pod
		err := jsoniter.Unmarshal(request.Object.Raw, &pod)
		if err != nil {
			errMsg := fmt.Sprintf("[route.Mutating] /rename: failed to unmarshal object: %v", err)
			logger.Error(errMsg)
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:    http.StatusBadRequest,
					Message: errMsg,
				},
			}, nil
		}

		// skip static pod
		for k := range pod.Annotations {
			if k == "kubernetes.io/config.mirror" {
				errMsg := fmt.Sprintf("[route.Mutating] /rename: pod %s has kubernetes.io/config.mirror annotation, skip image rename", pod.Name)
				logger.Warn(errMsg)
				return &admissionv1.AdmissionResponse{
					Allowed: true,
					Result: &metav1.Status{
						Code:    http.StatusOK,
						Message: errMsg,
					},
				}, nil
			}
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
						Path:   "/metadata/annotations",
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
			logger.Error(errMsg)
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:    http.StatusInternalServerError,
					Message: errMsg,
				},
			}, nil
		}

		logger.Infof("[route.Mutating] /rename: patches: %s", string(patch))
		return &admissionv1.AdmissionResponse{
			Allowed:   true,
			Patch:     patch,
			PatchType: JSONPatch(),
			Result: &metav1.Status{
				Code:    http.StatusOK,
				Message: "success",
			},
		}, nil
	default:
		errMsg := fmt.Sprintf("[route.Mutating] /rename: received wrong kind request: %s, Only support Kind: Pod", request.Kind.Kind)
		logger.Error(errMsg)
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Code:    http.StatusForbidden,
				Message: errMsg,
			},
		}, nil
	}
}
