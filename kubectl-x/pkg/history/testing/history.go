package testing

import (
	"errors"

	"github.com/RRethy/kubectl-x/pkg/history"
)

// TODO: we don't need to do this
var _ history.Interface = &FakeHistory{}

type FakeHistory struct {
	Data    map[string][]string
	Written bool
}

func (fake *FakeHistory) Get(key string, index int) (string, error) {
	if values, ok := fake.Data[key]; ok && len(values) > index {
		return values[index], nil
	}
	return "", errors.New("not found")
}

func (fake *FakeHistory) Add(key, value string) {
	fake.Written = false
	fake.Data[key] = append([]string{value}, fake.Data[key]...)
}

func (fake *FakeHistory) Write() error {
	return nil
}
