package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestMemoryStoreIncr(t *testing.T) {
	key := "testKey"

	// testIncr() inline test function that will call Incr() with the given key N times,
	// sleeping between calls to ensure future eviction time is being processed correctly.

	testIncr := func(store RateLimiterStore, key string, hits int, sleep time.Duration) {
		for i := 0; i < hits; i++ {
			c, err := store.Incr(key)

			if err != nil {
				t.Errorf("Incr failed: %v", err)
			}

			expectedHits := uint64(i + 1)
			if c.hits != expectedHits {
				t.Errorf("Expected %d hits, got %d", expectedHits, c.hits)
			}

			// Check resetIn value (it should decrease over time but we can't assert exact values due to timing variances)
			if c.resetIn > uint64(hits) || c.resetIn == 0 {
				t.Errorf("Expected resetIn between 0 and 5, got %d", c.resetIn)
			}

			// Sleep for 900 milliseconds between calls
			time.Sleep(sleep)
		}
	}

	// Step 1: Create Store 5 seconds "window"
	// eviction time per each counter key
	store := NewMemoryStore(5 * time.Second)

	// Step 2: Test Incr() with given key
	testIncr(store, key, 5, 900*time.Millisecond)

	// Step 3: Wait 1 second, eviction should be triggered
	// TODO: assert that store map had key cleaned up
	time.Sleep(time.Second)

	// Step 4: Test Incr() just like in step 2, eviction
	// of key should had been triggered, meaning key Incr()
	// should behave just like the tests in Step 2.
	testIncr(store, key, 5, 900*time.Millisecond)
}

func TestRateLimiterMiddleware(t *testing.T) {
	// Step 1: Construct Middleware parameters

	limit := 5
	policy := func(r *http.Request) string {
		return "hits"
	}

	store := NewMemoryStore(5 * time.Second)
	middleware := RateLimiter(store, policy, uint64(limit))

	// Helper function to make a request and check expected values
	checkRequest := func(wantRemaining int, wantResetIn int, wantTooMany bool) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		middleware(handler).ServeHTTP(rr, req)

		// Assert middleware is returning correct
		// HTTP code based on wantTooMany param

		status := rr.Code
		if !wantTooMany && status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v, want %v", status, http.StatusOK)
		} else if wantTooMany && status != http.StatusTooManyRequests {
			t.Errorf("handler returned wrong status code: got %v, want %v", status, http.StatusTooManyRequests)
		}

		// Assert X-RateLimit-Remaining correct remaining hits
		remaining := rr.Header().Get("X-RateLimit-Remaining")
		if remaining != strconv.Itoa(wantRemaining) {
			t.Errorf("Expected remaining %d, got %s", wantRemaining, remaining)
		}

		// Assert X-RateLimit-Reset correct eviction time
		reset := rr.Header().Get("X-RateLimit-Reset")
		resetInt, _ := strconv.ParseInt(reset, 10, 64)

		if int(resetInt) != wantResetIn {
			t.Errorf("X-RateLimit-Reset should be %d, got: %s", wantResetIn, reset)
		}
	}

	// Step 1: Perform 5 requests
	for i := 0; i < limit; i++ {
		checkRequest(limit-1-i, limit-1-i, false)
		time.Sleep(time.Second)
	}

	// Step 2: Sleep for a second to wait
	// for counter key eviction to trigger
	time.Sleep(1 * time.Second)

	// Step 3: Perform another 5 requests,
	// counter must have been evicted now
	for i := 0; i < limit; i++ {
		checkRequest(limit-1-i, limit-1-i, false)
		time.Sleep(time.Second)
	}

	time.Sleep(1 * time.Second)

	// Step 4: consume allowed window hits
	// and issue additional requests to
	// assert too many requests response

	for i := 0; i < limit; i++ {
		checkRequest(limit-1-i, limit-1, false)
	}

	for i := 0; i < limit; i++ {
		checkRequest(0, limit-1-i, true)
		time.Sleep(time.Second)
	}
}
