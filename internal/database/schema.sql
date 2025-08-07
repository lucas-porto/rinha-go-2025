-- Schema para Rinha de Backend 2025

ALTER DATABASE rinha SET TIME ZONE 'UTC';
ALTER SYSTEM SET fsync = 'off';
ALTER SYSTEM SET synchronous_commit = 'off';
ALTER SYSTEM SET full_page_writes = 'off';
ALTER SYSTEM SET logging_collector = 'off';
ALTER SYSTEM SET log_statement = 'none';
ALTER SYSTEM SET max_connections = 200;

-- Tabela de pagamentos
CREATE UNLOGGED TABLE IF NOT EXISTS payments (
    correlation_id UUID PRIMARY KEY,
    amount DECIMAL NOT NULL,
    handler VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Tabela de cache
CREATE UNLOGGED TABLE IF NOT EXISTS cache (
    processor VARCHAR(20) PRIMARY KEY,
    failing BOOLEAN NOT NULL,
    min_response_time INTEGER NOT NULL,
    last_update TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Inserir dados iniciais do cache
INSERT INTO cache (processor, failing, min_response_time, last_update) 
VALUES 
    ('default', false, 0, NOW()),
    ('fallback', false, 0, NOW())
ON CONFLICT (processor) DO NOTHING;

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments (created_at);
CREATE INDEX IF NOT EXISTS idx_payments_handler ON payments (handler);
CREATE INDEX IF NOT EXISTS idx_payments_handler_created_at ON payments (handler, created_at); 