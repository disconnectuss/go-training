package main

import (
	"fmt"
	"sync"
	"time"
)

// Step 9: Goroutine & Channel — Go's concurrency model
//
// GOROUTINE: A lightweight thread managed by Go runtime (not OS thread)
//   - Start with "go" keyword: go myFunction()
//   - Costs only ~2KB of memory (OS thread = ~1MB)
//   - Go can run millions of goroutines simultaneously
//
// CHANNEL: A pipe that connects goroutines — they send/receive values through it
//   - Create with make(chan Type)
//   - Send:    ch <- value
//   - Receive: value := <-ch
//   - Channels BLOCK: send waits for receiver, receive waits for sender

// --- Example 1: Basic Goroutine ---

// RunBasicGoroutine shows how goroutines run concurrently
func RunBasicGoroutine() {
	fmt.Println("\n--- Example 1: Basic Goroutine ---")

	// "go" keyword starts a new goroutine — it runs CONCURRENTLY with main
	go func() {
		// This is an ANONYMOUS FUNCTION (lambda) running in its own goroutine
		for i := 1; i <= 3; i++ {
			fmt.Printf("  goroutine: %d\n", i)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Main goroutine continues without waiting
	for i := 1; i <= 3; i++ {
		fmt.Printf("  main: %d\n", i)
		time.Sleep(100 * time.Millisecond)
	}
	// Output is INTERLEAVED — both run at the same time!
}

// --- Example 2: Channel — sending data between goroutines ---

// RunChannelExample shows basic channel communication
func RunChannelExample() {
	fmt.Println("\n--- Example 2: Channel ---")

	// make(chan string) creates an UNBUFFERED channel
	// Unbuffered = sender blocks until receiver is ready (synchronization)
	messages := make(chan string)

	// Start a goroutine that SENDS a message
	go func() {
		time.Sleep(200 * time.Millisecond) // Simulate some work
		messages <- "hello from goroutine"  // Send to channel (blocks until someone receives)
	}()

	// Main goroutine RECEIVES the message
	// This line BLOCKS until the goroutine sends something
	msg := <-messages
	fmt.Printf("  received: %s\n", msg)
}

// --- Example 3: Buffered Channel ---

// RunBufferedChannel shows channels with capacity
func RunBufferedChannel() {
	fmt.Println("\n--- Example 3: Buffered Channel ---")

	// Buffered channel — can hold 2 values without blocking
	// Sender only blocks when buffer is FULL
	// Receiver only blocks when buffer is EMPTY
	ch := make(chan string, 2)

	ch <- "first"  // Doesn't block — buffer has space
	ch <- "second" // Doesn't block — buffer still has space
	// ch <- "third" // Would block! Buffer is full, no one is receiving

	fmt.Printf("  %s\n", <-ch) // "first" (FIFO order)
	fmt.Printf("  %s\n", <-ch) // "second"
}

// --- Example 4: Channel Direction — send-only and receive-only ---

// producer only SENDS to the channel (chan<- = send-only)
// The arrow shows the direction: data goes INTO the channel
func producer(ch chan<- int, count int) {
	for i := 1; i <= count; i++ {
		ch <- i * 10 // Send values: 10, 20, 30...
	}
	close(ch) // Close the channel — no more values will be sent
	// Receivers can still read remaining values after close
}

// consumer only RECEIVES from the channel (<-chan = receive-only)
func consumer(ch <-chan int) {
	// range on a channel: reads values until the channel is CLOSED
	// Without close(), this would block forever (deadlock!)
	for val := range ch {
		fmt.Printf("  consumed: %d\n", val)
	}
}

// RunDirectionalChannel demonstrates send-only and receive-only channels
func RunDirectionalChannel() {
	fmt.Println("\n--- Example 4: Directional Channel ---")

	ch := make(chan int, 5) // Both directions at creation
	producer(ch, 3)         // Function sees it as send-only
	consumer(ch)            // Function sees it as receive-only
}

// --- Example 5: Select — listening on multiple channels ---

// RunSelectExample shows how to wait on multiple channels simultaneously
func RunSelectExample() {
	fmt.Println("\n--- Example 5: Select ---")

	ch1 := make(chan string)
	ch2 := make(chan string)

	// Two goroutines sending at different speeds
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch1 <- "result from service A"
	}()

	go func() {
		time.Sleep(200 * time.Millisecond)
		ch2 <- "result from service B"
	}()

	// select waits on MULTIPLE channels — like a switch for channels
	// Whichever channel is ready first, that case runs
	for i := 0; i < 2; i++ {
		select {
		case msg := <-ch1:
			fmt.Printf("  ch1: %s\n", msg)
		case msg := <-ch2:
			fmt.Printf("  ch2: %s\n", msg)
		}
	}
}

// --- Example 6: WaitGroup — waiting for multiple goroutines to finish ---

// RunWaitGroupExample shows sync.WaitGroup for coordinating goroutines
func RunWaitGroupExample() {
	fmt.Println("\n--- Example 6: WaitGroup ---")

	// WaitGroup waits for a collection of goroutines to finish
	// Think of it as a counter: Add(1) increments, Done() decrements, Wait() blocks until 0
	var wg sync.WaitGroup

	names := []string{"Fatma", "Ahmet", "Elif"}

	for _, name := range names {
		wg.Add(1) // Tell WaitGroup: "one more goroutine to wait for"

		go func(n string) {
			defer wg.Done() // When this goroutine finishes, decrement the counter

			time.Sleep(100 * time.Millisecond)
			fmt.Printf("  done processing: %s\n", n)
		}(name) // Pass name as argument — don't capture loop variable!
		// Why pass as arg? Loop variable changes each iteration
		// All goroutines would see the LAST value without this
	}

	wg.Wait() // Block until all goroutines call Done()
	fmt.Println("  all goroutines finished!")
}

// --- Example 7: Worker Pool — practical concurrency pattern ---

// Job represents a task to be processed
type Job struct {
	ID    int
	Input string
}

// Result represents the output of a processed job
type Result struct {
	JobID  int
	Output string
}

// worker processes jobs from the jobs channel and sends results to the results channel
// Multiple workers run concurrently — this is the WORKER POOL pattern
func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs { // Read jobs until channel is closed
		// Simulate processing time
		time.Sleep(50 * time.Millisecond)

		result := Result{
			JobID:  job.ID,
			Output: fmt.Sprintf("worker-%d processed '%s'", id, job.Input),
		}
		results <- result // Send result
	}
}

// RunWorkerPool demonstrates the worker pool pattern
// 3 workers process 5 jobs concurrently
func RunWorkerPool() {
	fmt.Println("\n--- Example 7: Worker Pool ---")

	jobs := make(chan Job, 5)       // Buffered: holds up to 5 jobs
	results := make(chan Result, 5) // Buffered: holds up to 5 results

	// Start 3 workers
	var wg sync.WaitGroup
	for w := 1; w <= 3; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// Send 5 jobs
	jobInputs := []string{"parse", "validate", "transform", "save", "notify"}
	for i, input := range jobInputs {
		jobs <- Job{ID: i + 1, Input: input}
	}
	close(jobs) // No more jobs — workers will exit their range loop

	// Wait for all workers to finish, then close results
	go func() {
		wg.Wait()
		close(results) // Safe to close after all workers are done
	}()

	// Collect all results
	for result := range results {
		fmt.Printf("  job %d: %s\n", result.JobID, result.Output)
	}
}
