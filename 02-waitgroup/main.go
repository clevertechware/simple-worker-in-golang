package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(jobs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		fmt.Printf("Processing: %s\n", job)
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	fmt.Println("\n=== 2. Worker with WaitGroup ===")
	jobs := make(chan string)
	var wg sync.WaitGroup

	wg.Add(1)
	go worker(jobs, &wg)

	jobs <- "task-1"
	jobs <- "task-2"
	jobs <- "task-3"

	close(jobs)
	wg.Wait()
}
