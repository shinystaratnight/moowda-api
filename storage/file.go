package storage

import (
	"github.com/labstack/echo"

	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// config is like {"path":"static"}
func NewLocalStorage(config string) (FileStorage, error) {
	var cf map[string]string

	err := json.Unmarshal([]byte(config), &cf)
	if err != nil {
		return nil, err
	}

	if _, ok := cf["path"]; !ok {
		return nil, fmt.Errorf("storage config has no path key")
	}
	if _, ok := cf["baseURI"]; !ok {
		return nil, fmt.Errorf("storage config has no baseURI key")
	}
	return &localStorage{path: cf["path"], baseURI: cf["baseURI"]}, nil
}

type localStorage struct {
	path    string
	baseURI string
}

func (s *localStorage) Store(c echo.Context, name string, sourceFile io.ReadCloser) (string, error) {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s%s", hex.EncodeToString(randBytes), filepath.Ext(name))
	f, err := os.OpenFile(fmt.Sprintf("%s/%s", s.path, filename), os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, sourceFile)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", s.baseURI, filename), nil
}

func init() {
	Register("local", NewLocalStorage)
}
