package mmock

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSpyMockOf(t *testing.T) {
	underlying := &underlyingFull{
		calls: map[string]int{},
	}
	spy := NewSpyMockOf[mockedMy, my](underlying)
	_, err := spy.DoSomething("x", 1)
	assert.Error(t, err)
	_, err = spy.DoSomething("x", 1)
	assert.Error(t, err)
	spy.AssertMethodCalled(t, spy.DoSomething)
	spy.AssertNumberOfMethodCalls(t, spy.DoSomething, 2)
	assert.Equal(t, 2, underlying.calls["DoSomething"])
}

func TestNewSpyMockOf_PanicsWithBadMockImpl(t *testing.T) {
	underlying := &underlyingFull{
		calls: map[string]int{},
	}
	type otherMockedImpl struct {
		MockMethods
	}
	assert.Panics(t, func() {
		// panics because OtherMockedImpl does not implement my
		_ = NewSpyMockOf[otherMockedImpl, my](underlying)
	})
}

func TestSpyMock(t *testing.T) {
	underlying := &underlyingMin{
		calls: map[string]int{},
	}
	spy := NewMockOf[mockedMy, my]()
	spy.SetSpyOf(underlying)
	//spy.OnMethod(spy.DoSomething).Return(nil, nil)
	_, err := spy.DoSomething("x", 1)
	assert.Error(t, err)
	_, err = spy.DoSomething("x", 1)
	assert.Error(t, err)
	spy.AssertMethodCalled(t, spy.DoSomething)
	spy.AssertNumberOfMethodCalls(t, spy.DoSomething, 2)
	assert.Equal(t, 2, underlying.calls["DoSomething"])
}

func TestSpyMock_Once(t *testing.T) {
	underlying := &underlyingMin{
		calls: map[string]int{},
	}
	spy := NewMockOf[mockedMy, my]()
	spy.SetSpyOf(underlying)
	spy.OnMethod(spy.DoSomething).Once().Return(nil, nil)
	_, err := spy.DoSomething("x", 1)
	assert.NoError(t, err)
	_, err = spy.DoSomething("x", 1)
	assert.Error(t, err)
	spy.AssertMethodCalled(t, spy.DoSomething)
	spy.AssertNumberOfMethodCalls(t, spy.DoSomething, 2)
	assert.Equal(t, 1, underlying.calls["DoSomething"])
}

func TestSpyMock_MatchingArgs(t *testing.T) {
	underlying := &underlyingMin{
		calls: map[string]int{},
	}
	spy := NewMockOf[mockedMy, my]()
	spy.SetSpyOf(underlying)
	spy.OnMethod(spy.DoSomething, "x").Return(nil, nil)
	_, err := spy.DoSomething("x", 1)
	assert.NoError(t, err)
	_, err = spy.DoSomething("y", 1)
	assert.Error(t, err)
	spy.AssertMethodCalled(t, spy.DoSomething)
	spy.AssertNumberOfMethodCalls(t, spy.DoSomething, 2)
	assert.Equal(t, 1, underlying.calls["DoSomething"])
}

func TestSpyMock_PanicsWithNonImplementedMethod(t *testing.T) {
	underlying := &underlyingMin{
		calls: map[string]int{},
	}
	spy := NewMockOf[mockedMy, my]()
	spy.SetSpyOf(underlying)
	assert.Panics(t, func() {
		_, _ = spy.DoSomethingElse("", 0)
	})

	spy.OnMethod(spy.DoSomethingElse).Return(SomeStruct{}, nil)
	_, err := spy.DoSomethingElse("", 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, underlying.calls["DoSomethingElse"])
}

type underlyingMin struct {
	calls map[string]int
}

func (um *underlyingMin) DoSomething(s string, i int) (*SomeStruct, error) {
	um.calls["DoSomething"] = um.calls["DoSomething"] + 1
	return nil, errors.New("foo")
}

/* Don't implement this method to ensure calling it panics
func (um *underlyingMin) DoSomethingElse(s string, i int) (SomeStruct, error) {
	um.calls["DoSomethingElse"] = um.calls["DoSomethingElse"] + 1
	return SomeStruct{}, nil
}
*/

type underlyingFull struct {
	calls map[string]int
}

func (um *underlyingFull) DoSomething(s string, i int) (*SomeStruct, error) {
	um.calls["DoSomething"] = um.calls["DoSomething"] + 1
	return nil, errors.New("foo")
}

func (um *underlyingFull) DoSomethingElse(s string, i int) (SomeStruct, error) {
	um.calls["DoSomethingElse"] = um.calls["DoSomethingElse"] + 1
	return SomeStruct{}, nil
}
