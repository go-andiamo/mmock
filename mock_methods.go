package mmock

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"reflect"
	"regexp"
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
	if !setMockOf(r) {
		panic(fmt.Sprintf("type '%s' is not MockMethods (add field mmock.MockMethods)", reflect.TypeOf(r).Elem().String()))
	}
	return r
}

func setMockOf(mocked any) (ok bool) {
	msu, ok := mocked.(mockSetup)
	if ok {
		msu.setMockOf(mocked)
	}
	return
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

type Spying interface {
	// SetSpyOf sets the mock to be a spy mock
	//
	// The wrapped arg is implementation to be spied on - any methods that are called on the mock
	// but have not been expected (by using On or OnMethod) will call this underlying - but you can still assert
	// that the method has been called
	SetSpyOf(wrapped any)
}

type mockSetup interface {
	Spying
	setMockOf(mocked any)
}

func (mm *MockMethods) setMockOf(mocked any) {
	mm.mockOf = mocked
}

func (mm *MockMethods) SetSpyOf(wrapped any) {
	mm.wrapped = wrapped
}

// MockMethods is the replacement for mock.Mock
type MockMethods struct {
	mock.Mock
	mockOf  any
	wrapped any
}

func (mm *MockMethods) Called(arguments ...interface{}) mock.Arguments {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("could not retrieve caller information")
	}
	methodName := parseMethodName(runtime.FuncForPC(pc).Name())
	return mm.MethodCalled(methodName, arguments...)
}

func (mm *MockMethods) MethodCalled(methodName string, arguments ...interface{}) (result mock.Arguments) {
	if mm.wrapped != nil {
		defer func() {
			if r := recover(); r != nil {
				// assuming that panic was raised by testify Mock.MethodCalled?
				if msg, ok := r.(string); ok && strings.Contains(msg, "mock:") {
					result = mm.callWrapped(methodName, arguments...)
				} else {
					panic(r) // some other panic?
				}
			}
		}()
	}
	result = mm.Mock.MethodCalled(methodName, arguments...)
	return
}

func (mm *MockMethods) callWrapped(methodName string, arguments ...interface{}) (result mock.Arguments) {
	ul := reflect.ValueOf(mm.wrapped)
	m := ul.MethodByName(methodName)
	if !m.IsValid() {
		panic(fmt.Sprintf("spy mock .Wrapped does not implement method '%s'", methodName))
	}
	// simulate mock method called by first ensuring that .On() has been set-up...
	// this is so that methods that weren't mocked using .On but called directly into wrapped can still be asserted to have been called
	outs := make([]any, m.Type().NumOut()) // don't care about the actual return args because they'll never get used
	mm.Mock.On(methodName, arguments...).Once().Return(outs...)
	mm.Mock.MethodCalled(methodName, arguments...)
	// now call the actual underlying wrapped...
	argVs := make([]reflect.Value, len(arguments))
	for i, v := range arguments {
		argVs[i] = reflect.ValueOf(v)
	}
	rArgs := m.Call(argVs)
	for _, ra := range rArgs {
		result = append(result, ra.Interface())
	}
	return
}

// OnAllMethods setups expected calls on every method of the mock
//
// Use the errs arg to specify that methods that return an error should return an error when called
func (mm *MockMethods) OnAllMethods(errs bool) {
	if mm.mockOf == nil {
		panic("cannot mock all methods")
	}
	exms := excludeMethods()
	to := reflect.TypeOf(mm.mockOf)
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
		if mm.mockOf == nil {
			return methodName, -1
		}
		if m, ok := reflect.TypeOf(mm.mockOf).MethodByName(methodName); ok {
			return methodName, m.Type.NumIn() - 1
		}
		panic(fmt.Sprintf("method '%s' does not exist", methodName))
	} else if to.Kind() != reflect.Func {
		panic("not a method")
	}

	fn := parseMethodName(runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name())
	if mm.mockOf != nil {
		if _, ok := reflect.TypeOf(mm.mockOf).MethodByName(fn); !ok {
			panic(fmt.Sprintf("method '%s' does not exist", fn))
		}
	}
	return fn, to.NumIn()
}

var gccRegex = regexp.MustCompile("\\.pN\\d+_")

func parseMethodName(methodName string) string {
	// Code from original testify mock...
	// Next four lines are required to use GCCGO function naming conventions.
	// For Ex:  github_com_docker_libkv_store_mock.WatchTree.pN39_github_com_docker_libkv_store_mock.Mock
	// uses interface information unlike golang github.com/docker/libkv/store/mock.(*Mock).WatchTree
	// With GCCGO we need to remove interface information starting from pN<dd>.
	if gccRegex.MatchString(methodName) {
		methodName = gccRegex.Split(methodName, -1)[0]
	}
	parts := strings.Split(methodName, ".")
	methodName = parts[len(parts)-1]
	if i := strings.Index(methodName, "-"); i != -1 {
		methodName = methodName[:i]
	}
	return methodName
}
