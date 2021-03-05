# Ratelimit

A Simple Rate Limiting Library

## Description

This library provides a RateLimiter which implements sync.Locker and a Limiter interface. 
RateLimiter is intended to be used like a Mutex to provide a timed lock for doing jobs. It also 
provides a method for users to provide their own Increase and Decrease functions to modify the 
behavior of the RateLimiter.

## Example

This example demonstrates some basic functionality with using the RateLimiter like a Mutex, but 
also by increasing and decreasing the rate limit.
```go
package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
	"github.com/stew3254/ratelimit"
)

func doWork(l ratelimit.Limiter, wg *sync.WaitGroup, i int) {
	defer wg.Done()
	l.Lock()
	log.Println("Test", i)
	l.Unlock()
}

func main() {
	r := ratelimit.NewRateLimiter(1, 10, time.Second, time.Second)
	go func() {
		for {
			n := rand.Intn(3)
			if n == 0 {
				r.Increase()
			} else {
				r.Decrease()
			}
			time.Sleep(100*time.Millisecond)
		}
	}()
	
	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go doWork(r, &wg, i)
	}
	
	wg.Wait()
}
```