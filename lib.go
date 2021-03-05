package main

import (
	"errors"
	"time"
)

type LimitChange struct {}

// Increase the rate limit.
// LimitChange requires 1 argument of type time.Duration to be passed in
// This argument is the unit by which the piecewise function works around
func (l *LimitChange) Increase(limit time.Duration, states ...interface{}) (time.Duration, error) {
	// Error since we didn't get what we wanted
	if len(states) == 0 {
		return 0, errors.New("need a proper time unit (time.Duration) to be passed in")
	}

	// Expect a RateLimiter to be passed in so we can use its state
	r := states[0].(*RateLimiter)

	// We have no limit, so just give it one of the unit
	if limit == 0 {
		limit = 1 * r.Unit
		// If we have a fractional r.Unit
	} else if limit < r.Unit {
		// limit = limit + 5*r.Unit
		limit *= 2
		// If it's less than or equal to zero just set it equal to that r.Unit size
		// ie. limit 6/10s -> 1 because 6ds + 5ds == 11ds == 1.1s
		if limit > r.Unit {
			limit = r.Unit
		}

		// We don't have a fractional r.Unit and we can just add 5 times that r.Unit to it
	} else {
		// Multiply by 1.5x
		limit *= 3
		limit /= 2
		// This will chop off everything after 3 decimal places
		limit -= limit %(r.Unit/1000)
	}
	return limit, nil
}

// Decrease the rate limit
func (l *LimitChange) Decrease(limit time.Duration, states ...interface{}) (time.Duration, error) {
	// Error since we didn't get what we wanted
	if len(states) == 0 {
		return 0, errors.New("need a proper time unit (time.Duration) to be passed in")
	}

	// Expect a RateLimiter to be passed in so we can use its state
	r := states[0].(*RateLimiter)

	// We have no limit so just leave it alone
	if limit == 0 {
		return limit, nil
		// If we are above the ratelimit
	} else if limit > r.Unit {
		// Subtract a r.Unit from the wait limit
		limit -= r.Unit
		// Just to clean up the numbers a bit set it to a measure of that r.Unit
		if limit < r.Unit {
			limit = r.Unit
		}
	} else {
		var newUnit time.Duration = r.Unit/10
		for limit <= newUnit {
			newUnit /= 10
		}

		// Subtract by the new unit
		limit -= newUnit

		if limit < newUnit {
			limit = newUnit
		}
	}
	return limit, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
