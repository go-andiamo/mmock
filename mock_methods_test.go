package mmock

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestMockedMethods(t *testing.T) {
	mocked := new(mockedMy)
	mocked.OnMethod(mocked.DoSomething).Return(&SomeStruct{SomeValue: "a"}, nil)

	r, err := mocked.DoSomething("x", 1)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "a", r.SomeValue)
	mocked.AssertMethodCalled(t, mocked.DoSomething)
	mocked.AssertNumberOfMethodCalls(t, mocked.DoSomething, 1)
	mocked.AssertMethodNotCalled(t, mocked.DoSomethingElse)

	mocked.OnMethod(mocked.DoSomethingElse).Return(SomeStruct{SomeValue: "a"}, nil)
	r2, err := mocked.DoSomethingElse("y", 2)
	assert.NoError(t, err)
	assert.Equal(t, "a", r2.SomeValue)
	mocked.AssertMethodCalled(t, mocked.DoSomethingElse)
	mocked.AssertNumberOfMethodCalls(t, mocked.DoSomethingElse, 1)
}

func TestMockedMethods_PanicsOnUnknownMethod(t *testing.T) {
	mocked := new(mockedMy)
	mocked.MockOf = &mockedMy{}

	oth := &anotherMock{}
	assert.Panics(t, func() {
		mocked.OnMethod(oth.OtherMockMethod).Return()
	})
}

func TestMockedMethods_ByNameString(t *testing.T) {
	mocked := NewMockOf[mockedMy, my]()
	mocked.OnMethod("DoSomething", mock.Anything, mock.Anything).Return(&SomeStruct{SomeValue: "a"}, nil)

	r, err := mocked.DoSomething("x", 1)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "a", r.SomeValue)
	mocked.AssertMethodCalled(t, "DoSomething")
	mocked.AssertNumberOfMethodCalls(t, "DoSomething", 1)
	mocked.AssertMethodNotCalled(t, "DoSomethingElse")

	mocked.OnMethod("DoSomethingElse", mock.Anything, mock.Anything).Return(SomeStruct{SomeValue: "a"}, nil)
	r2, err := mocked.DoSomethingElse("y", 2)
	assert.NoError(t, err)
	assert.Equal(t, "a", r2.SomeValue)
	mocked.AssertMethodCalled(t, "DoSomethingElse")
	mocked.AssertNumberOfMethodCalls(t, "DoSomethingElse", 1)

	// now with method name string but without having to specify args...
	mocked = new(mockedMy)
	mocked.MockOf = &mockedMy{}
	mocked.OnMethod("DoSomething").Return(&SomeStruct{SomeValue: "a"}, nil)

	r, err = mocked.DoSomething("x", 1)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "a", r.SomeValue)
	mocked.AssertMethodCalled(t, "DoSomething")
	mocked.AssertMethodCalled(t, "DoSomething", mock.Anything, mock.Anything)
	mocked.AssertMethodNotCalled(t, "DoSomething", mock.Anything, 2)
	mocked.AssertNumberOfMethodCalls(t, "DoSomething", 1)
	mocked.AssertMethodNotCalled(t, "DoSomethingElse")
	// and panics with unknown method name...
	assert.Panics(t, func() {
		mocked.OnMethod("Unknown method").Return(nil)
	})

	mocked = &mockedMy{}
	mocked.OnMethod("DoSomething", mock.Anything, mock.Anything).Return(&SomeStruct{SomeValue: "a"}, nil)
	r, err = mocked.DoSomething("x", 1)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "a", r.SomeValue)
	mocked.AssertMethodNotCalled(t, "DoSomething")
	mocked.AssertMethodCalled(t, "DoSomething", "x", 1)
}

func TestMockedMethods_Panics(t *testing.T) {
	mocked := new(mockedMy)
	assert.Panics(t, func() {
		mocked.OnMethod(false) // method should be a string or method func
	})
}

func TestMockedMethods_OnAllMethods(t *testing.T) {
	mocked := new(mockedMy)
	mocked.MockOf = &mockedMy{}

	mocked.OnAllMethods(false)
	r, err := mocked.DoSomething("a", 1)
	assert.NoError(t, err)
	assert.Nil(t, r)
	mocked.AssertMethodCalled(t, mocked.DoSomething)
	_, err = mocked.DoSomethingElse("b", 2)
	assert.NoError(t, err)
	mocked.AssertMethodCalled(t, mocked.DoSomethingElse)

	// all errors...
	mocked = new(mockedMy)
	mocked.MockOf = &mockedMy{}
	mocked.OnAllMethods(true)
	_, err = mocked.DoSomething("a", 1)
	assert.Error(t, err)
	mocked.AssertMethodCalled(t, mocked.DoSomething)
	_, err = mocked.DoSomethingElse("b", 2)
	assert.Error(t, err)
	mocked.AssertMethodCalled(t, mocked.DoSomethingElse)
}

func TestMockedMethods_OnAllMethods_Panics(t *testing.T) {
	mocked := new(mockedMy)

	assert.Panics(t, func() {
		mocked.OnAllMethods(false)
	})
}

func TestNewMock(t *testing.T) {
	m := NewMock[mockedMy]()
	assert.NotNil(t, m)
	m.OnAllMethods(true)
	r, err := m.DoSomething("a", 1)
	assert.Nil(t, r)
	assert.Error(t, err)
	m.AssertMethodCalled(t, m.DoSomething)

	assert.Panics(t, func() {
		// panics because anotherMock does not have field MockMethods
		NewMock[anotherMock]()
	})
}

func TestNewMockOf(t *testing.T) {
	m := NewMockOf[mockedMy, my]()
	assert.NotNil(t, m)
	m.OnAllMethods(true)
	r, err := m.DoSomething("a", 1)
	assert.Nil(t, r)
	assert.Error(t, err)
	m.AssertMethodCalled(t, m.DoSomething)

	type otherMockedImpl struct {
		MockMethods
	}
	assert.Panics(t, func() {
		// panics because OtherMockedImpl does not implement my
		NewMockOf[otherMockedImpl, my]()
	})
}

type SomeStruct struct {
	SomeValue string
}

// my is the interface to be mocked
type my interface {
	DoSomething(s string, i int) (*SomeStruct, error)
	DoSomethingElse(s string, i int) (SomeStruct, error)
}

// make sure the mock implements the interface
var _ my = &mockedMy{}

// mockedMy is the mocked my implementation
type mockedMy struct {
	MockMethods // use this instead of mock.Mock
}

func (mm *mockedMy) DoSomething(s string, i int) (*SomeStruct, error) {
	args := mm.Called(s, i)
	return As[*SomeStruct](args, 0), As[error](args, 1)
}

func (mm *mockedMy) DoSomethingElse(s string, i int) (SomeStruct, error) {
	args := mm.Called(s, i)
	return As[SomeStruct](args, 0), As[error](args, 1)
}

type anotherMock struct {
}

func (o *anotherMock) OtherMockMethod() {
}
