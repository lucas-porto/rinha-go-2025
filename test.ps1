# Script de teste para Windows PowerShell
Write-Host "🧪 Testando Rinha Backend API..." -ForegroundColor Green

# Aguardar serviços estarem prontos
Write-Host "⏳ Aguardando serviços estarem prontos..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Teste 1: Health Check
Write-Host "📋 Teste 1: Health Check" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost/healthcheck" -Method GET -UseBasicParsing
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "Erro: $($_.Exception.Message)" -ForegroundColor Red
}

# Teste 2: Enviar Pagamento
Write-Host "📋 Teste 2: Enviar Pagamento" -ForegroundColor Cyan
$payment1 = @{
    correlationId = "123e4567-e89b-12d3-a456-426614174000"
    amount = 100.50
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest -Uri "http://localhost/payments" -Method POST -Body $payment1 -ContentType "application/json" -UseBasicParsing
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "Erro: $($_.Exception.Message)" -ForegroundColor Red
}

# Teste 3: Enviar Outro Pagamento
Write-Host "📋 Teste 3: Enviar Outro Pagamento" -ForegroundColor Cyan
$payment2 = @{
    correlationId = "987fcdeb-51a2-43d1-b789-123456789abc"
    amount = 250.75
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest -Uri "http://localhost/payments" -Method POST -Body $payment2 -ContentType "application/json" -UseBasicParsing
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "Erro: $($_.Exception.Message)" -ForegroundColor Red
}

# Aguardar processamento
Write-Host "⏳ Aguardando processamento..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# Teste 4: Consultar Resumo
Write-Host "📋 Teste 4: Consultar Resumo" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost/payments-summary?from=2024-01-01T00:00:00Z&to=2024-12-31T23:59:59Z" -Method GET -UseBasicParsing
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Resposta: $($response.Content)" -ForegroundColor Gray
} catch {
    Write-Host "Erro: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "✅ Testes concluídos!" -ForegroundColor Green 