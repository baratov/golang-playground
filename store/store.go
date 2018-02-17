package store

import (
	"encoding/gob"
	"fmt"
	"os"
	"sync"
	"time"
)

const keyNotFoundFmt = "key '%s' not found"
const filename = "./store.gob"

type item struct {
	Value      interface{} // interface{} says nothing?
	Expiration time.Time
}

func (item *item) isExpired() bool {
	return time.Now().After(item.Expiration)
}

type Store struct {
	mu      sync.RWMutex    // https://github.com/golang/go/wiki/MutexOrChannel
	items   map[string]item // sync.Map could give synchronization out of the box and help to avoid cache contention
	updates chan bool
}

func New() *Store {
	return create(make(map[string]item))
}

func Restore() *Store {
	return create(load())
}

func create(items map[string]item) *Store {
	s := &Store{
		items: items,
		updates: make(chan bool),
	}
	go s.runExpiration()
	go s.runFlushing()
	return s
}

func (s *Store) Get(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock() // gives performance overhead

	return s.get(key)
}

func (s *Store) get(key string) (interface{}, error) {
	i, ok := s.items[key]
	if ok && !i.isExpired() {
		return i.Value, nil
	}
	return nil, fmt.Errorf(keyNotFoundFmt, key)
}

func (s *Store) Set(key string, value interface{}, ttl time.Duration) {
	i := item{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}

	s.mu.Lock()
	s.set(key, i)
	s.mu.Unlock()

	s.updates <- true
}

func (s *Store) set(key string, i item) {
	s.items[key] = i
}

func (s *Store) Update(key string, value interface{}, ttl time.Duration) error {
	i := item{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}

	s.mu.Lock()
	err := s.update(key, i)
	s.mu.Unlock()

	s.updates <- true

	return err
}

func (s *Store) update(key string, i item) error {
	_, err := s.get(key)
	if err == nil {
		s.items[key] = i
	}
	return err
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	s.delete(key)
	s.mu.Unlock()

	s.updates <- true
}

func (s *Store) delete(key string) {
	delete(s.items, key)
}

func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.keys()
}

func (s *Store) keys() []string {
	keys := make([]string, 0, len(s.items))
	for key, item := range s.items {
		if !item.isExpired() {
			keys = append(keys, key)
		}
	}
	return keys
}

func (s *Store) runExpiration() { // will not be garbage collected
	c := time.Tick(time.Second)
	for {
		select {
		case <-c:
			s.expire()
		}
	}
}

func (s *Store) expire() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, item := range s.items {
		if item.isExpired() {
			delete(s.items, key)
		}
	}
}

// calls store.flush every 2 seconds or after 5 updates
func (s *Store) runFlushing() {
	flushingInterval := time.Hour * 2
	timer := time.NewTimer(flushingInterval)
	flushingCount := 5
	counter := 0

	for {
		select {
		case <-timer.C:
			s.flush()
			timer.Reset(flushingInterval)
			counter = 0
		case <-s.updates:
			counter++
			if counter >= flushingCount {
				s.flush()
				timer.Reset(flushingInterval)
				counter = 0
			}
		}
	}
}

func (s *Store) flush() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(s.items)
	if err != nil {
		panic(err)
	}
}

func load() map[string]item {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	m := make(map[string]item)

	err = decoder.Decode(&m)
	if err != nil {
		panic(err)
	}
	return m
}
