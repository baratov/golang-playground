package store_test

import (
	"github.com/baratov/golang-playground/store"
	"runtime"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	store := store.New()
	store.Set("someKey", 123)

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

func TestUpdate(t *testing.T) {
	store := store.New()
	store.Set("someKey", 123)

	err := store.Update("someKey", 234)
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

	err := store.Update("someKey", 234)
	expected := "key 'someKey' not found"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected error is %s, but found %s", expected, actual)
	}
}

func TestDelete(t *testing.T) {
	store := store.New()
	store.Set("someKey", 123)

	err := store.Delete("someKey")
	if err != nil {
		t.Errorf("Error found %s", err.Error())
	}

	val, err := store.Get("someKey")
	expected := "key 'someKey' not found"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected error is %s, but found %s", expected, actual)
	}
	if val != nil {
		t.Errorf("Expected value is nil, but found %s", val)
	}
}

func TestDelete_NonExistingKey(t *testing.T) {
	store := store.New()

	err := store.Delete("someKey")
	expected := "key 'someKey' not found"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected error is %s, but found %s", expected, actual)
	}
}

func TestKeys(t *testing.T) {
	store := store.New()
	store.Set("someKey", 123)
	store.Set("otherKey", 234)
	store.Set("someKey", 345)

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
	runtime.GOMAXPROCS(runtime.NumCPU())
	t.Parallel()

	store := store.New()

	for i := 0; i < 100000; i++ {
		go store.Set("key", i)
		go store.Get("key")
		go store.Update("key", i)
		go store.Delete("key")
		go store.Keys()
	}
}
