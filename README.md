# MMOCK (Method Mock)
[![GoDoc](https://godoc.org/github.com/go-andiamo/mmock?status.svg)](https://pkg.go.dev/github.com/go-andiamo/mmock)
[![Latest Version](https://img.shields.io/github/v/tag/go-andiamo/mmock.svg?sort=semver&style=flat&label=version&color=blue)](https://github.com/go-andiamo/mmock/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-andiamo/mmock)](https://goreportcard.com/report/github.com/go-andiamo/mmock)

mmock (method mock) is an overlay for the popular https://pkg.go.dev/github.com/stretchr/testify/mock 
that allows mocked methods to be specified by function rather than string name.
The problem with specifying methods by string name is that refactoring tools will miss these.

As per this [issue](https://github.com/stretchr/testify/issues/425)

For example, where you might have done...
```go
  myMock.On("DoSomething", mock.Anything).Return(error.New(""))
```
mmock allows you to do...
```go
  myMock.OnMethod(myMock.DoSomething).Return(error.New(""))
```

## Installation

To install mmock, use go get:

    go get github.com/go-andiamo/mmock

To update mmock to the latest version, run:

    go get -u github.com/go-andiamo/mmock

## Usage

To utilise mmock, simply replace the embedded mock instance in your structure.  For example, where you had...  
```go
package main

import "github.com/stretchr/testify/mock"

type MyTestObject struct {
  mock.Mock
}
```
replace with...
```go
package main

import "github.com/go-andiamo/mmock"

type MyTestObject struct {
  mmock.MockMethods
}
```

Your mock will now have all the same assertion, on call etc. methods so no other changes are needed - 
but you'll be able to utilise additional methods to specify mocking methods by function (rather than by name)...
```go
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
```
Things to notice:
* when using `.OnMethod()` you only need to specify as many args that need matching - because
  mmock knows about the method, it can fill in remaining args with `mock.Anything`
* the same is true of `.AssertMethodCalled()` and `.AssertMethodNotCalled()` - unspecified args are filled with `mock.Anything`
* use `.OnAllMethods()` to mock all methods (optionally making all return an error)
* use `mmock.As()` generic function in your mocked methods to return correct types

## Mock generator
Mmock comes with a programmatic mock generator, e.g.
```go
  f, _ := os.Create("internal/mock_thingy.go")
  _ = mmock.MockGenerateFile[internal.Thingy]("", f)
```
or...
```go
  code, _ := mmock.MockGenerate[internal.Thingy]("")
  println(string(code))
```
Caveat emptor! The generated code is designed to save you work but is not 100% guaranteed to produce
compilable code (it can get confused with convoluted or conflicting package names) and may sometimes
require manual intervention.
