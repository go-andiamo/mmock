package main

import (
	"errors"
	"github.com/go-andiamo/mmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MyInterface interface {
	DoSomething(a string) (string, error)
	DoSomethingElse(a int) (int, error)
}

type MyTestObject struct {
	mmock.MockMethods
}

func (m *MyTestObject) DoSomething(a string) (string, error) {
	args := m.Called(a)
	return mmock.As[string](args, 0), mmock.As[error](args, 1)
}

func (m *MyTestObject) DoSomethingElse(a int) (int, error) {
	args := m.Called(a)
	return mmock.As[int](args, 0), mmock.As[error](args, 1)
}

func TestMyMock_DoSomething(t *testing.T) {
	mocked := mmock.NewMockOf[MyTestObject, MyInterface]()
	mocked.OnMethod(mocked.DoSomething).Return("", errors.New("foo"))

	_, err := mocked.DoSomething("a")
	assert.Error(t, err)
	mocked.AssertMethodCalled(t, mocked.DoSomething)
}

func TestMyMock_AllMethods(t *testing.T) {
	mocked := mmock.NewMockOf[MyTestObject, MyInterface]()
	mocked.OnAllMethods(true) // all methods return an error (where the method returns an error

	_, err := mocked.DoSomething("a")
	assert.Error(t, err)
	mocked.AssertMethodCalled(t, mocked.DoSomething)
	mocked.AssertMethodNotCalled(t, mocked.DoSomethingElse)

	_, err = mocked.DoSomethingElse(1)
	assert.Error(t, err)
	mocked.AssertMethodCalled(t, mocked.DoSomethingElse)
}
