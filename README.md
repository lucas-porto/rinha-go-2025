# Rinha de Backend 2025 - Go Ultra-Otimizado

Sistema de pagamentos de alta performance para a Rinha de Backend 2025, implementado em Go com otimizações extremas para máxima velocidade e confiabilidade. **P99 reduzido em 89.8% e inconsistências em 99.2%!**

## **🏆 Resultados de Performance**

### **Métricas Finais**
- **P99**: 38.13ms (redução de 89.8%!)
- **Inconsistências**: R$ 1.930 (redução de 99.2%!)
- **Total Líquido**: R$ 151.557 (melhor resultado)
- **Throughput**: 15.194 req/s (+8.2%)
- **Lag**: 2.621 (mínimo)

### **Evolução dos Resultados**
| **Métrica** | **Inicial** | **Otimização 1** | **Otimização 2** | **Melhoria Total** |
|-------------|-------------|------------------|------------------|-------------------|
| **P99** | 375.75ms | 407.92ms | **38.13ms** | **89.8% redução!** 🚀 |
| **Inconsistências** | R$ 232.193 | R$ 955 | R$ 1.930 | **99.2% redução!** 🚀 |
| **Total Líquido** | R$ 141.898 | R$ 136.881 | **R$ 151.557** | **+6.8%** ⬆️ |

## **🚀 Arquitetura Ultra-Otimizada**

### **Componentes**
- **API** (2 instâncias): FastHTTP + Unix Sockets + Worker Integrado
- **Worker Integrado**: Pool de 500 goroutines + Channel-based queuing
- **PostgreSQL**: Otimizado com batch operations
- **Nginx**: Load balancer com Unix Sockets
- **Connection Pooling**: HTTP clients otimizados

### **Tecnologias**
- **Go 1.21** com FastHTTP
- **PostgreSQL 15** com batch operations
- **Docker & Docker Compose**
- **Unix Sockets** para comunicação local
- **Connection Pooling** para HTTP clients
- **Retry Logic** com backoff exponencial

## **⚡ Otimizações Implementadas**

### **1. Connection Pooling (Principal Otimização)**
```go
// Default Processor
defaultClient: &fasthttp.Client{
    ReadTimeout:         300 * time.Millisecond,
    WriteTimeout:        300 * time.Millisecond,
    MaxConnsPerHost:     1000, // Pool de conexões
    MaxIdleConnDuration: 10 * time.Second,
}

// Fallback Processor
fallbackClient: &fasthttp.Client{
    ReadTimeout:         3 * time.Second,
    WriteTimeout:        3 * time.Second,
    MaxConnsPerHost:     100, // Pool menor
    MaxIdleConnDuration: 10 * time.Second,
}
```

### **2. Retry Logic com Backoff Exponencial**
```go
maxRetries := 2
baseDelay := 10 * time.Millisecond

// Backoff exponencial: 10ms, 20ms
delay := baseDelay * time.Duration(1<<attempt)
time.Sleep(delay)
```

### **3. Worker Integrado**
- **Pool Size**: 500 goroutines
- **Queue Buffer**: 100.000 items
- **Channel-based**: Comunicação in-memory
- **Sem Unix Sockets**: Eliminado overhead

### **4. Batch Database Operations**
```go
batchSize := 10 // Otimizado
flushInterval := 10 * time.Millisecond
retries := 1 // Reduzido para performance
```

### **5. Timestamp Correto**
- **RequestedAt**: Capturado no momento da solicitação
- **Não processamento**: Evita inconsistências
- **Timezone**: UTC para consistência

### **6. PostgreSQL Otimizado**
```sql
shared_buffers=64MB
effective_cache_size=256MB
work_mem=4MB
effective_io_concurrency=200
checkpoint_completion_target=0.9
```

## **🔧 Quick Start**

### **Pré-requisitos**
- Docker Desktop
- k6 (para testes de performance)

### **1. Clone e Setup**
```bash
git clone <repository>
cd rinha-go-2025
```

### **2. Build e Deploy**
```bash
cd build
docker-compose build --no-cache
docker-compose up -d
```

### **3. Verificar Status**
```bash
docker-compose ps
```

### **4. Executar Testes**
```bash
cd ../rinha-test
k6 run rinha.js
```

## **📁 Estrutura do Projeto**

```
rinha-go-2025/
├── build/
│   ├── docker-compose.yml      # Multi-container setup
│   ├── Dockerfile              # Multi-stage build
│   ├── nginx.conf              # Load balancer config
│   └── entrypoint.sh           # Startup script
├── cmd/
│   └── api/main.go             # Entry point com Unix Sockets
├── internal/
│   ├── database/
│   │   ├── connection.go       # Connection pooling
│   │   ├── operations.go       # Batch operations
│   │   └── schema.sql          # Database schema
│   ├── handlers/
│   │   └── api_handlers.go     # HTTP handlers
│   ├── models/
│   │   └── payment.go          # Data models
│   ├── processor/
│   │   └── processor.go        # Payment processor + connection pooling
│   └── worker/
│       └── worker.go           # Integrated worker
├── rinha-test/
│   ├── rinha.js                # k6 test script
│   └── partial-results.json    # Test results
└── README.md
```

## **⚙️ Configurações de Performance**

### **Docker Compose**
```yaml
services:
  api-1:
    build: .
    environment:
      - API_PORT=8080
      - SOCKET_NAME=api_1.sock
      - DEFAULT_PROCESSOR_URL=http://host.docker.internal:8001
      - FALLBACK_PROCESSOR_URL=http://host.docker.internal:8002
    volumes:
      - /var/run:/var/run  # Unix Sockets
    user: "0:0"            # Permissions
    ports:
      - "9999:8080"        # k6 compatibility
```

### **Nginx Config**
```nginx
upstream api {
    server unix:/var/run/api_1.sock;
    server unix:/var/run/api_2.sock;
    keepalive 64;
}
```

## **🎯 Fallback Strategy**

### **Payment Processors**
1. **Default**: Processador principal (300ms timeout)
2. **Fallback**: Processador secundário (3s timeout)

### **Retry Strategy**
- **Max Retries**: 2 por processor
- **Backoff**: Exponencial (10ms, 20ms)
- **Health Check**: A cada 120s
- **Circuit Breaker**: Removido para performance

## **📊 Endpoints**

### **API**
- `POST /payments` - Criar pagamento
- `GET /payments-summary?from=X&to=Y` - Resumo de pagamentos
- `GET /healthcheck` - Health check
- `POST /purge-payments` - Limpar pagamentos (teste)

### **Payment Processors**
- `POST /payments` - Processar pagamento
- `GET /payments/service-health` - Health check

## **🔍 Monitoramento**

### **Logs em Tempo Real**
```bash
docker-compose logs -f api-1
docker-compose logs -f api-2
```

### **Health Checks**
```bash
curl http://localhost/healthcheck
curl http://localhost:8001/payments/service-health
curl http://localhost:8002/payments/service-health
```

### **Database Status**
```bash
docker exec -it rinha-postgres psql -U postgres -d rinha -c "SELECT COUNT(*) FROM payments;"
```

## **🚀 Principais Conquistas**

### **Performance**
- **P99**: 38.13ms (89.8% redução)
- **Throughput**: 15.194 req/s (+8.2%)
- **Inconsistências**: R$ 1.930 (99.2% redução)

### **Arquitetura**
- **Worker Integrado**: Eliminou overhead de comunicação
- **Connection Pooling**: Reutilização de conexões HTTP
- **Retry Otimizado**: Backoff exponencial inteligente
- **Batch Operations**: Reduziu I/O do banco

### **Confiabilidade**
- **Fallback Funcional**: Processador secundário ativo
- **Timestamp Correto**: Eliminou inconsistências
- **Health Checks**: Monitoramento contínuo

## **📈 Análise de Resultados**

### **Antes das Otimizações**
- P99: 375.75ms
- Inconsistências: R$ 232.193
- Lag: 2.657
- Throughput: 14.042 req/s

### **Após Otimizações**
- P99: 38.13ms ⬇️
- Inconsistências: R$ 1.930 ⬇️
- Lag: 2.621 ⬇️
- Throughput: 15.194 req/s ⬆️

## **🤝 Contribuição**

1. Fork o projeto
2. Crie uma branch para sua feature
3. Commit suas mudanças
4. Push para a branch
5. Abra um Pull Request

## **📄 Licença**

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

---

**🏆 Desenvolvido para a Rinha de Backend 2025 - Resultados Excepcionais Alcançados!**