package ratelimiter

import (
	"context"
	"fmt"
	"github.com/earcamone/gwy-playground/api/middleware/errorscheme"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	// third-party imports
	"github.com/go-redis/redis/v8"
)

// RateLimiterPolicy is just a callback that should generate the key to
// track hits allowed over the specified "window" of time by RateLimiter
// middleware.
//
// For example: for a global rate limiter that tracks total requests to
// your API over a window, just return a fixed key. For a policy per API
// end-point, return the request path in the http.Request and for a rate
// limiting policy by IP, return the client request IP.
//
// Kindly check documentation for many already implemented policies.
type RateLimiterPolicy func(r *http.Request) string

// RateLimiterStore implements a "counters per key" store that will
// perform an increment and return the counter upon each Incr() call.
//
// The returned Counter struct holds the counter value for the key
// in the value attribute, and its remaining time in seconds until
// it is evicted from the store.
type RateLimiterStore interface {
	Incr(key string) (*Counter, error)
}

// Counter holds the counter current value and the remaining
// time until it is evicted from the store it was retrieved.
type Counter struct {
	// hits holds the number of client issued Incr()
	hits uint64

	// resetIn holds the number of seconds till window reset
	resetIn uint64
}

// RateLimiter is the Rate Limiter Middleware constructor, "limit" stands
// for the number of allowed hits per the specified "window" time period.
// "policy" function generates the key used to track each "client" hit,
// for more information regarding "policy" functions, check RateLimiterPolicy
func RateLimiter(limiter RateLimiterStore, policy RateLimiterPolicy, limit uint64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := limiter.Incr(policy(r))

			if err != nil {
				errorscheme.WithError(r, &errorscheme.AppError{
					Code:     http.StatusInternalServerError,
					Message:  "rate limit error",
					Internal: err,
				})

				return
			}

			remaining := int(limit - c.hits)
			if remaining < 0 {
				remaining = 0
			}

			// Add
			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(int64(limit), 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(int64(remaining), 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(int64(c.resetIn), 10))

			if int(limit-c.hits) < 0 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type redisStore struct {
	client   *redis.Client
	script   string
	eviction time.Duration
}

func NewRedisStore(client *redis.Client, eviction time.Duration) RateLimiterStore {
	// Check if we can connect to Redis
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("RateLimiter middleware: failed to connect to Redis: %v", err)
	}

	return &redisStore{
		client:   client,
		script:   "721ec230be7c41aaf7c8fcf6413575a0d5dee104",
		eviction: eviction,
	}
}

func (store *redisStore) Incr(key string) (*Counter, error) {
	result, err := store.client.EvalSha(context.Background(), store.script, []string{key}, store.eviction.Seconds()).Slice()

	if err != nil {
		return nil, fmt.Errorf("error executing counter increment script: %v", err)
	}

	counter := &Counter{
		hits:    uint64(result[0].(int64)),
		resetIn: uint64(result[1].(int64)),
	}

	return counter, nil
}

type memoryStore struct {
	mutex    sync.RWMutex
	counters map[string]*counter

	allowed  uint64
	eviction time.Duration
}

func NewMemoryStore(window time.Duration) RateLimiterStore {
	return &memoryStore{
		eviction: window,

		// counter holds the number of hits issued per
		// "window" period per client provided Incr() key
		counters: map[string]*counter{},
	}
}

type counter struct {
	createdAt time.Time
	value     atomic.Uint64
}

func (store *memoryStore) Incr(key string) (*Counter, error) {
	store.mutex.Lock()
	c, ok := store.counters[key]

	if !ok {
		// Counter first time hit: create it
		c = &counter{
			createdAt: time.Now(),
		}

		store.counters[key] = c

		// Defer counter eviction scheme, we remove
		// the counter instead of setting it to zero
		// because additionally helps us free unused
		// resources (evicted counters for clients that
		// never issued hits again will be clean up)

		go func() {
			time.Sleep(store.eviction)

			store.mutex.Lock()
			defer store.mutex.Unlock()
			delete(store.counters, key)
		}()
	}

	store.mutex.Unlock()

	// NOTE: it is safe to use counter nodes without lock,
	// as eviction thread will simply remove nodes from map
	// but will not manipulate nodes values. Additionally,
	// counters are atomic, so also no risk of access here,
	// map access lock is bottleneck to all requests, atomic
	// counter locks are bottlenecks per counter matching
	// key, thus we separate locks per map and counters

	c.value.Add(1)

	resetIn := c.createdAt.Add(store.eviction).Sub(time.Now()).Seconds()
	if resetIn <= 0 {
		resetIn = 0
	}

	return &Counter{
		hits:    c.value.Load(),
		resetIn: uint64(resetIn),
	}, nil
}
