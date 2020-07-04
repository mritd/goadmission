package route

import (
	"fmt"
	"net/http"
)

func init() {
	RegisterHandle(HandleFunc{
		Path:   "/",
		Method: http.MethodGet,
		Func: func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, "## AvailableRoutes\n\n")
			for k := range funcMap {
				_, _ = fmt.Fprintln(w, "=> "+k)
			}
		},
	})
}
