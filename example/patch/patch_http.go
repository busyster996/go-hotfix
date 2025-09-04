package patch

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/busyster996/go-hotfix"
	"github.com/busyster996/go-hotfix/example/handler"
)

func PatchTestHandler() *hotfix.FuncPatch {
	log.Println("[Patch] invoke PatchTestHandler()")
	fn := func(h *handler.HttpSvc, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world")
	}
	log.Println("[Patch] invoke PatchTestHandler() end")
	return &hotfix.FuncPatch{
		FuncName:   "Index",
		FuncValue:  reflect.ValueOf(fn),
		StructType: reflect.TypeOf(&handler.HttpSvc{}),
	}
}
