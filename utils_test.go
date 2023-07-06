package mmock

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestAs(t *testing.T) {
	args := mock.Arguments{"a", 16, errors.New("")}
	a0 := As[string](args, 0)
	assert.Equal(t, "a", a0)
	a1 := As[int](args, 1)
	assert.Equal(t, 16, a1)
	a2 := As[error](args, 2)
	assert.Error(t, a2)
}

func TestAs2(t *testing.T) {
	args := mock.Arguments{"a", errors.New("")}
	a0, a1 := As2[string, error](args)
	assert.Equal(t, "a", a0)
	assert.Error(t, a1)
}

func TestAs3(t *testing.T) {
	args := mock.Arguments{"a", 16, errors.New("")}
	a0, a1, a2 := As3[string, int, error](args)
	assert.Equal(t, "a", a0)
	assert.Equal(t, 16, a1)
	assert.Error(t, a2)
}

func TestAs4(t *testing.T) {
	args := mock.Arguments{"a", 16, errors.New(""), true}
	a0, a1, a2, a3 := As4[string, int, error, bool](args)
	assert.Equal(t, "a", a0)
	assert.Equal(t, 16, a1)
	assert.Error(t, a2)
	assert.True(t, a3)
}
