package models

import (
	"sync"
	"time"
)

// Representa um pagamento recebido
type Payment struct {
	CorrelationID string    `json:"correlationId"`
	Amount        float64   `json:"amount"`
	RequestedAt   time.Time `json:"requestedAt"`
}

// Representa o resumo de pagamentos por processor
type Summary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

// Representa o resumo total de pagamentos
type PaymentSummary struct {
	Default  Summary `json:"default"`
	Fallback Summary `json:"fallback"`
}

// Representa um trabalho de pagamento
type PaymentJob struct {
	Payment Payment
	Result  chan error
}

type ProcessorStatus struct {
	URL             string
	Failing         bool
	MinResponseTime int
	LastUpdate      time.Time
	Mu              sync.RWMutex
}

type PaymentProcessor struct {
	DefaultProcessor  *ProcessorStatus
	FallbackProcessor *ProcessorStatus
	Cache             map[string]*ProcessorStatus
	Mu                sync.RWMutex
}
