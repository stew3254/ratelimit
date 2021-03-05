package main

import (
	"log"
	"sync"
	"time"
)

// Limiter is the abstract form of a Rate Limiter
// RateLimit implements this interface
type Limiter interface {
	Lock()
	Unlock()
	AddTokens(tokens int)
	AcquireWithCost(tokens int)
	SetLimit(duration time.Duration)
	IncreaseLimit(changer LimitChanger)
	DecreaseLimit(changer LimitChanger)
}

// LimitChanger is used for dependency injection
// Allows you to specify your own limit changing algorithms
type LimitChanger interface {
	Increase(limit time.Duration, states ...interface{}) (time.Duration, error)
	Decrease(limit time.Duration, states ...interface{}) (time.Duration, error)
}

// RateLimiter follows the token bucket algorithm approach
type RateLimiter struct {
	lock       sync.Mutex
	timeLock   sync.Mutex
	lastLocked time.Time
	Unit       time.Duration
	WaitLimit  time.Duration
	Tokens     int
	MaxTokens  int
}

// NewRateLimiter is a shortcut for making a new rate limiter
func NewRateLimiter(tokens, maxTokens int, wait, unit time.Duration) (r *RateLimiter) {
	return &RateLimiter{
		lock:       sync.Mutex{},
		timeLock:   sync.Mutex{},
		lastLocked: time.Now(),
		Unit:       unit,
		WaitLimit:  wait,
		Tokens:     tokens,
		MaxTokens:  maxTokens,
	}
}

// Get the floor of the number of Tokens we could have accrued by now
func (r *RateLimiter) updateTokens() {
	// This way tokens is never negative
	tokens := max(int(time.Now().Sub(r.lastLocked)/r.WaitLimit), 0)
	r.Tokens = min(tokens + r.Tokens, r.MaxTokens)
}

// Lock will get a blocking get the lock. This way it can be used like a mutex
func (r *RateLimiter) Lock() {
	// This already implements all of the functionality lock needs
	r.AcquireWithCost(1)
}

// Unlock the mutex
func (r *RateLimiter) Unlock() {
	r.lock.Unlock()
}

// Add up to the max tokens back into the pool
func (r *RateLimiter) AddTokens(tokens int) {
	r.timeLock.Lock()
	r.Tokens = min(r.MaxTokens, r.Tokens + tokens)
	r.timeLock.Unlock()
}

// Acquire the lock at a probably higher cost than 1 token
func (r *RateLimiter) AcquireWithCost(tokens int) {
	// Wait to get the lock
	for {
		// First check time in a safe way
		r.timeLock.Lock()
		// Check to see if we have enough Tokens to do our action
		if r.Tokens >= min(max(tokens, 1), r.MaxTokens) {
			// Say when last thing happened
			r.lastLocked = time.Now()
			// Get the possible amount of Tokens accrued by now and decrement by 1
			r.updateTokens()
			r.Tokens -= tokens
			// Formal lock is acquired
			r.lock.Lock()
			// Can finally unlock the other lock
			r.timeLock.Unlock()
			return
		}
		// Make sure to update Tokens to see if we got any more
		r.updateTokens()
		// Unlock if we didn't couldn't acquire the main lock
		r.timeLock.Unlock()
		// Hopefully this is okay and doesn't waste lots of time
		time.Sleep(r.WaitLimit - time.Now().Sub(r.lastLocked))
	}
}

// SetLimit allows you to specify what the duration should be
func (r *RateLimiter) SetLimit(limit time.Duration) {
	r.timeLock.Lock()
	r.WaitLimit = limit
	r.timeLock.Unlock()
}

// Increase the rate limit
func (r *RateLimiter) IncreaseLimit(changer LimitChanger) {
	r.timeLock.Lock()
	// This way the changer has access to the variables it needs for its algorithm
	limit, err := changer.Increase(r.WaitLimit, r)
	if err != nil {
		log.Println(err)
	} else {
		r.WaitLimit = limit
	}
	r.timeLock.Unlock()
}

// Decrease the rate limit
func (r *RateLimiter) DecreaseLimit(changer LimitChanger) {
	r.timeLock.Lock()
	// This way the changer has access to the variables it needs for its algorithm
	limit, err := changer.Decrease(r.WaitLimit, r)
	if err != nil {
		log.Println(err)
	} else {
		r.WaitLimit = limit
	}
	r.timeLock.Unlock()
}

// Increase uses the built in LimitChange
func (r *RateLimiter) Increase() {
	r.IncreaseLimit(&LimitChange{})
}

// Decrease uses the built in LimitChange
func (r *RateLimiter) Decrease() {
	r.DecreaseLimit(&LimitChange{})
}
