package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// --- Email worker ---

type EmailJob struct {
	To      string
	Subject string
	Body    string
}

type EmailSender interface {
	Send(to, subject, body string) error
}

// mockEmailSender fakes an SMTP round trip so the /signup handler can
// demonstrate returning immediately while the email is still "in flight".
type mockEmailSender struct{}

func (mockEmailSender) Send(to, subject, body string) error {
	log.Printf("Sending email to %s: %s\n", to, subject)
	time.Sleep(2 * time.Second)
	log.Printf("Email sent to %s\n", to)
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
	return w
}

// Start begins processing email jobs from the queue until the job channel is closed.
//
// It runs in a separate goroutine and should be started by calling Start. It drains
// any remaining buffered jobs before returning, so callers must stop enqueuing new
// work before calling Shutdown.
func (w *EmailWorker) Start() {
	w.wg.Add(1)
	defer w.wg.Done()
	for job := range w.jobs {
		if err := w.sender.Send(job.To, job.Subject, job.Body); err != nil {
			log.Printf("Failed to send email to %s: %v", job.To, err)
		}
	}
	log.Println("Email worker stopped: queue drained")
}

var ErrQueueFull = errors.New("email queue is full")

// Enqueue never blocks: if the buffer is full, it returns ErrQueueFull
func (w *EmailWorker) Enqueue(job EmailJob) error {
	select {
	case w.jobs <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

func (w *EmailWorker) Shutdown() {
	close(w.jobs)
	w.wg.Wait()
}

// --- HTTP API ---
type signupRequest struct {
	Email string `json:"email"`
}

func signupHandler(worker *EmailWorker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req signupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
			http.Error(w, "invalid or missing email", http.StatusBadRequest)
			return
		}

		// The welcome email is sent in the background: the HTTP response
		// does not wait for the 4-second mock SMTP round trip.
		err := worker.Enqueue(EmailJob{
			To:      req.Email,
			Subject: "Bienvenue !",
			Body:    "Votre compte est créé.",
		})
		if err != nil {
			http.Error(w, "server busy, try again later", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	worker := NewEmailWorker(mockEmailSender{}, 100)
	go worker.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /signup", signupHandler(worker))

	server := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		log.Println("Listening on :8080 (POST /signup)")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	worker.Shutdown()
}
