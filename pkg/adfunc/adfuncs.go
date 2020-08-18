package adfunc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	jsoniter "github.com/json-iterator/go"

	"github.com/mritd/goadmission/pkg/route"

	"github.com/mritd/goadmission/pkg/zaplogger"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	AdmissionTypeMutating   AdmissionType = "Mutating"
	AdmissionTypeValidating AdmissionType = "Validating"
)

type AdmissionType string

type AdmissionFunc struct {
	Type AdmissionType
	Path string
	Func func(request *admissionv1.AdmissionRequest) (*admissionv1.AdmissionResponse, error)
}

type admissionFuncMap map[string]AdmissionFunc

var funcMap = make(admissionFuncMap, 10)

var adfuncOnce sync.Once
var deserializer runtime.Decoder
var logger *zap.SugaredLogger

func Setup() {
	adfuncOnce.Do(func() {
		logger = zaplogger.NewSugar("adfunc")

		logger.Info("init kube deserializer...")
		deserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()

		logger.Info("init admission func...")
		for p, af := range funcMap {
			logger.Infof("load admission func: %s", af.Path)
			handlePath := strings.Replace(p, "_", "-", -1)
			if p != handlePath {
				logger.Warnf("admission func handler path does not support '_', it has been automatically converted to '-'(%s => %s)", p, handlePath)
			}

			route.RegisterHandler(route.HandleFunc{
				Path:   handlePath,
				Method: http.MethodPost,
				Func: func(w http.ResponseWriter, r *http.Request) {
					defer func() { _ = r.Body.Close() }()

					reqBs, err := ioutil.ReadAll(r.Body)
					if err != nil {
						route.ResponseErr(handlePath, err.Error(), http.StatusInternalServerError, w)
						return
					}
					if reqBs == nil || len(reqBs) == 0 {
						route.ResponseErr(handlePath, "request body is empty", http.StatusBadRequest, w)
						return
					}
					logger.Debugf("request body: %s", string(reqBs))

					reqReview := admissionv1.AdmissionReview{}
					if _, _, err := deserializer.Decode(reqBs, nil, &reqReview); err != nil {
						route.ResponseErr(handlePath, fmt.Sprintf("failed to decode req: %s", err), http.StatusInternalServerError, w)
						return
					}
					if reqReview.Request == nil {
						route.ResponseErr(handlePath, "admission review request is empty", http.StatusBadRequest, w)
						return
					}

					resp, err := af.Func(reqReview.Request)
					if err != nil {
						route.ResponseErr(handlePath, fmt.Sprintf("admission func response: %s", err), http.StatusForbidden, w)
						return
					}
					if resp == nil {
						route.ResponseErr(handlePath, "admission func response is empty", http.StatusInternalServerError, w)
						return
					}
					resp.UID = reqReview.Request.UID
					respReview := admissionv1.AdmissionReview{
						TypeMeta: reqReview.TypeMeta,
						Response: resp,
					}
					respBs, err := jsoniter.Marshal(respReview)
					if err != nil {
						route.ResponseErr(handlePath, fmt.Sprintf("failed to marshal response: %s", err), http.StatusInternalServerError, w)
						logger.Errorf("the expected response is: %v", respReview)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, err = w.Write(respBs)
					logger.Debugf("write response: %d: %s: %v", http.StatusOK, string(respBs), err)
				},
			})
		}

	})
}

func register(af AdmissionFunc) {
	if af.Path == "" {
		logger.Fatalf("admission func path is empty")
	}

	if af.Type == "" {
		logger.Fatalf("admission func type is empty")
	}

	handlePath := strings.ToLower(af.Path)
	if !strings.HasPrefix(handlePath, "/") {
		handlePath = "/" + handlePath
	}
	switch af.Type {
	case AdmissionTypeMutating:
		handlePath = "/mutating" + handlePath
	case AdmissionTypeValidating:
		handlePath = "/validating" + handlePath
	default:
		logger.Fatalf("unsupported admission func type")
	}

	if _, exist := funcMap[handlePath]; exist {
		logger.Fatalf("admission func [%s], type: %s already registered", af.Path, af.Type)
	}

	funcMap[handlePath] = af
}
