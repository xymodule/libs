package services

import (
	"testing"
)

func TestService(t *testing.T) {
	if GetService("/backends/snowflake") == nil {
		t.Log("get service failed")
	} else {
		t.Log("get service succeed")
	}

	if GetServiceWithId("/backends/snowflake", "5946f18904ab:elated_blackwell:50003") == nil {
		t.Log("get service with id failed")
	} else {
		t.Log("get service with id succeed")
	}
}
