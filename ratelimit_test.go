package main

import (
	"log"
	"sync"
	"testing"
	"time"
)

type changeLimitTest struct {
	limit, unit, expected time.Duration
}

var increaseTests = []changeLimitTest{
	{time.Second, time.Second, (3*time.Second)/2},
	{5*time.Second, time.Second, (15*time.Second)/2},
	{500*time.Millisecond, time.Second, time.Second},
	{100*time.Millisecond, time.Second, 200*time.Millisecond},
	{300*time.Millisecond, time.Second, 600*time.Millisecond},
	{800*time.Millisecond, time.Second, time.Second},
	{1000231*time.Microsecond, time.Second, (3*time.Second)/2},
	{0, time.Second, time.Second},
}

var decreaseTests = []changeLimitTest{
	{time.Second, time.Second, 900*time.Millisecond},
	{5*time.Second, time.Second, 4*time.Second},
	{1100*time.Millisecond, time.Second, time.Second},
	{500*time.Millisecond, time.Second, 400*time.Millisecond},
	{100*time.Millisecond, time.Second, 90*time.Millisecond},
	{101*time.Millisecond, time.Second, 100*time.Millisecond},
	{0, time.Second, 0},
}

func TestIncrease(t *testing.T) {
	r := NewRateLimiter(1, 10, time.Second, time.Second)
	for _, test := range increaseTests {
		r.WaitLimit = test.limit
		r.Unit = test.unit
		r.Increase()
		if r.WaitLimit != test.expected {
			t.Errorf("output %q does not match expected %q", r.WaitLimit, test.expected)
		}
	}
}

func TestDecrease(t *testing.T) {
	r := NewRateLimiter(1, 10, time.Second, time.Second)
	for _, test := range decreaseTests {
		r.WaitLimit = test.limit
		r.Unit = test.unit
		r.Decrease()
		if r.WaitLimit != test.expected {
			t.Errorf("output %q does not match expected %q", r.WaitLimit, test.expected)
		}
	}
}

func TestLocking(t *testing.T) {
	r := NewRateLimiter(1, 1, time.Second, time.Second)
	wg := sync.WaitGroup{}
	r.Lock()
	wg.Add(1)
	go func() {
		r.Lock()
		log.Println("I come second")
		r.Unlock()
		wg.Done()
	}()
	time.Sleep(time.Second)
	log.Println("I come first")
	r.Unlock()
	wg.Wait()
}

func TestRateLimiting(t *testing.T) {
	r := NewRateLimiter(3, 5, time.Second, time.Second)
	r.SetLimit(100*time.Millisecond)
	wg := sync.WaitGroup{}
	p := func(i int) {
		// Choose to use regular lock
		if i % 2 == 0 {
			r.Lock()
		} else {
			// Make it more expensive to lock
			r.AcquireWithCost(2)
		}
		log.Println(i)
		r.Unlock()
		wg.Done()
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go p(i)
	}
	wg.Wait()
}

func TestRateLimiter_AddTokens(t *testing.T) {
	r := NewRateLimiter(0, 10, time.Hour, time.Second)
	r.SetLimit(100*time.Millisecond)
	r.AddTokens(5)
	for i := 0; i < 5; i++ {
		r.Lock()
		log.Println("Added tokens test", i)
		r.Unlock()
	}
	
}