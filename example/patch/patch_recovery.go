package patch

import (
	"log"
	"net/http"
	"reflect"

	"github.com/busyster996/go-hotfix"
	"github.com/busyster996/go-hotfix/example/handler"
)

func PatchRecoveryHandler() *hotfix.FuncPatch {
	log.Println("[Patch] invoke PatchRecoveryHandler()")
	fn := func(h *handler.HttpSvc, w http.ResponseWriter, r *http.Request) {
		h.Hello(w, r)
	}
	log.Println("[Patch] invoke PatchRecoveryHandler() end")
	return &hotfix.FuncPatch{
		FuncName:   "Index",
		FuncValue:  reflect.ValueOf(fn),
		StructType: reflect.TypeOf(&handler.HttpSvc{}),
	}
}
