package ba

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnauthorized(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error()
	}

	rr := httptest.NewRecorder()

	SetUserPassword("user", "password")

	handler := HandlerFunc(ih)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Error()
	}
	if rr.Body.String() != "Unauthorized.\n" {
		t.Error()
	}
}

func TestAuthorized(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	req.SetBasicAuth("user", "password")
	if err != nil {
		t.Error()
	}

	rr := httptest.NewRecorder()

	SetUserPassword("user", "password")

	handler := HandlerFunc(ih)

	handler.ServeHTTP(rr, req)

	if rr.Code == http.StatusUnauthorized {
		t.Error()
	}
	if rr.Body.String() != "Hi!" {
		t.Error()
	}
}

func TestRobots(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Error()
	}
	rr := httptest.NewRecorder()

	DisallowRobots(rr, req)

	if rr.Code != http.StatusOK {
		t.Error()
	}
	if rr.Body.String() != "User-agent: *\nDisallow: /\n" {
		t.Error()
	}
}

func ih(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi!"))
}
