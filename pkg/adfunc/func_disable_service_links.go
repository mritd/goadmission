package adfunc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mritd/goadmission/pkg/conf"

	appsv1 "k8s.io/api/apps/v1"

	jsoniter "github.com/json-iterator/go"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	admissionv1 "k8s.io/api/admission/v1"
)

func init() {
	register(AdmissionFunc{
		Type: AdmissionTypeMutating,
		Path: "/disable-service-links",
		Func: disableServiceLinks,
	})
}

// disableServiceLinks auto set enableServiceLinks of the target Deployment to false
// to prevent k8s environment variable injection
func disableServiceLinks(request *admissionv1.AdmissionRequest) (*admissionv1.AdmissionResponse, error) {
	switch request.Kind.Kind {
	case "Deployment":
		var deploy appsv1.Deployment
		err := jsoniter.Unmarshal(request.Object.Raw, &deploy)
		if err != nil {
			errMsg := fmt.Sprintf("[route.Mutating] /disable-service-links: failed to unmarshal object: %v", err)
			logger.Error(errMsg)
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:    http.StatusBadRequest,
					Message: errMsg,
				},
			}, nil
		}

		for label := range deploy.Labels {
			if label == conf.ForceEnableServiceLinksLabel {
				return &admissionv1.AdmissionResponse{
					Allowed: true,
					Result: &metav1.Status{
						Code:    http.StatusOK,
						Message: "success",
					},
				}, nil
			}
		}

		patches := []Patch{
			{
				Option: PatchOptionAdd,
				Path:   "/metadata/annotations",
				Value: map[string]string{
					fmt.Sprintf("disable-service-links-mutatingwebhook-%d.mritd.com", time.Now().Unix()): "true",
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
			errMsg := fmt.Sprintf("[route.Mutating] /disable-service-links: failed to marshal patch: %v", err)
			logger.Error(errMsg)
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:    http.StatusInternalServerError,
					Message: errMsg,
				},
			}, nil
		}

		logger.Infof("[route.Mutating] /disable-service-links: patches: %s", string(patch))
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
		errMsg := fmt.Sprintf("[route.Mutating] /disable-service-links: received wrong kind request: %s, Only support Kind: Deployment", request.Kind.Kind)
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
