package evaluator

import (
	stdjson "encoding/json"
	"errors"
	"sync"

	"github.com/alileza/frog/util/json"
)

type Report struct {
	Name   string
	Body   []byte
	Schema []byte
	Error  error
}

type Evaluator struct {
	messages sync.Map
}

func New() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Eval(name string, body []byte) *Report {
	b, ok := e.messages.Load(name)
	e.messages.Store(name, body)
	if !ok {
		return nil
	}
	prevBody := b.([]byte)

	bodySchema, err := json.ToType(body)
	if err != nil {
		return &Report{
			Name:   name,
			Body:   body,
			Schema: bodySchema,
			Error:  err,
		}
	}

	prevBodySchema, err := json.ToType(prevBody)
	if err != nil {
		return &Report{
			Name:   name,
			Body:   body,
			Schema: bodySchema,
			Error:  err,
		}
	}
	prevBodySchemaMap, err := json.ToMap(prevBodySchema)
	if err != nil {
		panic(err)
	}
	bodySchemaMap, err := json.ToMap(bodySchema)
	if err != nil {
		panic(err)
	}

	diffs, err := json.Cmp(prevBodySchemaMap, bodySchemaMap, "")
	if err != nil {
		panic(err)
	}

	diff, _ := stdjson.Marshal(diffs)
	return &Report{
		Name:   name,
		Body:   body,
		Schema: bodySchema,
		Error:  errors.New(string(diff)),
	}
}
