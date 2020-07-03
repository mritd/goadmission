package routers

import (
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/mritd/goadmission/pkg/internalmux"

	"github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/apis/admission"
)

type AdmissionFunc struct {
	Path string
	Func func(admissionReview *admission.AdmissionReview) (*admission.AdmissionResponse, error)
}

type admissionFunc struct {
	Path string
	Func func(w http.ResponseWriter, r *http.Request)
}

var adfs []admissionFunc
var routerOnce sync.Once
var deserializer runtime.Decoder

func register(af AdmissionFunc) {
	for _, r := range adfs {
		if af.Path == "" {
			logrus.Fatalf("admission func name is empty")
		}
		if strings.ToLower(af.Path) == strings.ToLower(r.Path) {
			logrus.Fatalf("admission func [%s] already registered", af.Path)
		}
		if af.Path == "" {
			logrus.Debugf("admission func [%s] use name as default path", af.Path)
			af.Path = strings.ToLower(af.Path)
		}
	}
	adfs = append(adfs, admissionFunc{
		Path: strings.ToLower(af.Path),
		Func: func(w http.ResponseWriter, r *http.Request) {
			defer func() { _ = r.Body.Close() }()
			respReview := &admission.AdmissionReview{
				Response: &admission.AdmissionResponse{},
			}
			w.Header().Set("Content-Type", "application/json")
			respBs, err := ioutil.ReadAll(r.Body)
		},
	})
}

func Setup() {
	routerOnce.Do(func() {
		logrus.Info("init deserializer...")
		deserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()

		logrus.Info("init router...")
		for _, r := range adfs {
			logrus.Infof("add admission func: %s", r.Path)
			internalmux.Router.HandleFunc(r.Path)
		}
	})
}
