# Rinha de Backend 2025 - Go

Sistema de pagamentos de alta performance para a Rinha de Backend 2025, implementado em Go com arquitetura otimizada para velocidade e confiabilidade.

## Arquitetura

### Componentes Principais

- **API** (2 instâncias): Servidor HTTP usando FastHTTP
- **Worker Separado**: Container dedicado para processamento de pagamentos
- **HAProxy**: Load balancer para distribuição de carga
- **PostgreSQL**: Banco de dados otimizado para performance
- **Connection Pooling**: Clientes HTTP otimizados com reutilização de conexões

### Tecnologias Utilizadas

- **Go 1.21** com FastHTTP para alta performance
- **HAProxy 2.8** para load balancing inteligente
- **PostgreSQL 15** com otimizações de performance
- **Docker & Docker Compose** para containerização
- **Connection Pooling** para HTTP clients
- **Retry Logic** com backoff exponencial

## Otimizações Implementadas

### 1. HAProxy Load Balancer

Configuração otimizada com balanceamento `leastconn` e health checks:

```haproxy
backend api_backend
    mode http
    balance leastconn
    option httpchk GET /healthcheck
    http-check expect status 204
    http-reuse aggressive
    
    default-server check inter 2s rise 2 fall 3 maxconn 1000
    server api-1 api-1:8080 check
    server api-2 api-2:8080 check
```

### 2. Worker Separado

Container dedicado para processamento isolado de pagamentos:

```yaml
worker:
  build:
    context: ..
    dockerfile: build/Dockerfile
    args:
      TARGET: worker
  deploy:
    resources:
      limits:
        cpus: "0.3"
        memory: "80MB"
```

### 3. Connection Pooling Otimizado

Configuração de clientes HTTP com pools de conexão:

```go
// Default Processor
defaultClient: &fasthttp.Client{
    ReadTimeout:         300 * time.Millisecond,
    WriteTimeout:        300 * time.Millisecond,
    MaxConnsPerHost:     1000,
    MaxIdleConnDuration: 10 * time.Second,
}

// Fallback Processor
fallbackClient: &fasthttp.Client{
    ReadTimeout:         3 * time.Second,
    WriteTimeout:        3 * time.Second,
    MaxConnsPerHost:     100,
    MaxIdleConnDuration: 10 * time.Second,
}
```

### 4. Retry Logic com Backoff Exponencial

Sistema de retry inteligente para processadores de pagamento:

```go
maxRetries := 2
baseDelay := 10 * time.Millisecond

// Backoff exponencial: 10ms, 20ms
delay := baseDelay * time.Duration(1<<attempt)
time.Sleep(delay)
```

### 5. Worker Pool Otimizado

- Pool de 1000 goroutines para processamento paralelo
- Buffer de 200.000 items na fila
- Comunicação baseada em channels
- Container separado para isolamento de recursos

### 6. Batch Database Operations

Operações em lote para reduzir I/O do banco:

```go
batchSize := 5
flushInterval := 5 * time.Millisecond
retries := 1
```

### 7. Timestamp Correto

Captura do timestamp no momento da solicitação para evitar inconsistências:

```go
requestTimestamp := time.Now()
payment.RequestedAt = requestTimestamp
```

### 8. PostgreSQL Otimizado

Configurações de performance para o banco de dados:

```sql
shared_buffers=64MB
effective_cache_size=256MB
work_mem=4MB
effective_io_concurrency=200
checkpoint_completion_target=0.9
```

## Quick Start

### Pré-requisitos

- Docker Desktop
- k6 (para testes de performance)

### 1. Clone e Setup

```bash
git clone <repository>
cd rinha-go-2025
```

### 2. Build e Deploy

```bash
cd build
docker-compose build --no-cache
docker-compose up -d
```

### 3. Verificar Status

```bash
docker-compose ps
```

### 4. Executar Testes

```bash
cd ../rinha-test
k6 run rinha.js
```

## Estrutura do Projeto

```
rinha-go-2025/
├── build/
│   ├── docker-compose.yml      # Multi-container setup
│   ├── Dockerfile              # Multi-stage build
│   ├── haproxy.cfg             # HAProxy load balancer config
│   └── entrypoint.sh           # Startup script
├── cmd/
│   ├── api/main.go             # API entry point
│   └── worker/main.go          # Worker entry point
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
│       └── worker.go           # Worker pool
├── rinha-test/
│   ├── rinha.js                # k6 test script
│   └── partial-results.json    # Test results
└── README.md
```

## Configurações de Performance

### Docker Compose

```yaml
services:
  api-1:
    build: .
    environment:
      - API_PORT=8080
      - DEFAULT_PROCESSOR_URL=http://payment-processor-default:8080
      - FALLBACK_PROCESSOR_URL=http://payment-processor-fallback:8080
    deploy:
      resources:
        limits:
          cpus: "0.3"
          memory: "70MB"

  worker:
    build: .
    args:
      TARGET: worker
    deploy:
      resources:
        limits:
          cpus: "0.3"
          memory: "80MB"

  haproxy:
    image: haproxy:2.8-alpine
    ports:
      - "9999:9999"
    volumes:
      - ./haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg
```

### HAProxy Config

```haproxy
global
    maxconn 20000
    nbthread 2
    tune.bufsize 32768

backend api_backend
    mode http
    balance leastconn
    option httpchk GET /healthcheck
    http-check expect status 204
    server api-1 api-1:8080 check
    server api-2 api-2:8080 check
```

## Estratégia de Fallback

### Processadores de Pagamento

1. **Default**: Processador principal com timeout de 300ms
2. **Fallback**: Processador secundário com timeout de 3s

### Estratégia de Retry

- Máximo de 2 tentativas por processador
- Backoff exponencial (10ms, 20ms)
- Health check a cada 120s
- Circuit breaker removido para otimizar performance

## Endpoints

### API

- `POST /payments` - Criar pagamento
- `GET /payments-summary?from=X&to=Y` - Resumo de pagamentos
- `GET /healthcheck` - Health check (status 204)
- `POST /purge-payments` - Limpar pagamentos (teste)

### HAProxy

- `GET /healthcheck` - Health check do load balancer
- `GET http://localhost:9797` - Health check interno

### Processadores de Pagamento

- `POST /payments` - Processar pagamento
- `GET /payments/service-health` - Health check

## Monitoramento

### Logs em Tempo Real

```bash
docker-compose logs -f api-1
docker-compose logs -f api-2
docker-compose logs -f worker
docker-compose logs -f haproxy
```

### Health Checks

```bash
curl http://localhost:9999/healthcheck
curl http://localhost:8080/healthcheck
curl http://localhost:8081/healthcheck
```

### Status do Banco de Dados

```bash
docker exec -it rinha-postgres psql -U postgres -d rinha -c "SELECT COUNT(*) FROM payments;"
```

## Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature
3. Commit suas mudanças
4. Push para a branch
5. Abra um Pull Request

## Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

---

Desenvolvido para a Rinha de Backend 2025 com arquitetura otimizada para alta performance e confiabilidade.