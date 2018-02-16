package store

// https://medium.com/@matryer/5-simple-tips-and-tricks-for-writing-unit-tests-in-golang-619653f90742

import (
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	store := New()
	store.Set("someKey", 123, time.Second)

	item := store.items["someKey"]
	if item.Value != 123 {
		t.Errorf("Expected value is %v, but found %v", 123, item.Value)
	}
	if l := len(store.items); l != 1 {
		t.Errorf("Expected len of internal map is 1, but found %v", l)
	}
}

func TestStoreDeletesExpiredItems(t *testing.T) {
	store := New()
	store.Set("someKey", 123, time.Second)
	time.Sleep(time.Second * 2)

	if l := len(store.items); l != 0 {
		t.Errorf("Expected len of internal map is 0, but found %v", l)
	}
}
