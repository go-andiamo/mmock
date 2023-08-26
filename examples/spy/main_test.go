package main

import (
	"github.com/go-andiamo/mmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpyMockedThingy_MethodNotMockedCallsUnderlying(t *testing.T) {
	// call a method that hasn't been set-up...
	underlying := &ActualThing{}
	spy := mmock.NewSpyMockOf[MockThing, Thingy](underlying)
	err := spy.DoSomething()
	assert.NoError(t, err)
	assert.Equal(t, 1, underlying.calls)
	spy.AssertMethodCalled(t, spy.DoSomething)
	spy.AssertNumberOfMethodCalls(t, spy.DoSomething, 1)
	err = spy.DoSomething()
	assert.Error(t, err)
	assert.Equal(t, 2, underlying.calls)
	spy.AssertMethodCalled(t, spy.DoSomething)
	spy.AssertNumberOfMethodCalls(t, spy.DoSomething, 2)
}

func TestSpyMockedThingy_MethodMocked(t *testing.T) {
	// call a method that has been set-up...
	underlying := &ActualThing{}
	spy := mmock.NewSpyMockOf[MockThing, Thingy](underlying)
	spy.OnMethod(spy.DoSomething).Return(nil)
	err := spy.DoSomething()
	assert.NoError(t, err)
	assert.Equal(t, 0, underlying.calls)
	spy.AssertMethodCalled(t, spy.DoSomething)
	spy.AssertNumberOfMethodCalls(t, spy.DoSomething, 1)
	err = spy.DoSomething()
	assert.NoError(t, err)
	assert.Equal(t, 0, underlying.calls)
	spy.AssertMethodCalled(t, spy.DoSomething)
	spy.AssertNumberOfMethodCalls(t, spy.DoSomething, 2)
}

type MockThing struct {
	mmock.MockMethods
}

func (m *MockThing) DoSomething() error {
	args := m.Called()
	return args.Error(0)
}
