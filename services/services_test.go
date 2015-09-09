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
}
