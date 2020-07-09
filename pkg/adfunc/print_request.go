package adfunc

import (
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/mritd/goadmission/pkg/route"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	route.Register(route.AdmissionFunc{
		Path: "/print",
		Func: func(review *admissionv1.AdmissionReview) (*admissionv1.AdmissionResponse, error) {
			bs, err := jsoniter.MarshalIndent(review, "", "    ")
			if err != nil {
				return nil, err
			}
			logrus.Infof("print request: %s", string(bs))

			return &admissionv1.AdmissionResponse{
				Allowed: true,
				Result: &metav1.Status{
					Code:    http.StatusOK,
					Message: "Hello World",
				},
			}, nil
		},
	})
}
