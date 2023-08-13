package mmock

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// NewMock creates a new mock of a specified type
//
// Note: It ensures that the type implements mocks.MockMethods (and panics if it does not!)
//
// Example usage:
//
//	type MockedSomething struct {
//	  mocks.MockMethods
//	}
//	myMock := NewMock[MockedSomething]()
func NewMock[T any]() *T {
	r := new(T)
	if !setMockMethodsMockOf(r) {
		panic(fmt.Sprintf("type '%s' is not MockMethods (add field Mocks.MockMethods)", reflect.TypeOf(r).Elem().String()))
	}
	return r
}

// NewMockOf creates a new mock of a specified type
//
// same as NewMock except that it also checks that the specified type implements the interface specified by the second generic arg (and panics if it does not!)
//
// Example usage:
//
//	type my interface {
//	  SomeMethod()
//	}
//	type MockedSomething struct {
//	  mocks.MockMethods
//	}
//	func (m *MockedSomething) SomeMethod() {
//	}
//	myMock := NewMockOf[MockedSomething, my]()
func NewMockOf[T any, I any]() *T {
	r := NewMock[T]()
	if _, ok := interface{}(r).(I); !ok {
		i := new(I)
		panic(fmt.Sprintf("type '%s' does not implement interface '%s'", reflect.TypeOf(r).Elem().String(), reflect.TypeOf(i).Elem().Name()))
	}
	return r
}

func setMockMethodsMockOf(mocked any) (ok bool) {
	mv := reflect.ValueOf(mocked).Elem()
	for i := 0; !ok && i < mv.NumField(); i++ {
		ok = setMockOfField(mocked, mv.Field(i))
	}
	return
}

var mockMethodsType = reflect.TypeOf(MockMethods{})

func setMockOfField(mocked any, fld reflect.Value) (ok bool) {
	if ft := fld.Type(); ft == mockMethodsType {
		for i := 0; !ok && i < fld.NumField(); i++ {
			if ft.Field(i).Name == mockOfFieldName {
				fld.Field(i).Set(reflect.ValueOf(mocked))
				ok = true
			}
		}
	}
	return
}

const mockOfFieldName = "MockOf"

type MockMethods struct {
	mock.Mock
	MockOf any
}

// OnAllMethods setups expected calls on every method of the mock
//
// Use the errs arg to specify that methods that return an error should return an error when called
func (mm *MockMethods) OnAllMethods(errs bool) {
	if mm.MockOf == nil {
		panic("cannot mock all methods")
	}
	exms := excludeMethods()
	to := reflect.TypeOf(mm.MockOf)
	for i := to.NumMethod() - 1; i >= 0; i-- {
		method := to.Method(i)
		if !exms[method.Name] {
			ins, outs := methodInsAndOuts(method, errs)
			mm.OnMethod(method.Name, ins...).Return(outs...)
		}
	}
}

// OnMethod is the same as Mock.On() (https://pkg.go.dev/github.com/stretchr/testify/mock#Call.On)
//
// Except the method can be specified by func pointer or name
//
//go:noinline
func (mm *MockMethods) OnMethod(method any, arguments ...any) *mock.Call {
	methodName, argCount := mm.getMethodNameAndNumArgs(method)
	for i := argCount - len(arguments); i > 0; i-- {
		arguments = append(arguments, mock.Anything)
	}
	return mm.Mock.On(methodName, arguments...)
}

// AssertNumberOfMethodCalls is the same as Mock.AssertNumberOfCalls() (https://pkg.go.dev/github.com/stretchr/testify/mock#Mock.AssertNumberOfCalls)
//
// Except the method can be specified by func pointer or name
//
//go:noinline
func (mm *MockMethods) AssertNumberOfMethodCalls(t *testing.T, method any, expectedCalls int) bool {
	methodName, _ := mm.getMethodNameAndNumArgs(method)
	return mm.Mock.AssertNumberOfCalls(t, methodName, expectedCalls)
}

// AssertMethodCalled is the same as Mock.AssertCalled() (https://pkg.go.dev/github.com/stretchr/testify/mock#Mock.AssertCalled)
//
// # Except the method can be specified by func pointer or name
//
// Also, if the number of arguments specified is less than the expected args of the method then
// the arguments is padded with mock.Anything
//
//go:noinline
func (mm *MockMethods) AssertMethodCalled(t *testing.T, method any, arguments ...any) bool {
	methodName, ins := mm.getMethodNameAndNumArgs(method)
	for i := ins - len(arguments); i > 0; i-- {
		arguments = append(arguments, mock.Anything)
	}
	return mm.Mock.AssertCalled(t, methodName, arguments...)
}

// AssertMethodNotCalled is the same as Mock.AssertNotCalled() (https://pkg.go.dev/github.com/stretchr/testify/mock#Mock.AssertNotCalled)
//
// # Except the method can be specified by func pointer or name
//
// Also, if the number of arguments specified is less than the expected args of the method then
// the arguments is padded with mock.Anything
//
//go:noinline
func (mm *MockMethods) AssertMethodNotCalled(t *testing.T, method any, arguments ...any) bool {
	methodName, ins := mm.getMethodNameAndNumArgs(method)
	for i := ins - len(arguments); i > 0; i-- {
		arguments = append(arguments, mock.Anything)
	}
	return mm.Mock.AssertNotCalled(t, methodName, arguments...)
}

func methodInsAndOuts(method reflect.Method, errs bool) (ins []any, outs []any) {
	inCount := method.Type.NumIn() - 1
	ins = make([]any, inCount)
	for i := 0; i < inCount; i++ {
		ins[i] = mock.Anything
	}
	outCount := method.Type.NumOut()
	outs = make([]any, outCount)
	for i := 0; i < outCount; i++ {
		if errs {
			ot := method.Type.Out(i)
			if ot.String() == "error" {
				outs[i] = errors.New("")
			} else {
				outs[i] = nil
			}
		} else {
			outs[i] = nil
		}
	}
	return
}

func excludeMethods() map[string]bool {
	result := map[string]bool{}
	to := reflect.TypeOf(&MockMethods{})
	for i := to.NumMethod() - 1; i >= 0; i-- {
		result[to.Method(i).Name] = true
	}
	to = reflect.TypeOf(&mock.Mock{})
	for i := to.NumMethod() - 1; i >= 0; i-- {
		result[to.Method(i).Name] = true
	}
	return result
}

func (mm *MockMethods) getMethodNameAndNumArgs(method any) (string, int) {
	to := reflect.TypeOf(method)
	if to.Kind() == reflect.String {
		methodName := method.(string)
		if mm.MockOf == nil {
			return methodName, -1
		}
		if m, ok := reflect.TypeOf(mm.MockOf).MethodByName(methodName); ok {
			return methodName, m.Type.NumIn() - 1
		}
		panic(fmt.Sprintf("method '%s' does not exist", methodName))
	} else if to.Kind() != reflect.Func {
		panic("not a method")
	}
	fn := runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name()
	fn = fn[strings.LastIndex(fn, ".")+1:]
	if i := strings.Index(fn, "-"); i != -1 {
		fn = fn[:i]
	}
	if mm.MockOf != nil {
		if _, ok := reflect.TypeOf(mm.MockOf).MethodByName(fn); !ok {
			panic(fmt.Sprintf("method '%s' does not exist", fn))
		}
	}
	return fn, to.NumIn()
}
