package store

import (
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"
)

const (
	errKeyNotFoundFmt = "key '%v' not found"
	errWrongTypeFmt   = "wrong type for key '%v'"

	defFilename      = "./store.gob"
	defExpInterval   = time.Second
	defFlushInterval = time.Second * 2
	defFlushCount    = 5
)

type item struct {
	Value      interface{} // interface{} says nothing
	Expiration time.Time
}

func (item *item) isExpired() bool {
	return time.Now().After(item.Expiration)
}

type setting func(*Store)

type Store struct {
	mu                 sync.RWMutex    // https://github.com/golang/go/wiki/MutexOrChannel
	items              map[string]item // sync.Map could give synchronization out of the box and help to avoid cache contention
	updates            chan bool
	stop               chan bool
	expirationInterval time.Duration
	flushingInterval   time.Duration
	flushingCount      int
	wg                 sync.WaitGroup
	filename           string
}

func New(settings ...setting) *Store {
	s := &Store{
		items:              make(map[string]item),
		updates:            make(chan bool, 5),
		stop:               make(chan bool),
		filename:           defFilename,
		expirationInterval: defExpInterval,
		flushingInterval:   defFlushInterval,
		flushingCount:      defFlushCount,
	}

	for _, setting := range settings {
		setting(s)
	}

	s.wg.Add(1)
	go s.runFlushing()
	go s.runExpiration()
	return s
}

func WithCustomFilename(filename string) setting {
	return func(s *Store) {
		s.filename = filename
	}
}

func WithRestoreFromFile(filename string) setting {
	return func(s *Store) {
		s.items = load(filename)
	}
}

func (s *Store) Stop() {
	s.stop <- true
	s.stop <- true // looks strange, change to close(s.stop)
	s.wg.Wait()
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
	return nil, fmt.Errorf(errKeyNotFoundFmt, key)
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

func (s *Store) GetMapEntry(key, innerKey string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getMapEntry(key, innerKey)
}

// option 1 - simple and stupid
/*func (s *Store) getMapEntry(key, innerKey string) (interface{}, error) {
	val, err := s.get(key)
	if err != nil {
		return nil, err
	}

	var result interface{}
	var ok bool

	switch val.(type) {
	case map[string]bool:
		result, ok = val.(map[string]bool)[innerKey]
	case map[string]string:
		result, ok = val.(map[string]string)[innerKey]
	case map[string]int32:
		result, ok = val.(map[string]int32)[innerKey]
	case map[string]int64:
		result, ok = val.(map[string]int64)[innerKey]
	case map[string]float32:
		result, ok = val.(map[string]float32)[innerKey]
	case map[string]float64:
		result, ok = val.(map[string]float64)[innerKey]
	default:
		return nil, fmt.Errorf(errWrongTypeFmt, key)
	}

	if !ok {
		return nil, fmt.Errorf(errKeyNotFoundFmt, innerKey)
	}

	return result, nil
}*/

// option 2 - O(N)
/*func (s *Store) getMapEntry(key, innerKey string) (interface{}, error) {
	val, err := s.get(key)
	if err != nil {
		return nil, err
	}

	m := reflect.ValueOf(val)
	if m.Kind() == reflect.Map {
		for _, key := range m.MapKeys() {
			if key.Interface() == innerKey {
				return m.MapIndex(key).Interface(), nil
			}
		}
		return nil, fmt.Errorf(errKeyNotFoundFmt, innerKey)
	} else {
		return nil, fmt.Errorf(errWrongTypeFmt, key)
	}
}*/

// option 3
func (s *Store) getMapEntry(key, innerKey string) (result interface{}, err error) {
	val, err := s.get(key)
	if err != nil {
		return nil, err
	}

	// deferred recover works only with naked return?
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(errKeyNotFoundFmt, innerKey)
		}
	}()

	m := reflect.ValueOf(val)
	if m.Kind() == reflect.Map {
		result = m.MapIndex(reflect.ValueOf(innerKey)).Interface()
	} else {
		err = fmt.Errorf(errWrongTypeFmt, key)
	}
	return
}

func (s *Store) runExpiration() {
	c := time.Tick(s.expirationInterval)
	for {
		select {
		case <-c:
			s.expire()
		case <-s.stop:
			return
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

// calls store.flush by timer or after number of updates
func (s *Store) runFlushing() {
	defer s.wg.Done()

	timer := time.NewTimer(s.flushingInterval) // do I need defer timer.Stop() here?
	counter := 0

	flushAndReset := func() {
		s.flush()
		timer.Reset(s.flushingInterval)
		counter = 0
	}

	for {
		select {
		case <-timer.C:
			flushAndReset()
		case <-s.updates:
			counter++
			if counter >= s.flushingCount {
				flushAndReset()
			}
		case <-s.stop:
			flushAndReset()
			return
		}
	}
}

func (s *Store) flush() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := os.Create(s.filename)
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

func load(filename string) map[string]item {
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
