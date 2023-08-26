package main

import "errors"

type Thingy interface {
	DoSomething() error
}

var _ Thingy = &ActualThing{}

type ActualThing struct {
	calls int
}

func (a *ActualThing) DoSomething() error {
	a.calls++
	if a.calls > 1 {
		return errors.New("can't do something more than once")
	}
	return nil
}
