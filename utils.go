package mmock

import "github.com/stretchr/testify/mock"

func As[T any](args mock.Arguments, index int) T {
	r := args.Get(index)
	if r != nil {
		return r.(T)
	}
	var rn T
	return rn
}

func As1[T1 any](args mock.Arguments) T1 {
	return As[T1](args, 0)
}

func As2[T1 any, T2 any](args mock.Arguments) (T1, T2) {
	return As[T1](args, 0), As[T2](args, 1)
}

func As3[T1 any, T2 any, T3 any](args mock.Arguments) (T1, T2, T3) {
	return As[T1](args, 0), As[T2](args, 1), As[T3](args, 2)
}

func As4[T1 any, T2 any, T3 any, T4 any](args mock.Arguments) (T1, T2, T3, T4) {
	return As[T1](args, 0), As[T2](args, 1), As[T3](args, 2), As[T4](args, 3)
}
