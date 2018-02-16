package store_test

import (
	"github.com/baratov/golang-playground/store"
	"runtime"
	"strings"
	"testing"
	"time"
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
	if actual := err.Error(); actual != expected { // hmm
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
	t.Parallel()

	s := store.New()

	for i := 0; i < 100000; i++ {
		go s.Set("key", i, time.Second)
		go s.Get("key")
		go s.Update("key", i, time.Second)
		go s.Delete("key")
		go s.Keys()
	}
}

func TestPersistence(t *testing.T) {
	s := store.New()
	s.Set("someKey", 123, time.Minute)
	time.Sleep(time.Second * 3)

	restored := store.Restore()
	val, err := restored.Get("someKey")
	if err != nil {
		t.Errorf("Error found %s", err.Error())
	}
	if val != 123 {
		t.Errorf("Expected value is 123, but found %s", val)
	}
}
