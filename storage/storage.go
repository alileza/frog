package storage

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	stdjson "encoding/json"

	"github.com/alileza/frog/util/json"
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

func hash(text string) (string, error) {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (s *Storage) Store(name string, schema, body []byte, err error) error {
	os.Mkdir(s.basePath, DefaultPermission)
	os.Mkdir(s.basePath+"/"+name, DefaultPermission)
	os.Mkdir(s.basePath+"/"+name+"/errors", DefaultPermission)

	ioutil.WriteFile(s.path(name, "schema.json"), schema, DefaultPermission)
	ioutil.WriteFile(s.path(name, "body.json"), body, DefaultPermission)
	if err != nil {
		var diffs []json.Diff
		if err := stdjson.Unmarshal([]byte(err.Error()), &diffs); err != nil {
			return err
		}
		for _, d := range diffs {
			out, _ := stdjson.Marshal(d)
			ioutil.WriteFile(s.path(name, "errors/"+d.Field+".json"), out, DefaultPermission)
		}
	}

	return nil
}

func (s *Storage) Handler() http.Handler {
	return http.FileServer(http.Dir(s.basePath))
}
