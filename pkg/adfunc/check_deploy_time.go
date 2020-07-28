package adfunc

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/mritd/goadmission/pkg/conf"

	"github.com/mritd/goadmission/pkg/route"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	route.Register(route.AdmissionFunc{
		Type: route.Validating,
		Path: "/check-deploy-time",
		Func: func(review *admissionv1.AdmissionReview) (*admissionv1.AdmissionResponse, error) {
			switch review.Request.Kind.Kind {
			case "Deployment":
				var deploy appsv1.Deployment
				err := jsoniter.Unmarshal(review.Request.Object.Raw, &deploy)
				if err != nil {
					errMsg := fmt.Sprintf("[route.Validating] /check-deploy-time: failed to unmarshal object: %v", err)
					logrus.Error(errMsg)
					return &admissionv1.AdmissionResponse{
						Allowed: false,
						Result: &metav1.Status{
							Code:    http.StatusBadRequest,
							Message: errMsg,
						},
					}, nil
				}
				for label := range deploy.Labels {
					if label == conf.ForceDeployLabel {
						return &admissionv1.AdmissionResponse{
							Allowed: true,
							Result: &metav1.Status{
								Code:    http.StatusOK,
								Message: "success",
							},
						}, nil
					}
				}

				err = checkTime(conf.AllowDeployTime)
				if err != nil {
					return &admissionv1.AdmissionResponse{
						Allowed: false,
						Result: &metav1.Status{
							Code:    http.StatusForbidden,
							Message: err.Error(),
						},
					}, nil
				} else {
					return &admissionv1.AdmissionResponse{
						Allowed: true,
						Result: &metav1.Status{
							Code:    http.StatusOK,
							Message: "success",
						},
					}, nil
				}
			default:
				errMsg := fmt.Sprintf("[route.Validating] /check-deploy-time: received wrong kind request: %s, Only support Kind: Deployment", review.Request.Kind.Kind)
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

func checkTime(allowTime []string) error {
	const timeLayout = "15:04"
	currentTime, _ := time.Parse(timeLayout, time.Now().Format(timeLayout))
	for _, allowStr := range allowTime {
		allowSlc := strings.Split(allowStr, "~")
		if len(allowSlc) != 2 {
			errMsg := fmt.Sprintf("[route.Validating] /check-deploy-time: allow time format is invalid: %s", allowStr)
			logrus.Error(errMsg)
			return errors.New(errMsg)
		}

		startTime, startErr := time.Parse(timeLayout, allowSlc[0])
		if startErr != nil {
			errMsg := fmt.Sprintf("[route.Validating] /check-deploy-time: failed to parse allow time: %s :%v", allowSlc[0], startErr)
			logrus.Error(errMsg)
			return errors.New(errMsg)
		}
		endTime, endErr := time.Parse(timeLayout, allowSlc[1])
		if endErr != nil {
			errMsg := fmt.Sprintf("[route.Validating] /check-deploy-time: failed to parse allow time: %s :%v", allowSlc[0], endErr)
			logrus.Error(errMsg)
			return errors.New(errMsg)
		}
		if currentTime.After(startTime) && currentTime.Before(endTime) {
			return nil
		}
	}

	return fmt.Errorf("[route.Validating] /check-deploy-time: the current time(%s) is not in the range of %v", currentTime.Format(timeLayout), allowTime)
}
