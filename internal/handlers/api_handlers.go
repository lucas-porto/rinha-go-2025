package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"rinha-go-2025/internal/database"
	"rinha-go-2025/internal/models"
	"rinha-go-2025/internal/worker"
	"time"

	"github.com/valyala/fasthttp"
)

type APIHandlers struct {
	worker *worker.Worker
}

func NewAPIHandlersWithWorker(w *worker.Worker) *APIHandlers {
	return &APIHandlers{
		worker: w,
	}
}

// HandleHealthcheck processa requisições de health check
func HandleHealthcheck(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

// HandlePurgePayments limpa todos os pagamentos do banco
func HandlePurgePayments(ctx *fasthttp.RequestCtx) {
	if err := database.PurgePayments(); err != nil {
		ctx.Error("Database error", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// Processa requisições de resumo de pagamentos
func HandleSummary(ctx *fasthttp.RequestCtx) {
	from := string(ctx.QueryArgs().Peek("from"))
	to := string(ctx.QueryArgs().Peek("to"))

	if from == "" || to == "" {
		ctx.Error("Missing from/to parameters", fasthttp.StatusBadRequest)
		return
	}

	summary, err := database.GetPaymentsSummary(from, to)
	if err != nil {
		ctx.Error("Database error", fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetContentType("application/json")
	if err := json.NewEncoder(ctx).Encode(summary); err != nil {
		ctx.Error("Failed to encode response", fasthttp.StatusInternalServerError)
	}
}

// Processa requisições de pagamento
func (h *APIHandlers) HandlePayment(ctx *fasthttp.RequestCtx) {
	requestTimestamp := time.Now()

	var payment models.Payment

	if err := json.Unmarshal(ctx.PostBody(), &payment); err != nil {
		ctx.Error("Invalid JSON", fasthttp.StatusBadRequest)
		return
	}

	if payment.CorrelationID == "" {
		payment.CorrelationID = generateUUID()
	}

	payment.RequestedAt = requestTimestamp

	if err := h.worker.EnqueuePayment(payment); err != nil {
		ctx.Error("Worker unavailable", fasthttp.StatusServiceUnavailable)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusAccepted)
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (h *APIHandlers) Router(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/payments":
		h.HandlePayment(ctx)
	case "/payments-summary":
		HandleSummary(ctx)
	case "/healthcheck":
		HandleHealthcheck(ctx)
	case "/purge-payments":
		HandlePurgePayments(ctx)
	default:
		ctx.Error("Not Found", fasthttp.StatusNotFound)
	}
}
