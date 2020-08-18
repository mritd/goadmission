package adfunc

import (
	"net/http"

	jsoniter "github.com/json-iterator/go"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	register(AdmissionFunc{
		Type: AdmissionTypeMutating,
		Path: "/print",
		Func: printRequest,
	})

	register(AdmissionFunc{
		Type: AdmissionTypeValidating,
		Path: "/print",
		Func: printRequest,
	})
}

func printRequest(request *admissionv1.AdmissionRequest) (*admissionv1.AdmissionResponse, error) {
	bs, err := jsoniter.MarshalIndent(request, "", "    ")
	if err != nil {
		return nil, err
	}
	logger.Infof("print request: %s", string(bs))

	return &admissionv1.AdmissionResponse{
		Allowed: true,
		Result: &metav1.Status{
			Code:    http.StatusOK,
			Message: "Hello World",
		},
	}, nil
}
