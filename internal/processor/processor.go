package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"rinha-go-2025/internal/database"
	"rinha-go-2025/internal/models"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

type Processor struct {
	paymentProcessor *models.PaymentProcessor
	defaultClient    *fasthttp.Client
	fallbackClient   *fasthttp.Client
}

var (
	processor *Processor
	once      sync.Once
)

func GetProcessor() *Processor {
	once.Do(func() {
		defaultURL := os.Getenv("DEFAULT_PROCESSOR_URL")
		if defaultURL == "" {
			defaultURL = "http://payment-processor-default:8080"
		}

		fallbackURL := os.Getenv("FALLBACK_PROCESSOR_URL")
		if fallbackURL == "" {
			fallbackURL = "http://payment-processor-fallback:8080"
		}

		paymentProc := &models.PaymentProcessor{
			DefaultProcessor: &models.ProcessorStatus{
				URL:             defaultURL,
				Failing:         false,
				MinResponseTime: 0,
				LastUpdate:      time.Now(),
			},
			FallbackProcessor: &models.ProcessorStatus{
				URL:             fallbackURL,
				Failing:         false,
				MinResponseTime: 0,
				LastUpdate:      time.Now(),
			},
			Cache: make(map[string]*models.ProcessorStatus),
		}

		paymentProc.Cache["default"] = paymentProc.DefaultProcessor
		paymentProc.Cache["fallback"] = paymentProc.FallbackProcessor

		processor = &Processor{
			paymentProcessor: paymentProc,
		}

		// Inicializar connection pools
		processor.defaultClient = &fasthttp.Client{
			ReadTimeout:         300 * time.Millisecond,
			WriteTimeout:        300 * time.Millisecond,
			MaxConnsPerHost:     1000,
			MaxIdleConnDuration: 10 * time.Second,
		}

		processor.fallbackClient = &fasthttp.Client{
			ReadTimeout:         3 * time.Second,
			WriteTimeout:        3 * time.Second,
			MaxConnsPerHost:     100,
			MaxIdleConnDuration: 10 * time.Second,
		}

		processor.loadStatusFromDB()
	})
	return processor
}

func (p *Processor) ProcessPayment(payment models.Payment) error {
	start := time.Now()

	maxRetries := 2
	baseDelay := 10 * time.Millisecond

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := p.tryProcessorWithClient(payment, p.paymentProcessor.DefaultProcessor, "default", p.defaultClient); err == nil {
			p.updateProcessorStatus("default", false, int(time.Since(start).Milliseconds()))
			return nil
		}

		if attempt < maxRetries {
			delay := baseDelay * time.Duration(1<<attempt)
			time.Sleep(delay)
		}
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := p.tryProcessorWithClient(payment, p.paymentProcessor.FallbackProcessor, "fallback", p.fallbackClient); err == nil {
			p.updateProcessorStatus("fallback", false, int(time.Since(start).Milliseconds()))
			return nil
		}

		if attempt < maxRetries {
			delay := baseDelay * time.Duration(1<<attempt)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("todos os payment processors falharam temporariamente")
}

func (p *Processor) tryProcessorWithClient(payment models.Payment, proc *models.ProcessorStatus, handler string, client *fasthttp.Client) error {
	if proc.Failing {
		return fmt.Errorf("processor %s está marcado como falhando", handler)
	}

	if time.Since(proc.LastUpdate) > 120*time.Second {
		if err := p.checkHealth(proc, handler); err != nil {
		}
	}

	body, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("erro ao serializar pagamento: %v", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(proc.URL + "/payments")
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBody(body)

	if err := client.Do(req, resp); err != nil {
		return fmt.Errorf("erro na requisição para %s: %v", handler, err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("%s retornou status: %d", handler, resp.StatusCode())
	}

	if err := database.SavePayment(payment, handler); err != nil {
		return fmt.Errorf("erro ao salvar pagamento no banco: %v", err)
	}

	return nil
}

func (p *Processor) checkHealth(proc *models.ProcessorStatus, handler string) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(proc.URL + "/payments/service-health")
	req.Header.SetMethod("GET")

	var client *fasthttp.Client
	if handler == "default" {
		client = p.defaultClient
	} else {
		client = p.fallbackClient
	}

	if err := client.Do(req, resp); err != nil {
		return err
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("health check retornou status: %d", resp.StatusCode())
	}

	return nil
}

func (p *Processor) updateProcessorStatus(handler string, failing bool, responseTime int) {
	p.paymentProcessor.Mu.Lock()
	defer p.paymentProcessor.Mu.Unlock()

	proc := p.paymentProcessor.Cache[handler]
	if proc != nil {
		proc.Failing = failing
		if responseTime > 0 {
			proc.MinResponseTime = responseTime
		}
		proc.LastUpdate = time.Now()

		// Salvar status no banco
		p.saveStatusToDB(handler, failing, responseTime)
	}
}

func (p *Processor) saveStatusToDB(handler string, failing bool, responseTime int) {
	query := `
		INSERT INTO cache (processor, failing, min_response_time, last_update) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (processor) 
		DO UPDATE SET 
			failing = EXCLUDED.failing,
			min_response_time = EXCLUDED.min_response_time,
			last_update = EXCLUDED.last_update
	`

	db, err := database.GetConnectionPool()
	if err != nil {
		return
	}

	_, err = db.Exec(context.Background(), query, handler, failing, responseTime, time.Now())
	if err != nil {
		fmt.Printf("Erro ao salvar status do processor %s: %v\n", handler, err)
	}
}

func (p *Processor) loadStatusFromDB() {
	db, err := database.GetConnectionPool()
	if err != nil {
		return
	}

	query := `SELECT processor, failing, min_response_time, last_update FROM cache`
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var processor, lastUpdateStr string
		var failing bool
		var minResponseTime int

		if err := rows.Scan(&processor, &failing, &minResponseTime, &lastUpdateStr); err != nil {
			continue
		}

		lastUpdate, _ := time.Parse(time.RFC3339, lastUpdateStr)

		if proc := p.paymentProcessor.Cache[processor]; proc != nil {
			proc.Failing = failing
			proc.MinResponseTime = minResponseTime
			proc.LastUpdate = lastUpdate
		}
	}
}
