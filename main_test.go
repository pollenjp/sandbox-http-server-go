package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSample(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	rootHandler(w, r)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatal("unexpected status code")
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal("unexpected error")
	}
	const expected = `{"content_type":"","msg":"hello","a":"","b":""}` + "\n"
	if s := string(b); s != expected {
		t.Fatalf("unexpected response:\nexpected: %s\nactual: %s", s, expected)
	}
}
