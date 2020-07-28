package adfunc

import (
	"fmt"
	"net/http"
	"time"

	appsv1 "k8s.io/api/apps/v1"

	jsoniter "github.com/json-iterator/go"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sirupsen/logrus"

	"github.com/mritd/goadmission/pkg/route"
	admissionv1 "k8s.io/api/admission/v1"
)

func init() {
	route.Register(route.AdmissionFunc{
		Type: route.Mutating,
		Path: "/disable_service_links",
		Func: func(review *admissionv1.AdmissionReview) (*admissionv1.AdmissionResponse, error) {
			switch review.Request.Kind.Kind {
			case "Deployment":
				var deploy appsv1.Deployment
				err := jsoniter.Unmarshal(review.Request.Object.Raw, &deploy)
				if err != nil {
					errMsg := fmt.Sprintf("[route.Mutating] /disable_service_links: failed to unmarshal object: %v", err)
					logrus.Error(errMsg)
					return &admissionv1.AdmissionResponse{
						Allowed: false,
						Result: &metav1.Status{
							Code:    http.StatusBadRequest,
							Message: errMsg,
						},
					}, nil
				}

				patches := []Patch{
					{
						Option: PatchOptionAdd,
						Path:   "/metadata/annotations",
						Value: map[string]string{
							fmt.Sprintf("disable_service_links-mutatingwebhook-%d.mritd.me", time.Now().Unix()): "true",
						},
					},
					{
						Option: PatchOptionReplace,
						Path:   "/spec/template/spec/enableServiceLinks",
						Value:  false,
					},
				}

				patch, err := jsoniter.Marshal(patches)
				if err != nil {
					errMsg := fmt.Sprintf("[route.Mutating] /disable_service_links: failed to marshal patch: %v", err)
					logrus.Error(errMsg)
					return &admissionv1.AdmissionResponse{
						Allowed: false,
						Result: &metav1.Status{
							Code:    http.StatusInternalServerError,
							Message: errMsg,
						},
					}, nil
				}

				logrus.Infof("[route.Mutating] /disable_service_links: patches: %s", string(patch))
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
				errMsg := fmt.Sprintf("[route.Mutating] /disable_service_links: received wrong kind request: %s, Only support Kind: Deployment", review.Request.Kind.Kind)
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
