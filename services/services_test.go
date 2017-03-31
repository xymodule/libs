package services

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestService(t *testing.T) {
	Init("/backends", []string{"172.16.42.1:2379"}, []string{"snowflake"})
	spew.Dump(_default_pool)
	if GetService("/backends/snowflake") == nil {
		t.Log("get service failed")
	} else {
		t.Log("get service succeed")
	}

	if GetServiceWithId("/backends/snowflake", "snowflake1") == nil {
		t.Log("get service with id failed")
	} else {
		t.Log("get service with id succeed")
	}
}
