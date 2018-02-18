package store_test

import (
	"github.com/baratov/golang-playground/store"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
	"fmt"
)

func TestGet(t *testing.T) {
	s := store.New()
	s.Set("someKey", 123, time.Second)

	val, err := s.Get("someKey")
	if err != nil {
		t.Errorf("Error found %s", err.Error())
	}
	if val != 123 {
		t.Errorf("Expected value is 123, but found %s", val)
	}
}

func TestGet_NonExistingKey(t *testing.T) {
	s := store.New()

	val, err := s.Get("someKey")
	expected := "key 'someKey' not found"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected error is %s, but found %s", expected, actual)
	}
	if val != nil {
		t.Errorf("Expected value is nil, but found %s", val)
	}
}

func TestGet_ExpiredKey(t *testing.T) {
	s := store.New()
	s.Set("someKey", 123, time.Second)
	time.Sleep(time.Second)

	val, err := s.Get("someKey")
	expected := "key 'someKey' not found"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected error is %s, but found %s", expected, actual)
	}
	if val != nil {
		t.Errorf("Expected value is nil, but found %s", val)
	}
}

func TestUpdate(t *testing.T) {
	s := store.New()
	s.Set("someKey", 123, time.Second)

	err := s.Update("someKey", 234, time.Second)
	if err != nil {
		t.Errorf("Error found %s", err.Error())
	}

	val, err := s.Get("someKey")
	if err != nil {
		t.Errorf("Error found %s", err.Error())
	}
	if val != 234 {
		t.Errorf("Expected value is 234, but found %s", val)
	}
}

func TestUpdate_NonExistingKey(t *testing.T) {
	s := store.New()

	err := s.Update("someKey", 234, time.Second)
	expected := "key 'someKey' not found"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected error is %s, but found %s", expected, actual)
	}
}

func TestUpdate_ExpiredKey(t *testing.T) {
	s := store.New()
	s.Set("someKey", 123, time.Second)
	time.Sleep(time.Second)

	err := s.Update("someKey", 234, time.Second)
	expected := "key 'someKey' not found"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected error is %s, but found %s", expected, actual)
	}
}

func TestDelete(t *testing.T) {
	s := store.New()
	s.Set("someKey", 123, time.Second)

	s.Delete("someKey")

	val, err := s.Get("someKey")
	expected := "key 'someKey' not found"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected error is %s, but found %s", expected, actual)
	}
	if val != nil {
		t.Errorf("Expected value is nil, but found %s", val)
	}
}

func TestKeys(t *testing.T) {
	s := store.New()
	s.Set("someKey", 123, time.Second)
	s.Set("otherKey", 234, time.Second)
	s.Set("someKey", 345, time.Second)

	keys := s.Keys()
	if len(keys) != 2 || keys[0] != "someKey" || keys[1] != "otherKey" { // the order is not guaranteed I guess
		expected := strings.Join([]string{"someKey", "otherKey"}, ",")
		actual := strings.Join(keys, ",")
		t.Errorf("Expected array is [%s], but found [%s]", expected, actual)
	}
}

func TestKeys_EmptyStore(t *testing.T) {
	s := store.New()

	keys := s.Keys()
	if len(keys) != 0 {
		actual := strings.Join(keys, ",")
		t.Errorf("Expected array is empty, but found [%s]", actual)
	}
}

func TestKeys_ExpiredKey(t *testing.T) {
	s := store.New()
	s.Set("someKey", 123, time.Second)
	s.Set("otherKey", 123, 2*time.Second)
	time.Sleep(time.Second)

	keys := s.Keys()
	expected := "otherKey"
	if len(keys) != 1 || keys[0] != expected {
		actual := strings.Join(keys, ",")
		t.Errorf("Expected array is [%s], but found [%s]", expected, actual)
	}
}

func TestConcurrentAccess(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU()) //test won't work on single-core CPU

	s := store.New()

	for i := 0; i < 100000; i++ {
		go s.Set("key", i, time.Second)
		go s.Get("key")
		go s.Update("key", i, time.Second)
		go s.Delete("key")
		go s.Keys()
	}

	s.Stop()
}

func TestFlushingByTimer(t *testing.T) {
	filename := fmt.Sprintf("%s%d%s", "./store_", time.Now().Unix(), ".gob")

	s := store.New(
		store.WithFilename(filename),
	)
	s.Set("someKey", 123, time.Minute)
	time.Sleep(time.Second * 3) // magic number to not overlap restore with flushing

	r := store.New(
		store.WithFilename(filename),
		store.WithRestoreFromFile(filename),
	)
	val, err := r.Get("someKey")
	if err != nil {
		t.Errorf("Error found %s", err.Error())
	}
	if val != 123 {
		t.Errorf("Expected value is 123, but found %s", val)
	}

	//teardown
	if err := os.Remove(filename); err != nil {
		t.Error(err.Error())
	}
}

func TestFlushingByUpdatesCount(t *testing.T) {
	filename := fmt.Sprintf("%s%d%s", "./store_", time.Now().Unix(), ".gob")

	s := store.New(
		store.WithFilename(filename),
	)
	s.Set("key1", 1, time.Hour)
	s.Set("key2", 2, time.Hour)
	s.Set("key3", 3, time.Hour)
	s.Update("key1", 4, time.Hour)
	s.Delete("key2")
	time.Sleep(time.Second * 3) // magic number not to overlap with flushing

	r := store.New(
		store.WithFilename(filename),
		store.WithRestoreFromFile(filename),
	)
	val, err := r.Get("key1")
	if err != nil {
		t.Errorf("Error found %s", err.Error())
	}
	if val != 4 {
		t.Errorf("Expected value is 4, but found %s", val)
	}

	val, err = r.Get("key3")
	if err != nil {
		t.Errorf("Error found %s", err.Error())
	}
	if val != 3 {
		t.Errorf("Expected value is 3, but found %s", val)
	}

	//teardown
	if err := os.Remove(filename); err != nil {
		t.Error(err.Error())
	}
}

func TestFlushingByStop(t *testing.T) {
	filename := fmt.Sprintf("%s%d%s", "./store_", time.Now().Unix(), ".gob")

	s := store.New(
		store.WithFilename(filename),
	)
	s.Set("someKey", 123, time.Minute)
	s.Stop()

	r := store.New(
		store.WithFilename(filename),
		store.WithRestoreFromFile(filename),
	)
	val, err := r.Get("someKey")
	if err != nil {
		t.Errorf("Error found %s", err.Error())
	}
	if val != 123 {
		t.Errorf("Expected value is 123, but found %s", val)
	}

	//teardown
	if err := os.Remove(filename); err != nil {
		t.Error(err.Error())
	}
}
