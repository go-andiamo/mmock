package stuff

import "context"

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

type SomeStruct struct {
}
