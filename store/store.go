package store

import (
	"fmt"
	"sync"
	"time"
)

const keyNotFoundFmt = "key '%s' not found"

type item struct {
	value      interface{} // interface{} says nothing
	expiration time.Time
}

func (item *item) isExpired() bool {
	return time.Now().After(item.expiration)
}

type Store struct {
	mutex sync.RWMutex    // https://github.com/golang/go/wiki/MutexOrChannel
	items map[string]item // sync.Map could give synchronization out of the box
}

func New() *Store {
	return &Store{
		items: make(map[string]item),
	}
}

func (store *Store) Get(key string) (interface{}, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock() // gives performance overhead

	return store.get(key)
}

func (store *Store) get(key string) (interface{}, error) {
	i, ok := store.items[key]
	if ok && !i.isExpired() {
		return i.value, nil
	}
	return nil, fmt.Errorf(keyNotFoundFmt, key)
}

func (store *Store) Set(key string, value interface{}, ttl time.Duration) {
	i := item{
		value:      value,
		expiration: time.Now().Add(ttl),
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.set(key, i)
}

func (store *Store) set(key string, i item) {
	store.items[key] = i
}

func (store *Store) Update(key string, value interface{}, ttl time.Duration) error {
	i := item{
		value:      value,
		expiration: time.Now().Add(ttl),
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	return store.update(key, i)
}

func (store *Store) update(key string, i item) error {
	_, err := store.get(key)
	if err == nil {
		store.items[key] = i
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
