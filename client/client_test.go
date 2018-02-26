package client_test

// integration tests for main use-cases
// server running on localhost:8080 is required

import (
	"github.com/baratov/golang-playground/client"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	c := client.New("http://localhost:8080/",
		client.BasicAuthorization("username", "password"),
		client.LogLatency())

	origVal := "some_string_value"
	err := c.Set("testKey", origVal, time.Second)
	if err != nil {
		t.Errorf("Error found: %v", err.Error())
	}

	val, err := c.Get("testKey")
	if err != nil {
		t.Errorf("Error found: %v", err.Error())
	}
	if val != origVal {
		t.Errorf("Excpected value is %v, but found %v", origVal, val)
	}
}

func TestUpdate(t *testing.T) {
	c := client.New("http://localhost:8080/",
		client.BasicAuthorization("username", "password"),
		client.LogLatency())

	err := c.Set("testKey", "some_string_value", time.Second)
	if err != nil {
		t.Errorf("Error found: %v", err.Error())
	}

	err = c.Update("testKey", "some_updated_value", time.Second)
	if err != nil {
		t.Errorf("Error found: %v", err.Error())
	}

	val, err := c.Get("testKey")
	if err != nil {
		t.Errorf("Error found: %v", err.Error())
	}
	if val != "some_updated_value" {
		t.Errorf("Excpected value is %v, but found %v", "some_updated_value", val)
	}
}

func TestDelete(t *testing.T) {
	c := client.New("http://localhost:8080/",
		client.BasicAuthorization("username", "password"),
		client.LogLatency())

	err := c.Set("testKey", "some_string_value", time.Second)
	if err != nil {
		t.Errorf("Error found: %v", err.Error())
	}
	err = c.Delete("testKey")
	if err != nil {
		t.Errorf("Error found: %v", err.Error())
	}

	val, err := c.Get("testKey")
	expected := "key 'testKey' not found"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected error is %v, but found %v", expected, actual)
	}
	if val != nil {
		t.Errorf("Expected value is nil, but found %v", val)
	}
}
