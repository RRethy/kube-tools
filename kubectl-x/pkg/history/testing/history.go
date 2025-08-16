package testing

import (
	"github.com/RRethy/kubectl-x/pkg/history"
)

// TODO: we don't need to do this
var _ history.Interface = &FakeHistory{}

type FakeHistory struct {
	Data    map[string][]string
	Written bool
}

func (fake *FakeHistory) Get(key string, index int) (string, error) {
	return fake.Data[key][index], nil
}

func (fake *FakeHistory) Add(key, value string) {
	fake.Written = false
	fake.Data[key] = append([]string{value}, fake.Data[key]...)
}

func (fake *FakeHistory) Write() error {
	return nil
}
