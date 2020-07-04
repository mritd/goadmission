package route

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"

	jsoniter "github.com/json-iterator/go"

	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/admission"
)

type AdmissionFunc struct {
	Path string
	Func func(admissionReview *admission.AdmissionReview) (*admission.AdmissionResponse, error)
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

func Register(af AdmissionFunc) {
	if af.Path == "" {
		logrus.Fatalf("admission func path is empty")
	}

	_, ok := funcMap[strings.ToLower(af.Path)]
	if ok {
		logrus.Fatalf("admission func [%s] already registered", af.Path)
	}

	funcMap[strings.ToLower(af.Path)] = HandleFunc{
		Path:   strings.ToLower(af.Path),
		Method: http.MethodPost,
		Func: func(w http.ResponseWriter, r *http.Request) {
			defer func() { _ = r.Body.Close() }()
			w.Header().Set("Content-Type", "application/json")

			reqBs, err := ioutil.ReadAll(r.Body)
			if err != nil {
				responseErr(err.Error(), http.StatusInternalServerError, w)
				return
			}
			if reqBs == nil || len(reqBs) == 0 {
				responseErr("request body is empty", http.StatusBadRequest, w)
				return
			}

			reqReview := admission.AdmissionReview{}
			if _, _, err := deserializer.Decode(reqBs, nil, &reqReview); err != nil {
				responseErr(fmt.Sprintf("failed to decode req: %s", err), http.StatusInternalServerError, w)
				return
			}
			if reqReview.Request == nil {
				responseErr("admission review request is empty", http.StatusBadRequest, w)
				return
			}

			resp, err := af.Func(&reqReview)
			if err != nil {
				responseErr(fmt.Sprintf("admission func response: %s", err), http.StatusForbidden, w)
				return
			}
			if resp == nil {
				responseErr("admission func response is empty", http.StatusInternalServerError, w)
				return
			}
			resp.UID = reqReview.Request.UID
			respReview := admission.AdmissionReview{
				Response: resp,
			}
			respBs, err := jsoniter.Marshal(respReview)
			if err != nil {
				responseErr(fmt.Sprintf("failed to marshal response: %s", err), http.StatusInternalServerError, w)
				logrus.Errorf("the expected response is: %v", respReview)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(respBs)
			logrus.Debugf("write response: %d: %v: %s", http.StatusOK, respReview, err)
		},
	}
}

func RegisterHandle(hf HandleFunc) {
	if hf.Path == "" {
		logrus.Fatalf("handle func path is empty")
	}
	_, ok := funcMap[strings.ToLower(hf.Path)]
	if ok {
		logrus.Fatalf("handle func [%s] already registered", hf.Path)
	}
	funcMap[strings.ToLower(hf.Path)] = hf
}

func responseErr(msg string, httpCode int, w http.ResponseWriter) {
	logrus.Errorf("response err: %s", msg)
	review := &admission.AdmissionReview{
		Response: &admission.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: msg,
			},
		},
	}
	bs, err := jsoniter.Marshal(review)
	if err != nil {
		logrus.Errorf("failed to marshal response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to marshal response: %s", err)))
	}

	w.WriteHeader(httpCode)
	_, err = w.Write(bs)
	logrus.Debugf("write err response: %d: %v: %s", httpCode, review, err)
}

func loggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logrus.Errorf("err: %s, trace: %s", err, debug.Stack())
				}
			}()

			start := time.Now()
			next.ServeHTTP(w, r)
			logrus.Debugf("received request: %s %s %s", time.Since(start), strings.ToLower(r.Method), r.URL.EscapedPath())
		}
		return http.HandlerFunc(fn)
	}
}

var router *mux.Router

func Setup() http.Handler {
	routerOnce.Do(func() {
		logrus.Info("init kube deserializer...")
		deserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()

		logrus.Info("init global http router...")
		router = mux.NewRouter().StrictSlash(true)
		for p, f := range funcMap {
			logrus.Infof("load handle func: %s", p)
			router.HandleFunc(f.Path, f.Func).Methods(f.Method)
		}
	})

	return loggingMiddleware()(router)
}
