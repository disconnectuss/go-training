package main

import (
	"sync"
	"testing"
	"time"
)

// Step 9 Test: Goroutines run concurrently (faster than sequential)
func TestGoroutinesConcurrent(t *testing.T) {
	start := time.Now()

	// Run 3 "tasks" concurrently using goroutines
	// Each task takes 100ms — if sequential, total = 300ms
	// If concurrent, total ≈ 100ms
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond)
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Should take ~100ms (concurrent), not ~300ms (sequential)
	if elapsed > 200*time.Millisecond {
		t.Errorf("expected ~100ms (concurrent), took %v (too slow, likely sequential)", elapsed)
	}
}

// Step 9 Test: Channel sends and receives values correctly
func TestChannelCommunication(t *testing.T) {
	ch := make(chan int)

	// Send value in a goroutine
	go func() {
		ch <- 42
	}()

	// Receive in the test goroutine
	received := <-ch

	if received != 42 {
		t.Errorf("expected 42, got %d", received)
	}
}

// Step 9 Test: Buffered channel holds values without blocking
func TestBufferedChannel(t *testing.T) {
	ch := make(chan string, 2)

	// These don't block because buffer has space
	ch <- "a"
	ch <- "b"

	// Receive in order (FIFO)
	first := <-ch
	second := <-ch

	if first != "a" {
		t.Errorf("expected 'a', got '%s'", first)
	}
	if second != "b" {
		t.Errorf("expected 'b', got '%s'", second)
	}
}

// Step 9 Test: close() and range on channel
func TestChannelCloseAndRange(t *testing.T) {
	ch := make(chan int, 3)

	// Send values and close
	ch <- 10
	ch <- 20
	ch <- 30
	close(ch)

	// range reads all values until channel is closed
	var received []int
	for val := range ch {
		received = append(received, val)
	}

	if len(received) != 3 {
		t.Errorf("expected 3 values, got %d", len(received))
	}
	if received[0] != 10 || received[1] != 20 || received[2] != 30 {
		t.Errorf("expected [10,20,30], got %v", received)
	}
}

// Step 9 Test: select picks the first ready channel
func TestSelectFirstReady(t *testing.T) {
	fast := make(chan string, 1)
	slow := make(chan string, 1)

	fast <- "fast wins" // Already has a value (buffered)

	// select should pick fast because it's ready immediately
	var result string
	select {
	case result = <-fast:
	case result = <-slow:
	}

	if result != "fast wins" {
		t.Errorf("expected 'fast wins', got '%s'", result)
	}
}

// Step 9 Test: Worker pool processes all jobs
func TestWorkerPool(t *testing.T) {
	jobs := make(chan Job, 5)
	results := make(chan Result, 5)

	// Start 2 workers
	var wg sync.WaitGroup
	for w := 1; w <= 2; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// Send 3 jobs
	jobs <- Job{ID: 1, Input: "task-a"}
	jobs <- Job{ID: 2, Input: "task-b"}
	jobs <- Job{ID: 3, Input: "task-c"}
	close(jobs)

	// Close results after workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var collected []Result
	for r := range results {
		collected = append(collected, r)
	}

	// All 3 jobs should produce results
	if len(collected) != 3 {
		t.Errorf("expected 3 results, got %d", len(collected))
	}
}

// Step 9 Test: WaitGroup waits for all goroutines
func TestWaitGroupCompletesAll(t *testing.T) {
	var wg sync.WaitGroup
	counter := 0
	var mu sync.Mutex // Mutex protects counter from race conditions

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// mu.Lock/Unlock prevents two goroutines from writing counter at the same time
			// Without mutex: DATA RACE — counter could be wrong
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()

	if counter != 5 {
		t.Errorf("expected counter=5, got %d", counter)
	}
}
