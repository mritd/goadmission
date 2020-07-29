package route

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/mritd/goadmission/pkg/zaplogger"
	"go.uber.org/zap"

	"github.com/gorilla/mux"

	jsoniter "github.com/json-iterator/go"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AdmissionFuncType string

const (
	Mutating   AdmissionFuncType = "Mutating"
	Validating AdmissionFuncType = "Validating"
)

type AdmissionFunc struct {
	Type AdmissionFuncType
	Path string
	Func func(review *admissionv1.AdmissionReview) (*admissionv1.AdmissionResponse, error)
}

type HandleFunc struct {
	Path   string
	Method string
	Func   func(w http.ResponseWriter, r *http.Request)
}

type handleFuncMap map[string]HandleFunc

var funcMap = make(handleFuncMap, 10)
var routerOnce sync.Once
var deserializer runtime.Decoder
var logger *zap.SugaredLogger

func Register(af AdmissionFunc) {
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
	case Mutating:
		handlePath = "/mutating" + handlePath
	case Validating:
		handlePath = "/validating" + handlePath
	default:
		logger.Fatalf("unsupported admission func type")
	}

	if _, ok := funcMap[handlePath]; ok {
		logger.Fatalf("admission func [%s], type: %s already registered", af.Path, af.Type)
	}

	funcMap[handlePath] = HandleFunc{
		Path:   handlePath,
		Method: http.MethodPost,
		Func: func(w http.ResponseWriter, r *http.Request) {
			defer func() { _ = r.Body.Close() }()
			w.Header().Set("Content-Type", "application/json")

			reqBs, err := ioutil.ReadAll(r.Body)
			if err != nil {
				responseErr(handlePath, err.Error(), http.StatusInternalServerError, w)
				return
			}
			if reqBs == nil || len(reqBs) == 0 {
				responseErr(handlePath, "request body is empty", http.StatusBadRequest, w)
				return
			}
			logger.Debugf("request body: %s", string(reqBs))

			reqReview := admissionv1.AdmissionReview{}
			if _, _, err := deserializer.Decode(reqBs, nil, &reqReview); err != nil {
				responseErr(handlePath, fmt.Sprintf("failed to decode req: %s", err), http.StatusInternalServerError, w)
				return
			}
			if reqReview.Request == nil {
				responseErr(handlePath, "admission review request is empty", http.StatusBadRequest, w)
				return
			}

			resp, err := af.Func(&reqReview)
			if err != nil {
				responseErr(handlePath, fmt.Sprintf("admission func response: %s", err), http.StatusForbidden, w)
				return
			}
			if resp == nil {
				responseErr(handlePath, "admission func response is empty", http.StatusInternalServerError, w)
				return
			}
			resp.UID = reqReview.Request.UID
			respReview := admissionv1.AdmissionReview{
				TypeMeta: reqReview.TypeMeta,
				Response: resp,
			}
			respBs, err := jsoniter.Marshal(respReview)
			if err != nil {
				responseErr(handlePath, fmt.Sprintf("failed to marshal response: %s", err), http.StatusInternalServerError, w)
				logger.Errorf("the expected response is: %v", respReview)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(respBs)
			logger.Debugf("write response: %d: %s: %v", http.StatusOK, string(respBs), err)
		},
	}
}

func RegisterHandle(hf HandleFunc) {
	if hf.Path == "" {
		logger.Fatalf("handle func path is empty")
	}
	_, ok := funcMap[strings.ToLower(hf.Path)]
	if ok {
		logger.Fatalf("handle func [%s] already registered", hf.Path)
	}
	funcMap[strings.ToLower(hf.Path)] = hf
}

func responseErr(handlePath, msg string, httpCode int, w http.ResponseWriter) {
	logger.Errorf("handle func [%s] response err: %s", handlePath, msg)
	review := &admissionv1.AdmissionReview{
		Response: &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: msg,
			},
		},
	}
	bs, err := jsoniter.Marshal(review)
	if err != nil {
		logger.Errorf("failed to marshal response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to marshal response: %s", err)))
	}

	w.WriteHeader(httpCode)
	_, err = w.Write(bs)
	logger.Debugf("write err response: %d: %v: %v", httpCode, review, err)
}

func loggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Errorf("err: %v, trace: %s", err, string(debug.Stack()))
				}
			}()

			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Debugf("received request: %s %s %s", time.Since(start), strings.ToLower(r.Method), r.URL.EscapedPath())
		}
		return http.HandlerFunc(fn)
	}
}

var router *mux.Router

func Setup() {
	routerOnce.Do(func() {
		logger = zaplogger.NewSugar("route")
		logger.Info("init kube deserializer...")
		deserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()

		logger.Info("init global http router...")
		router = mux.NewRouter().StrictSlash(true)
		for p, f := range funcMap {
			logger.Infof("load handle func: %s", p)
			router.HandleFunc(f.Path, f.Func).Methods(f.Method)
		}
	})
}

func Router() http.Handler {
	return loggingMiddleware()(router)
}
