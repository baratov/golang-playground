package store

import (
	"fmt"
)

const keyNotFoundFmt = "key '%s' not found"

type Store struct {
	items map[string]interface{}
}

func New() Store {
	return Store{
		items: make(map[string]interface{}),
	}
}

func (store Store) Get(key string) (interface{}, error) {
	value, ok := store.items[key]
	if ok {
		return value, nil
	}
	return nil, fmt.Errorf(keyNotFoundFmt, key)
}

func (store Store) Set(key string, value interface{}) {
	store.items[key] = value
}

func (store Store) Update(key string, value interface{}) error {
	_, err := store.Get(key)
	if err == nil {
		store.items[key] = value
	}
	return err
}

func (store Store) Delete(key string) error {
	_, err := store.Get(key)
	if err == nil {
		delete(store.items, key)
	}
	return err
}

func (store Store) Keys() []string {
	keys := make([]string, len(store.items))

	i := 0
	for key := range store.items {
		keys[i] = key
		i++
	}

	return keys
}
