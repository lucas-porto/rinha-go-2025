# Rinha de Backend 2025 - Go

Sistema criado em Go para participação da rinha de backend com intuito de aprender e me aventurar 
nesta área.

Se quiserem ver a implementação aqui está o [repo](https://github.com/lucas-porto/rinha-go-2025).

## Tecnologias Utilizadas

### **Backend**
- **Go 1.21** com FastHTTP
- **Worker separado** com pool de 1000 goroutines
- **Connection pooling** para HTTP clients
- **Retry logic** com backoff exponencial
- **Batch operations** para database

### **Infraestrutura**
- **PostgreSQL 15** otimizado
- **HAProxy 2.8** com load balancing inteligente
- **Docker & Docker Compose**
- **Container separado** para worker

## Otimizações Implementadas

### **1. HAProxy Load Balancer**
- Balanceamento `leastconn` para distribuição inteligente
- Health checks configurados (status 204)
- Reutilização agressiva de conexões
- Timeouts otimizados

### **2. Worker Separado**
- Container dedicado para processamento
- Pool de 1000 goroutines
- Buffer de 200.000 items
- Isolamento completo de recursos

### **3. Connection Pooling**
- Reutilização de conexões HTTP
- Timeouts otimizados (300ms default, 3s fallback)
- Pool de 1000 conexões para default, 100 para fallback

### **4. Batch Database Operations**
- Agrupamento por handler (default/fallback)
- Batch size de 5 registros
- Flush automático a cada 5ms

### **5. Retry Logic**
- Backoff exponencial (10ms, 20ms)
- Max 2 retries por processor
- Health checks a cada 120s

## Arquitetura

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    k6       │    │   HAProxy   │    │   API-1     │
│   (Load)    │───▶│ (Load Bal)  │───▶│ (FastHTTP)  │
└─────────────┘    └─────────────┘    └─────────────┘
                          │                   │
                          ▼                   ▼
                   ┌─────────────┐    ┌─────────────┐
                   │   API-2     │    │ PostgreSQL  │
                   │ (FastHTTP)  │    │ (Batch)     │
                   └─────────────┘    └─────────────┘
                            │                   ▲
                            └───────────────────┘
                                    │
                                    ▼
                            ┌─────────────┐
                            │   Worker    │
                            │ (Separado)  │
                            └─────────────┘
```

**Fluxo de Requisições:**
1. **k6** envia requisições para **HAProxy** (porta 9999)
2. **HAProxy** faz load balancing entre **API-1** e **API-2**
3. **Cada API** envia pagamentos para o **Worker** via HTTP
4. **Worker** processa e salva no **PostgreSQL** em batch
5. **Ambas as APIs** se conectam ao mesmo banco de dados

## Configurações de Performance

### **Docker Compose**
- 2 instâncias da API
- Worker em container separado
- PostgreSQL otimizado
- HAProxy com load balancing

### **Otimizações PostgreSQL**
```sql
shared_buffers=64MB
effective_cache_size=256MB
work_mem=4MB
effective_io_concurrency=200
checkpoint_completion_target=0.9
```

## Estratégia de Fallback

- **Default**: Processador principal (300ms timeout)
- **Fallback**: Processador secundário (3s timeout)
- **Retry**: 2 tentativas com backoff exponencial
- **Health Check**: Monitoramento contínuo

## Endpoints

- `POST /payments` - Criar pagamento
- `GET /payments-summary?from=X&to=Y` - Resumo de pagamentos
- `GET /healthcheck` - Health check (status 204)
- `POST /purge-payments` - Limpar pagamentos (teste)

