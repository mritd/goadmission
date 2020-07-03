package internalmux

import (
	"sync"

	"github.com/gorilla/mux"
)

var Router *mux.Router
var routerOnce sync.Once

func Init() {
	routerOnce.Do(func() {
		Router = mux.NewRouter().StrictSlash(true)
	})
}
