package rpc

import (
	"log"
	"reflect"
	"unicode"
	"unicode/utf8"
)

type Service int

type service struct {
	name       string                 // name of service
	rcvr       reflect.Value          // receiver of methods for the service
	typ        reflect.Type           // type of the receiver
	methods    map[string]*methodType // registered methods
	middleware []func(c *Context) error
}

type methodType struct {
	method           reflect.Method
	RequestDataType  reflect.Type
	ResponseDataType reflect.Type
}

var typeOfContext = reflect.TypeOf((*Context)(nil)).Elem()
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

func suitableMethods(typ reflect.Type) map[string]*methodType {
	numMethod := typ.NumMethod()
	methods := make(map[string]*methodType)
	for m := 0; m < numMethod; m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		// Method needs three ins: , *http.request, *custom data.
		if mtype.NumIn() != 3 {
			log.Printf("rpc.Register: methods %q has %d input parameters; needs exactly 3\n", mname, mtype.NumIn())
			continue
		}
		// First arg needs be a pointer to a context
		requestType := mtype.In(1)
		if requestType.Kind() != reflect.Ptr || requestType.Elem() != typeOfContext {
			log.Printf("rpc.Register: first argument of methods %q needs to be %q not: %q\n", mname,typeOfContext, requestType.Elem())
			continue
		}
		// Second arg must be a pointer.
		dataType := mtype.In(2)
		if dataType.Kind() != reflect.Ptr {
			log.Printf("rpc.Register:  type of methods %q is not a pointer: %q\n", mname, dataType)
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(dataType) {
			log.Printf("rpc.Register: reply type of methods %q is not exported: %q\n", mname, dataType)
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 2 {
			log.Printf("rpc.Register: methods %q has %d output parameters; needs exactly one\n", mname, mtype.NumOut())
			continue
		}
		// The return type of the methods must be a pointer.
		if returnType := mtype.Out(0); returnType.Kind() != reflect.Ptr {
			log.Printf("rpc.Register: return type of methods %q is %q, must be error\n", mname, returnType)
			continue
		} else {
			// The second return type of the methods must be error.
			if returnType := mtype.Out(1); returnType != typeOfError {
				log.Printf("rpc.Register: return type of methods %q is %q, must be error\n", mname, returnType)
				continue
			}
			methods[mname] = &methodType{method: method, RequestDataType: dataType, ResponseDataType: returnType}
		}

	}
	return methods
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}
