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
		paymentQueue: make(chan models.PaymentJob, 200000), // Buffer de 200k
		workerPool:   make(chan struct{}, 1000),            // 1000 workers
	}
}

func (w *Worker) Start() error {
	w.startWorkerPool()
	return nil
}

func (w *Worker) startWorkerPool() {
	for i := 0; i < 1000; i++ {
		go w.worker()
	}
	// fmt.Println("Worker pool iniciado com 1000 workers")
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
	_, err := proc.ProcessPayment(payment)

	return err
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
