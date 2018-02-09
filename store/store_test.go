package store_test

import (
	"github.com/baratov/golang-playground/store"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	store := store.New()
	store.Set("someKey", 123, time.Second)

	val, err := store.Get("someKey")
	if err != nil {
		t.Errorf("Error found %s", err.Error())
	}
	if val != 123 {
		t.Errorf("Expected value is 123, but found %s", val)
	}
}

func TestGet_NonExistingKey(t *testing.T) {
	store := store.New()

	val, err := store.Get("someKey")
	expected := "key 'someKey' not found"
	if actual := err.Error(); actual != expected { // hmm
		t.Errorf("Expected error is %s, but found %s", expected, actual)
	}
	if val != nil {
		t.Errorf("Expected value is nil, but found %s", val)
	}
}

func TestGet_ExpiredKey(t *testing.T) {
	store := store.New()
	store.Set("someKey", 123, time.Second)
	time.Sleep(time.Second)

	val, err := store.Get("someKey")
	expected := "key 'someKey' not found"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected error is %s, but found %s", expected, actual)
	}
	if val != nil {
		t.Errorf("Expected value is nil, but found %s", val)
	}
}

func TestUpdate(t *testing.T) {
	store := store.New()
	store.Set("someKey", 123, time.Second)

	err := store.Update("someKey", 234, time.Second)
	if err != nil {
		t.Errorf("Error found %s", err.Error())
	}

	val, err := store.Get("someKey")
	if err != nil {
		t.Errorf("Error found %s", err.Error())
	}
	if val != 234 {
		t.Errorf("Expected value is 234, but found %s", val)
	}
}

func TestUpdate_NonExistingKey(t *testing.T) {
	store := store.New()

	err := store.Update("someKey", 234, time.Second)
	expected := "key 'someKey' not found"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected error is %s, but found %s", expected, actual)
	}
}

func TestDelete(t *testing.T) {
	store := store.New()
	store.Set("someKey", 123, time.Second)

	store.Delete("someKey")

	val, err := store.Get("someKey")
	expected := "key 'someKey' not found"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected error is %s, but found %s", expected, actual)
	}
	if val != nil {
		t.Errorf("Expected value is nil, but found %s", val)
	}
}

func TestKeys(t *testing.T) {
	store := store.New()
	store.Set("someKey", 123, time.Second)
	store.Set("otherKey", 234, time.Second)
	store.Set("someKey", 345, time.Second)

	keys := store.Keys()
	if len(keys) != 2 || keys[0] != "someKey" || keys[1] != "otherKey" { // the order is not guaranteed I guess
		expected := strings.Join([]string{"someKey", "otherKey"}, ",")
		actual := strings.Join(keys, ",")
		t.Errorf("Expected array is [%s], but found [%s]", expected, actual)
	}
}

func TestKey_EmptyStore(t *testing.T) {
	store := store.New()

	keys := store.Keys()
	if len(keys) != 0 {
		actual := strings.Join(keys, ",")
		t.Errorf("Expected array is empty, but found [%s]", actual)
	}
}

func TestConcurrentAccess(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU()) //test won't work on single-core CPU
	t.Parallel()

	store := store.New()

	for i := 0; i < 100000; i++ {
		go store.Set("key", i, time.Second)
		go store.Get("key")
		go store.Update("key", i, time.Second)
		go store.Delete("key")
		go store.Keys()
	}
}
