package hotfix

import (
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/traefik/yaegi/stdlib/syscall"
	"github.com/traefik/yaegi/stdlib/unrestricted"
	"github.com/traefik/yaegi/stdlib/unsafe"
)

type FuncPatch struct {
	StructType reflect.Type
	FuncName   string
	FuncValue  reflect.Value
}

var (
	convertFuncPatchErr   = errors.New("convert FuncPatch error")
	retrieveMethodNameErr = errors.New("retrieve method by name failed")
)

type patch struct {
	filePath string
	evalText string
	symbols  interp.Exports
}

func ApplyFunc(filePath string, evalText string, symbols interp.Exports) (*gomonkey.Patches, error) {
	p := &patch{
		filePath: filePath,
		evalText: evalText,
		symbols:  symbols,
	}

	return p.applyPatch()
}

func (p *patch) applyPatch() (patches *gomonkey.Patches, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v \n %s", r, string(debug.Stack()))
		}
	}()
	var res reflect.Value
	res, err = p.loadPatch()
	if err != nil {
		return
	}
	fp, ok := res.Interface().(*FuncPatch)
	if !ok {
		err = convertFuncPatchErr
		return
	}
	return p.monkeyFunc(fp.StructType, fp.FuncName, fp.FuncValue)
}

func (p *patch) loadPatch() (reflect.Value, error) {
	interpreter := interp.New(interp.Options{})
	for _, symbol := range []interp.Exports{
		stdlib.Symbols,
		unsafe.Symbols,
		syscall.Symbols,
		unrestricted.Symbols,
		interp.Symbols,
	} {
		if err := interpreter.Use(symbol); err != nil {
			return reflect.Value{}, err
		}
	}
	if err := interpreter.Use(p.symbols); err != nil {
		return reflect.Value{}, err
	}
	if _, err := interpreter.EvalPath(p.filePath); err != nil {
		return reflect.Value{}, err
	}
	return interpreter.Eval(p.evalText)
}

func (p *patch) monkeyFunc(source reflect.Type, methodName string, dest reflect.Value) (*gomonkey.Patches, error) {
	m, ok := source.MethodByName(methodName)
	if !ok {
		return nil, retrieveMethodNameErr
	}
	patches := gomonkey.NewPatches()
	return patches.ApplyCore(m.Func, dest), nil
}
