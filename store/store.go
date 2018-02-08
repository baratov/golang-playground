package store

import (
	"fmt"
	"sync"
)

const keyNotFoundFmt = "key '%s' not found"

type Store struct {
	mutex sync.RWMutex           // https://github.com/golang/go/wiki/MutexOrChannel
	items map[string]interface{} // sync.Map could give synchronization out of the box
} // interface{} says nothing

func New() *Store {
	return &Store{
		items: make(map[string]interface{}),
	}
}

func (store *Store) Get(key string) (interface{}, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock() // gives performance overhead

	return store.get(key)
}

func (store *Store) get(key string) (interface{}, error) {
	value, ok := store.items[key]
	if ok {
		return value, nil
	}
	return nil, fmt.Errorf(keyNotFoundFmt, key)
}

func (store *Store) Set(key string, value interface{}) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.set(key, value)
}

func (store *Store) set(key string, value interface{}) {
	store.items[key] = value
}

func (store *Store) Update(key string, value interface{}) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	return store.update(key, value)
}

func (store *Store) update(key string, value interface{}) error {
	_, err := store.get(key)
	if err == nil {
		store.items[key] = value
	}
	return err
}

func (store *Store) Delete(key string) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.delete(key)
}

func (store *Store) delete(key string) {
	delete(store.items, key)
}

func (store *Store) Keys() []string {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	return store.keys()
}

func (store *Store) keys() []string {
	keys := make([]string, len(store.items))

	i := 0
	for key := range store.items {
		keys[i] = key
		i++ // append is probably more readable
	}

	return keys
}
