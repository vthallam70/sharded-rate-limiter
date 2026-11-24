package main

import (
	limiter "github.com/vthallam70/sharded-rate-limiter"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUnavailableOwnerFailsClosed(t *testing.T) {
	nodes := []string{"http://127.0.0.1:1", "http://self"}
	key := ""
	for _, k := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"} {
		if owner(k, nodes) == nodes[0] {
			key = k
			break
		}
	}
	if key == "" {
		t.Fatal("no key mapped to failed node")
	}
	s := &server{self: nodes[1], nodes: nodes, limiter: limiter.New(1, 1), client: &http.Client{Timeout: time.Second}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/allow", strings.NewReader(`{"key":"`+key+`"}`))
	s.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
