package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// --- 3. Production-style email worker ---
type EmailJob struct {
	To      string
	Subject string
	Body    string
}

type EmailSender interface {
	Send(to, subject, body string) error
}

type mockEmailSender struct{}

func (mockEmailSender) Send(to, subject, body string) error {
	fmt.Printf("Sending email to %s: %s\n", to, subject)
	time.Sleep(100 * time.Millisecond)
	return nil
}

type EmailWorker struct {
	jobs   chan EmailJob
	wg     sync.WaitGroup
	sender EmailSender
}

func NewEmailWorker(sender EmailSender, bufferSize int) *EmailWorker {
	w := &EmailWorker{
		jobs:   make(chan EmailJob, bufferSize),
		sender: sender,
	}
	w.wg.Add(1)
	go w.run()
	return w
}

func (w *EmailWorker) run() {
	defer w.wg.Done()
	for job := range w.jobs {
		if err := w.sender.Send(job.To, job.Subject, job.Body); err != nil {
			log.Printf("Failed to send email to %s: %v", job.To, err)
		}
	}
}

func (w *EmailWorker) Enqueue(job EmailJob) {
	w.jobs <- job
}

func (w *EmailWorker) Shutdown() {
	close(w.jobs)
	w.wg.Wait()
}

func demoEmailWorker() {
	fmt.Println("\n=== 3. Production email worker ===")
	worker := NewEmailWorker(mockEmailSender{}, 100)

	worker.Enqueue(EmailJob{
		To:      "alice@example.com",
		Subject: "Bienvenue !",
		Body:    "Votre compte est créé.",
	})
	worker.Enqueue(EmailJob{
		To:      "bob@example.com",
		Subject: "Confirmation",
		Body:    "Votre commande est confirmée.",
	})

	worker.Shutdown()
}

// --- 4. Context-aware worker ---

func contextAwareWorker(ctx context.Context, jobs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Worker stopped: context cancelled")
			return
		case job, ok := <-jobs:
			if !ok {
				fmt.Println("Worker stopped: channel closed")
				return
			}
			fmt.Printf("Processing: %s\n", job)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func demoContextAwareWorker() {
	fmt.Println("\n=== 4. Context-aware worker ===")
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	jobs := make(chan string)
	var wg sync.WaitGroup

	wg.Add(1)
	go contextAwareWorker(ctx, jobs, &wg)

	// Send jobs until the context deadline cuts the worker off.
	for i := 1; i <= 10; i++ {
		select {
		case jobs <- fmt.Sprintf("task-%d", i):
		case <-ctx.Done():
			wg.Wait()
			return
		}
	}
	wg.Wait()
}

func main() {
	demoEmailWorker()
	demoContextAwareWorker()
}
