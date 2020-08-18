package route

import (
	"fmt"
	"net/http"
	"sort"
)

func init() {
	RegisterHandler(HandleFunc{
		Path:   "/",
		Method: http.MethodGet,
		Func:   available,
	})
	RegisterHandler(HandleFunc{
		Path:   "/available",
		Method: http.MethodGet,
		Func:   available,
	})
}

func available(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprint(w, "## AvailableRoutes\n\n")
	keys := make([]string, 0, len(funcMap))
	for k := range funcMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		_, _ = fmt.Fprintln(w, "=> "+k)
	}
}
