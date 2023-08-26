package mmock

import (
	"fmt"
	"reflect"
)

// NewSpyMockOf creates a new spy mock of a specified type and provides the underling wrapped implementation
//
// The wrapped arg is implementation to be spied on - any methods that are called on the mock
// but have not been expected (by using On or OnMethod) will call this underlying - but you can still assert
// that the method has been called
func NewSpyMockOf[T any, I any](wrapped I) *T {
	r := NewMock[T]()
	if _, ok := interface{}(r).(I); !ok {
		i := new(I)
		panic(fmt.Sprintf("type '%s' does not implement interface '%s'", reflect.TypeOf(r).Elem().String(), reflect.TypeOf(i).Elem().Name()))
	}
	setSpyOf(r, wrapped)
	return r
}

func setSpyOf(mocked any, wrapped any) {
	if msu, ok := mocked.(mockSetup); ok {
		msu.SetSpyOf(wrapped)
	}
}
