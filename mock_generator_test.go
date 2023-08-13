package mmock

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
)

type Thingy interface {
	DoSomething(ctx context.Context, a string) (*SomeStruct, error)
	DoSomethingElse(ctx context.Context, a *SomeStruct) (*SomeStruct, error)
	DoSomethingVars(m *map[string]any, a ...any) *map[*SomeStruct]*SomeStruct
	DoNothing()
	DoNothingWith(a ...any)
	DoNothingWithContext(ctx context.Context, a ...any)
	ReturnSomething() error
	WithVaradicSlices(a ...*[]*string) error
	WithVaradicMaps(a ...*map[any]*string) error
	ManyReturns() (string, int, int64, float64, bool, []string, error)
}

func TestNewMockDef(t *testing.T) {
	md := newMockDef[Thingy]("foo")
	assert.Equal(t, "foo", md.pkg)
	assert.Equal(t, "MockThingy", md.name)
	assert.Equal(t, "Thingy", md.intf)
	assert.Equal(t, 3, len(md.pkgs))
	assert.Equal(t, 10, len(md.fns))
	assert.Equal(t, "DoNothing", md.fns[0].name)
	assert.False(t, md.fns[0].isVaradic)
	assert.Equal(t, 0, len(md.fns[0].ins))
	assert.Equal(t, 0, len(md.fns[0].outs))
	assert.Equal(t, "DoNothingWith", md.fns[1].name)
	assert.Equal(t, 1, len(md.fns[1].ins))
	assert.Equal(t, 0, len(md.fns[1].outs))
	assert.Equal(t, "DoNothingWithContext", md.fns[2].name)
	assert.Equal(t, 2, len(md.fns[2].ins))
	assert.Equal(t, 0, len(md.fns[2].outs))
	assert.Equal(t, "DoSomething", md.fns[3].name)
	assert.Equal(t, 2, len(md.fns[3].ins))
	assert.Equal(t, 2, len(md.fns[3].outs))
	assert.Equal(t, "DoSomethingElse", md.fns[4].name)
	assert.Equal(t, 2, len(md.fns[4].ins))
	assert.Equal(t, 2, len(md.fns[4].outs))
	assert.Equal(t, "DoSomethingVars", md.fns[5].name)
	assert.Equal(t, 2, len(md.fns[5].ins))
	assert.Equal(t, 1, len(md.fns[5].outs))
	assert.Equal(t, "ManyReturns", md.fns[6].name)
	assert.Equal(t, 0, len(md.fns[6].ins))
	assert.Equal(t, 7, len(md.fns[6].outs))
	assert.Equal(t, "ReturnSomething", md.fns[7].name)
	assert.Equal(t, 0, len(md.fns[7].ins))
	assert.Equal(t, 1, len(md.fns[7].outs))
	assert.Equal(t, "WithVaradicMaps", md.fns[8].name)
	assert.True(t, md.fns[8].isVaradic)
	assert.Equal(t, 1, len(md.fns[8].ins))
	assert.Equal(t, 1, len(md.fns[8].outs))
	assert.Equal(t, "WithVaradicSlices", md.fns[9].name)
	assert.True(t, md.fns[9].isVaradic)
	assert.Equal(t, 1, len(md.fns[9].ins))
	assert.Equal(t, 1, len(md.fns[9].outs))
}

func TestMockGenerate(t *testing.T) {
	data, err := MockGenerate[Thingy]("")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(data))
}

func TestMockGenerateFile(t *testing.T) {
	tempPath, err := os.MkdirTemp("", "output")
	require.NoError(t, err)
	fo, err := os.Create(tempPath + "/mock_generate_output.go")
	assert.NoError(t, err)
	err = MockGenerateFile[Thingy]("github.com/example/output/v4", fo)
	assert.NoError(t, err)
	_ = fo.Close()

	fr, err := os.Open(tempPath + "/mock_generate_output.go")
	require.NoError(t, err)
	defer func() {
		_ = fr.Close()
	}()
	data, err := io.ReadAll(fr)
	require.NoError(t, err)
	assert.Equal(t, expectedFile, string(data))
	//println(string(data))
}

func TestPackagePathToPackage(t *testing.T) {
	pkg, core := packagePathToPackage(pkgMmock)
	assert.Equal(t, "mmock", pkg)
	assert.False(t, core)

	pkgPath := "github.com/go-chi/chi/v5"
	pkg, core = packagePathToPackage(pkgPath)
	assert.Equal(t, "chi", pkg)
	assert.False(t, core)

	pkgPath = "context"
	pkg, core = packagePathToPackage(pkgPath)
	assert.Equal(t, "context", pkg)
	assert.True(t, core)

	pkgPath = ""
	pkg, core = packagePathToPackage(pkgPath)
	assert.Equal(t, "", pkg)
	assert.True(t, core)
}

const expectedFile = `package output

import (
	"context"
	"github.com/go-andiamo/mmock"
)

type MockThingy struct {
	mmock.MockMethods
}

func NewMockThingy() *MockThingy {
	return mmock.NewMockOf[MockThingy,Thingy]()
}

// make sure mock implements interface...
var _ Thingy = &MockThingy{}

func (m *MockThingy) DoNothing() {
	m.Called()
}

func (m *MockThingy) DoNothingWith(arg1 ...any) {
	args := make([]any, 0)
	args = append(args, arg1...)
	m.Called(args...)
}

func (m *MockThingy) DoNothingWithContext(arg1 context.Context, arg2 ...any) {
	args := make([]any, 0)
	args = append(args, arg1)
	args = append(args, arg2...)
	m.Called(args...)
}

func (m *MockThingy) DoSomething(arg1 context.Context, arg2 string) (*mmock.SomeStruct, error) {
	retArgs := m.Called(arg1, arg2)
	return mmock.As2[*mmock.SomeStruct, error](retArgs)
}

func (m *MockThingy) DoSomethingElse(arg1 context.Context, arg2 *mmock.SomeStruct) (*mmock.SomeStruct, error) {
	retArgs := m.Called(arg1, arg2)
	return mmock.As2[*mmock.SomeStruct, error](retArgs)
}

func (m *MockThingy) DoSomethingVars(arg1 *map[string]any, arg2 ...any) *map[*mmock.SomeStruct]*mmock.SomeStruct {
	args := make([]any, 0)
	args = append(args, arg1)
	args = append(args, arg2...)
	retArgs := m.Called(args...)
	return mmock.As1[*map[*mmock.SomeStruct]*mmock.SomeStruct](retArgs)
}

func (m *MockThingy) ManyReturns() (string, int, int64, float64, bool, []string, error) {
	retArgs := m.Called()
	return mmock.As[string](retArgs, 0), mmock.As[int](retArgs, 1), mmock.As[int64](retArgs, 2), mmock.As[float64](retArgs, 3), mmock.As[bool](retArgs, 4), mmock.As[[]string](retArgs, 5), mmock.As[error](retArgs, 6)
}

func (m *MockThingy) ReturnSomething() error {
	retArgs := m.Called()
	return mmock.As1[error](retArgs)
}

func (m *MockThingy) WithVaradicMaps(arg1 ...*map[any]*string) error {
	args := make([]any, 0)
	for _, v := range arg1 {
		args = append(args, v)
	}
	retArgs := m.Called(args...)
	return mmock.As1[error](retArgs)
}

func (m *MockThingy) WithVaradicSlices(arg1 ...*[]*string) error {
	args := make([]any, 0)
	for _, v := range arg1 {
		args = append(args, v)
	}
	retArgs := m.Called(args...)
	return mmock.As1[error](retArgs)
}
`

const expected = `package mmock

import (
	"context"
)

type MockThingy struct {
	mmock.MockMethods
}

func NewMockThingy() *MockThingy {
	return mmock.NewMockOf[MockThingy,Thingy]()
}

// make sure mock implements interface...
var _ Thingy = &MockThingy{}

func (m *MockThingy) DoNothing() {
	m.Called()
}

func (m *MockThingy) DoNothingWith(arg1 ...any) {
	args := make([]any, 0)
	args = append(args, arg1...)
	m.Called(args...)
}

func (m *MockThingy) DoNothingWithContext(arg1 context.Context, arg2 ...any) {
	args := make([]any, 0)
	args = append(args, arg1)
	args = append(args, arg2...)
	m.Called(args...)
}

func (m *MockThingy) DoSomething(arg1 context.Context, arg2 string) (*SomeStruct, error) {
	retArgs := m.Called(arg1, arg2)
	return mmock.As2[*SomeStruct, error](retArgs)
}

func (m *MockThingy) DoSomethingElse(arg1 context.Context, arg2 *SomeStruct) (*SomeStruct, error) {
	retArgs := m.Called(arg1, arg2)
	return mmock.As2[*SomeStruct, error](retArgs)
}

func (m *MockThingy) DoSomethingVars(arg1 *map[string]any, arg2 ...any) *map[*SomeStruct]*SomeStruct {
	args := make([]any, 0)
	args = append(args, arg1)
	args = append(args, arg2...)
	retArgs := m.Called(args...)
	return mmock.As1[*map[*SomeStruct]*SomeStruct](retArgs)
}

func (m *MockThingy) ManyReturns() (string, int, int64, float64, bool, []string, error) {
	retArgs := m.Called()
	return mmock.As[string](retArgs, 0), mmock.As[int](retArgs, 1), mmock.As[int64](retArgs, 2), mmock.As[float64](retArgs, 3), mmock.As[bool](retArgs, 4), mmock.As[[]string](retArgs, 5), mmock.As[error](retArgs, 6)
}

func (m *MockThingy) ReturnSomething() error {
	retArgs := m.Called()
	return mmock.As1[error](retArgs)
}

func (m *MockThingy) WithVaradicMaps(arg1 ...*map[any]*string) error {
	args := make([]any, 0)
	for _, v := range arg1 {
		args = append(args, v)
	}
	retArgs := m.Called(args...)
	return mmock.As1[error](retArgs)
}

func (m *MockThingy) WithVaradicSlices(arg1 ...*[]*string) error {
	args := make([]any, 0)
	for _, v := range arg1 {
		args = append(args, v)
	}
	retArgs := m.Called(args...)
	return mmock.As1[error](retArgs)
}
`
