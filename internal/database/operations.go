package database

import (
	"context"
	"fmt"
	"rinha-go-2025/internal/models"
	"sync"
	"time"
)

var (
	batchMutex         sync.Mutex
	batchQueueDefault  []models.Payment
	batchQueueFallback []models.Payment
	batchSize          = 10
	lastFlush          = time.Now()
)

func SavePayment(payment models.Payment, handler string) error {
	batchMutex.Lock()
	defer batchMutex.Unlock()

	if handler == "default" {
		batchQueueDefault = append(batchQueueDefault, payment)
	} else {
		batchQueueFallback = append(batchQueueFallback, payment)
	}

	// Flush se atingiu o tamanho do batch ou passou muito tempo
	totalBatchSize := len(batchQueueDefault) + len(batchQueueFallback)
	if totalBatchSize >= batchSize || time.Since(lastFlush) > 10*time.Millisecond {
		return flushAllBatches()
	}

	return nil
}

// Executa o batch insert para todos os handlers
func flushAllBatches() error {
	if len(batchQueueDefault) > 0 {
		if err := flushBatch(batchQueueDefault, "default"); err != nil {
			return err
		}
		batchQueueDefault = batchQueueDefault[:0]
	}

	if len(batchQueueFallback) > 0 {
		if err := flushBatch(batchQueueFallback, "fallback"); err != nil {
			return err
		}
		batchQueueFallback = batchQueueFallback[:0]
	}

	lastFlush = time.Now()
	return nil
}

// flushBatch executa o batch insert para um handler específico
func flushBatch(payments []models.Payment, handler string) error {
	if len(payments) == 0 {
		return nil
	}

	db, err := GetConnectionPool()
	if err != nil {
		return err
	}

	query := "INSERT INTO payments (correlation_id, amount, handler, created_at) VALUES "
	args := make([]interface{}, 0, len(payments)*4)

	for i, payment := range payments {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($%d, $%d, $%d, $%d)",
			i*4+1, i*4+2, i*4+3, i*4+4)

		args = append(args, payment.CorrelationID, payment.Amount, handler, payment.RequestedAt)
	}

	// Executar batch insert com retry reduzido
	var execErr error
	for retries := 0; retries < 1; retries++ {
		_, execErr = db.Exec(context.Background(), query, args...)
		if execErr == nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	return execErr
}

// Retorna o resumo de pagamentos por período
func GetPaymentsSummary(from, to string) (*models.PaymentSummary, error) {
	ForceFlushBatch()

	db, err := GetConnectionPool()
	if err != nil {
		return nil, err
	}

	query := `
		SELECT handler, COUNT(*) as total_requests, SUM(amount) as total_amount
		FROM payments
		WHERE created_at >= $1::timestamp AND created_at <= $2::timestamp
		GROUP BY handler
	`

	rows, err := db.Query(context.Background(), query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary := &models.PaymentSummary{}
	for rows.Next() {
		var handler string
		var totalRequests int
		var totalAmount float64

		rows.Scan(&handler, &totalRequests, &totalAmount)

		if handler == "default" {
			summary.Default = models.Summary{TotalRequests: totalRequests, TotalAmount: totalAmount}
		} else {
			summary.Fallback = models.Summary{TotalRequests: totalRequests, TotalAmount: totalAmount}
		}
	}

	return summary, nil
}

// ForceFlushBatch força o flush do batch atual
func ForceFlushBatch() {
	batchMutex.Lock()
	defer batchMutex.Unlock()

	flushAllBatches()
}

// PurgePayments remove todos os pagamentos do banco
func PurgePayments() error {
	ForceFlushBatch()

	pool, err := GetConnectionPool()
	if err != nil {
		return fmt.Errorf("erro ao obter pool de conexões: %v", err)
	}

	ctx := context.Background()
	_, err = pool.Exec(ctx, "DELETE FROM payments")
	if err != nil {
		return fmt.Errorf("erro ao limpar pagamentos: %v", err)
	}

	return nil
}
