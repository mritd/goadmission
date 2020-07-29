package adfunc

import (
	"sync"

	"github.com/mritd/goadmission/pkg/zaplogger"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
)

var adfuncOnce sync.Once
var logger *zap.SugaredLogger

type PatchOption string

var (
	PatchOptionAdd     PatchOption = "add"
	PatchOptionRemove  PatchOption = "remove"
	PatchOptionReplace PatchOption = "replace"
	PatchOptionMove    PatchOption = "move"
	PatchOptionCopy    PatchOption = "copy"
	PatchOptionTest    PatchOption = "test"
)

// RFC 6902
type Patch struct {
	Option PatchOption `json:"op"`
	Path   string      `json:"path"`
	Value  interface{} `json:"value,omitempty"`
	From   string      `json:"from,omitempty"`
}

func JSONPatch() *admissionv1.PatchType {
	p := admissionv1.PatchTypeJSONPatch
	return &p
}

func Setup() {
	adfuncOnce.Do(func() {
		logger = zaplogger.NewSugar("adfunc")
	})
}
