package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClosestPharmacy(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	expected := "Hello, World!"

	if body != expected {
		t.Errorf("expected body %q, got %q", expected, body)
	}
}
