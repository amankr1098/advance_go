package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newServer creates a test HTTP server that responds with the given status code and body.
func newServer(statusCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
}

// newSlowServer creates a test HTTP server that delays responding by the given duration.
func newSlowServer(delay time.Duration, statusCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(delay):
		}
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
}

// TestFetchAll_AllSucceed verifies that all results are returned when all URLs succeed.
func TestFetchAll_AllSucceed(t *testing.T) {
	s1 := newServer(200, "body1")
	s2 := newServer(200, "body2")
	s3 := newServer(200, "body3")
	defer s1.Close()
	defer s2.Close()
	defer s3.Close()

	urls := []string{s1.URL, s2.URL, s3.URL}
	ctx := context.Background()

	results, err := fetchAll(ctx, urls)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got: %d", len(results))
	}
}

// TestFetchAll_EmptyURLs verifies that an empty URL list returns empty results and no error.
func TestFetchAll_EmptyURLs(t *testing.T) {
	ctx := context.Background()

	results, err := fetchAll(ctx, []string{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got: %d", len(results))
	}
}

// TestFetchAll_One404_ReturnsError verifies that a 404 response causes fetchAll to return an error.
func TestFetchAll_One404_ReturnsError(t *testing.T) {
	s1 := newServer(200, "ok")
	s2 := newServer(404, "not found")
	defer s1.Close()
	defer s2.Close()

	urls := []string{s1.URL, s2.URL}
	ctx := context.Background()

	_, err := fetchAll(ctx, urls)
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}

// TestFetchAll_All404_ReturnsError verifies that all-404 URLs return an error.
func TestFetchAll_All404_ReturnsError(t *testing.T) {
	s1 := newServer(404, "not found")
	s2 := newServer(404, "not found")
	defer s1.Close()
	defer s2.Close()

	urls := []string{s1.URL, s2.URL}
	ctx := context.Background()

	_, err := fetchAll(ctx, urls)
	if err == nil {
		t.Fatal("expected error for all-404 responses, got nil")
	}
}

// TestFetchAll_ContextCancelledBeforeStart verifies that a pre-cancelled context returns an error immediately.
func TestFetchAll_ContextCancelledBeforeStart(t *testing.T) {
	s1 := newServer(200, "ok")
	defer s1.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before calling fetchAll

	_, err := fetchAll(ctx, []string{s1.URL})
	if err == nil {
		t.Fatal("expected error for pre-cancelled context, got nil")
	}
}

// TestFetchAll_ContextTimeout verifies that a very short timeout causes fetchAll to return an error.
func TestFetchAll_ContextTimeout(t *testing.T) {
	slow := newSlowServer(2*time.Second, 200, "ok")
	defer slow.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := fetchAll(ctx, []string{slow.URL})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// TestFetchAll_SlowURLCancelledOnError verifies that a slow URL is cancelled when another URL fails.
func TestFetchAll_SlowURLCancelledOnError(t *testing.T) {
	slow := newSlowServer(5*time.Second, 200, "ok")
	bad := newServer(500, "error")
	defer slow.Close()
	defer bad.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	_, err := fetchAll(ctx, []string{slow.URL, bad.URL})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error due to 500 response, got nil")
	}
	// Should return well before the slow server's 5s delay
	if elapsed > 3*time.Second {
		t.Fatalf("fetchAll took too long (%v), slow URL was not cancelled", elapsed)
	}
}

// TestFetchAll_500_ReturnsError verifies that a 500 response is treated as an error.
func TestFetchAll_500_ReturnsError(t *testing.T) {
	s := newServer(500, "internal server error")
	defer s.Close()

	ctx := context.Background()
	_, err := fetchAll(ctx, []string{s.URL})
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

// TestFetchAll_201_Succeeds verifies that 201 Created is treated as a success.
func TestFetchAll_201_Succeeds(t *testing.T) {
	s := newServer(201, "created")
	defer s.Close()

	ctx := context.Background()
	results, err := fetchAll(ctx, []string{s.URL})
	if err != nil {
		t.Fatalf("expected no error for 201 response, got: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got: %d", len(results))
	}
}

// TestFetchAll_ResultBodyContent verifies the actual body content is returned.
func TestFetchAll_ResultBodyContent(t *testing.T) {
	expected := "hello world"
	s := newServer(200, expected)
	defer s.Close()

	ctx := context.Background()
	results, err := fetchAll(ctx, []string{s.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0] != expected {
		t.Fatalf("expected body %q, got %q", expected, results[0])
	}
}
