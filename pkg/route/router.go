package route

import (
	"fmt"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HandleFunc struct {
	Path   string
	Method string
	Func   func(w http.ResponseWriter, r *http.Request)
}

type handleFuncMap map[string]HandleFunc

var funcMap = make(handleFuncMap, 10)
var routerOnce sync.Once
var logger *zap.SugaredLogger

func RegisterHandler(hf HandleFunc) {
	if hf.Path == "" {
		logger.Fatalf("handle func path is empty")
	}
	registeredHf, ok := funcMap[strings.ToLower(hf.Path)]
	if ok && registeredHf.Method == hf.Method {
		logger.Fatalf("handle func [%s] already registered", hf.Path)
	}
	funcMap[strings.ToLower(hf.Path)] = hf
}

func ResponseErr(handlePath, msg string, httpCode int, w http.ResponseWriter) {
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

	w.Header().Set("Content-Type", "application/json")
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
