package adfunc

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/mritd/goadmission/pkg/route"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/admission"
)

func init() {
	route.Register(route.AdmissionFunc{
		Path: "/print",
		Func: func(review *admission.AdmissionReview) (*admission.AdmissionResponse, error) {
			bs, err := jsoniter.MarshalIndent(review, "", "    ")
			if err != nil {
				return nil, err
			}
			logrus.Infof("\n%s\n", string(bs))

			return &admission.AdmissionResponse{
				Allowed: true,
				Result:  &metav1.Status{},
			}, nil
		},
	})
}
