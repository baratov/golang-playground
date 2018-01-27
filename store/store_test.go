package store

import (
	"testing"
)

func TestGetSet(t *testing.T) {
	store := NewStore()
	store.Set("someKey", 123)

	value:= store.Get("someKey")

	if value != 123 {
		t.Errorf("Expected value is 123, but found %d", value)
	}
}

func TestUpdate(t *testing.T) {
	store := NewStore()
	store.Set("someKey", 123)

	store.Update("someKey", 234)

	value := store.Get("someKey")

	if value != 234 {
		t.Errorf("Expected value is 234, but found %d", value)
	}
}

func TestDelete(t *testing.T) {
	store := NewStore()
	store.Set("someKey", 123)

	store.Delete("someKey")

	value := store.Get("someKey")
	if value != nil {
		t.Errorf("Expected value is nil, but found %d", value)
	}
}
