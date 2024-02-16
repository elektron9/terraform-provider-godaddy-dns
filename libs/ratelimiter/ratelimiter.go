package ratelimiter

import (
	"context"
	"errors"
	"sync"
	"time"
)

// simple token bucket rate limiter
// - keeps track of available tokens and the last time a token was added
// - on acquire: if there are tokens available, spend one and return immediately
// - if there are none
//   - check the time since a token was added, mb add some more
//   - spend one if now it is available
//   - if none, wait until next one will be available (while holding mutex)
type RateLimiter struct {
	mu sync.Mutex
	// interval between tokens, s
	period time.Duration
	// total bucket size; 1 means "no bursts"
	bucketSize int
	// current number of tokens in bucket
	numTokens int
	// the time last token was added
	lastTokenTime time.Time
}

// context to cancel, rate per second, burst (bucket) size
func New(rate, burst int) (*RateLimiter, error) {
	if rate <= 0 || burst <= 0 {
		return nil, errors.New("limiter rate and burst must be positive")
	}
	return &RateLimiter{
		period:        time.Second / time.Duration(rate),
		bucketSize:    burst,
		numTokens:     burst,
		lastTokenTime: time.Now(),
	}, nil
}

// block until token becomes available
func (s *RateLimiter) Wait() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// if bucket is empty: wait for 60 seconds since the last token was used, refill the bucket to (bucketSize), 
	// and set the last token time to time.Now() to set the new start of the period.
	// this is unique to the GoDaddy APIs, which limit us to 60 requests every minute per endpoint.
	// See: https://developer.godaddy.com/getstarted#terms
	if s.numTokens <= 0 {
		waitDuration := time.Since(s.lastTokenTime.Add(time.Second * 60))
		time.Sleep(waitDuration)
		s.numTokens = s.bucketSize
		s.lastTokenTime = time.Now()
	} else {
		// if token available: ok, take it and release caller from the wait
		s.numTokens--
		s.lastTokenTime = time.Now()
		return
	}
}

// block until token becomes available, with cancellable context
func (s *RateLimiter) WaitCtx(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// if bucket is empty: wait for 60 seconds since the last token was used, refill the bucket to (bucketSize), 
	// and set the last token time to time.Now() to set the new start of the period.
	// this is unique to the GoDaddy APIs, which limit us to 60 requests every minute per endpoint.
	// See: https://developer.godaddy.com/getstarted#terms
	if s.numTokens <= 0 {
		waitDuration := time.Since(s.lastTokenTime.Add(time.Second * 60))
		s.numTokens = s.bucketSize

		select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitDuration):
		}
		// if not cancelled: upd last token time and release from the wait
		s.lastTokenTime = time.Now()
		return nil
	}
	// if token available: ok, take it and release caller from the wait
	if s.numTokens > 0 {
		s.numTokens--
		s.lastTokenTime = time.Now()
		return nil
	}
	return nil
}
