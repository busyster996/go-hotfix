package symbols

import (
	"reflect"
)

var Symbols = map[string]map[string]reflect.Value{}

//go:generate go install github.com/traefik/yaegi/cmd/yaegi@latest
//go:generate yaegi extract github.com/busyster996/go-hotfix

//go:generate yaegi extract github.com/busyster996/go-hotfix/example/handler
