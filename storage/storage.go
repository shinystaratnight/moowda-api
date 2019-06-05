package storage

import (
	"github.com/labstack/echo"

	"io"
)

type FileStorage interface {
	Store(echo.Context, string, io.ReadCloser) (string, error)
}

// Instance is a function create a new FileStorage Instance
type Instance func(string) (FileStorage, error)

var Adapters = make(map[string]Instance)

// Register makes a file storage adapter available by the adapter name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, adapter Instance) {
	if adapter == nil {
		panic("storage: Register adapter is nil")
	}
	if _, ok := Adapters[name]; ok {
		panic("storage: Register called twice for adapter " + name)
	}
	Adapters[name] = adapter
}
