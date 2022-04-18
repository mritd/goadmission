package route

import "net/http"

func init() {
	RegisterHandler(HandleFunc{
		Path:   "/healthz",
		Method: http.MethodGet,
		Func: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		},
	})
	RegisterHandler(HandleFunc{
		Path:   "/health",
		Method: http.MethodGet,
		Func: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		},
	})
}
