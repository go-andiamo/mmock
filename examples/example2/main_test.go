package main

import (
	"github.com/go-andiamo/mmock"
	"github.com/go-andiamo/mmock/examples/example2/stuff"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGenerateFile(t *testing.T) {
	fo, err := os.Create("stuff/mock_thing.go")
	assert.NoError(t, err)
	err = mmock.MockGenerateFile[stuff.Thingy]("", fo)
	assert.NoError(t, err)
}

func TestGenerate(t *testing.T) {
	data, err := mmock.MockGenerate[stuff.Thingy]("")
	assert.NoError(t, err)
	println(string(data))
}
