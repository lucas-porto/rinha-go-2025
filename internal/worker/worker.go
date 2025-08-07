package worker

import (
	"context"
	"fmt"
	"rinha-go-2025/internal/database"
	"rinha-go-2025/internal/models"
	"rinha-go-2025/internal/processor"
)

type Worker struct {
	paymentQueue chan models.PaymentJob
	workerPool   chan struct{}
}

func New() *Worker {
	return &Worker{
		paymentQueue: make(chan models.PaymentJob, 100000), // Buffer de 100k
		workerPool:   make(chan struct{}, 500),             // 500 workers
	}
}

func (w *Worker) Start() error {
	w.startWorkerPool()
	return nil
}

func (w *Worker) startWorkerPool() {
	for i := 0; i < 500; i++ {
		go w.worker()
	}
	// fmt.Println("Worker pool iniciado com 500 workers")
}

func (w *Worker) worker() {
	for job := range w.paymentQueue {
		w.workerPool <- struct{}{}

		// Processar pagamento
		err := w.processPayment(job.Payment)

		job.Result <- err
		close(job.Result)

		<-w.workerPool
	}
}

func (w *Worker) EnqueuePayment(payment models.Payment) error {
	resultChan := make(chan error, 1)
	job := models.PaymentJob{
		Payment: payment,
		Result:  resultChan,
	}

	select {
	case w.paymentQueue <- job:
		return nil
	default:
		return fmt.Errorf("queue full")
	}
}

func (w *Worker) processPayment(payment models.Payment) error {
	// Usar o processor real
	proc := processor.GetProcessor()
	return proc.ProcessPayment(payment)
}

func (w *Worker) TestDatabaseConnection() error {
	pool, err := database.GetConnectionPool()
	if err != nil {
		return fmt.Errorf("erro ao obter pool de conexões: %v", err)
	}

	ctx := context.Background()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("erro ao fazer ping no banco: %v", err)
	}

	//fmt.Println("Conexão com banco estabelecida.")
	return nil
}
