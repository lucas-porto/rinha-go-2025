# Rinha de Backend 2025 - Go

Sistema criado em Go para participação da rinha de backend com intuito de aprender e me aventurar 
nesta área.

## Tecnologias Utilizadas

### **Backend**
- **Go 1.21** com FastHTTP
- **Worker integrado** com pool de 500 goroutines
- **Connection pooling** para HTTP clients
- **Retry logic** com backoff exponencial
- **Batch operations** para database

### **Infraestrutura**
- **PostgreSQL 15**
- **Nginx** com Unix Sockets
- **Docker & Docker Compose**
- **Unix Sockets** para comunicação local

## Otimizações Implementadas

### **1. Connection Pooling**
- Reutilização de conexões HTTP
- Timeouts otimizados (300ms default, 3s fallback)
- Pool de 1000 conexões para default, 100 para fallback

### **2. Worker Integrado**
- Pool de 500 goroutines
- Buffer de 100.000 items
- Comunicação in-memory via channels

### **3. Batch Database Operations**
- Agrupamento por handler (default/fallback)
- Batch size de 10 registros
- Flush automático a cada 10ms

### **4. Retry Logic**
- Backoff exponencial (10ms, 20ms)
- Max 2 retries por processor
- Health checks a cada 120s

### **5. Unix Sockets**
- Comunicação local ultra-rápida
- Elimina overhead de HTTP/TCP
- Nginx configurado para Unix Sockets

## Arquitetura

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    k6       │    │   Nginx     │    │   API-1     │
│   (Load)    │───▶│ (Unix Sock) │───▶│ (Worker)    │
└─────────────┘    └─────────────┘    └─────────────┘
                          │                   │
                          ▼                   ▼
                   ┌─────────────┐    ┌─────────────┐
                   │   API-2     │    │ PostgreSQL  │
                   │ (Worker)    │    │ (Batch)     │
                   └─────────────┘    └─────────────┘
                            │                   ▲
                            └───────────────────┘
```

**Fluxo de Requisições:**
1. **k6** envia requisições para **Nginx** (porta 80)
2. **Nginx** faz load balancing entre **API-1** e **API-2** via Unix Sockets
3. **Cada API** processa independentemente e salva no **PostgreSQL**
4. **Ambas as APIs** se conectam ao mesmo banco de dados

## Configurações de Performance

### **Docker Compose**
- 2 instâncias da API
- PostgreSQL otimizado
- Nginx com Unix Sockets
- Volumes para persistência

### **Otimizações PostgreSQL**
```sql
shared_buffers=64MB
effective_cache_size=256MB
work_mem=4MB
effective_io_concurrency=200
```

## Estratégia de Fallback

- **Default**: Processador principal (300ms timeout)
- **Fallback**: Processador secundário (3s timeout)
- **Retry**: 2 tentativas com backoff exponencial
- **Health Check**: Monitoramento contínuo

