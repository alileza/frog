package storage

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	DefaultPermission = 0755
)

type Storage struct {
	basePath string
}

func New(folder string) *Storage {
	return &Storage{folder}
}

func (s *Storage) path(p, filename string) string {
	return fmt.Sprintf("%s/%s/%s", s.basePath, p, filename)
}

func (s *Storage) Store(name string, schema, body []byte, err error) error {
	os.Mkdir(s.basePath, DefaultPermission)
	os.Mkdir(s.basePath+"/"+name, DefaultPermission)

	ioutil.WriteFile(s.path(name, "schema.json"), schema, DefaultPermission)
	ioutil.WriteFile(s.path(name, "body.json"), body, DefaultPermission)
	if err != nil {
		ioutil.WriteFile(s.path(name, "error.json"), []byte(err.Error()), DefaultPermission)
	} else {
		os.Remove(s.path(name, "error.json"))
	}

	return nil
}

func (s *Storage) Handler() http.Handler {
	return http.FileServer(http.Dir(s.basePath))
}
