package main

import (
	"fmt"
	"time"
)

// --- 1. Basic worker ---
func worker(jobs <-chan string) {
	for job := range jobs {
		fmt.Printf("Processing: %s\n", job)
		time.Sleep(100 * time.Millisecond) // fake a processing time
	}
	fmt.Println("Worker stopped")
}

func main() {
	fmt.Println("\n=== 1. Basic worker ===")
	jobs := make(chan string)

	go worker(jobs)

	jobs <- "task-1"
	jobs <- "task-2"
	jobs <- "task-3"

	close(jobs)

	time.Sleep(500 * time.Millisecond)
}
