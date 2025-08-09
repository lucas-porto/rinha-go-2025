package main

import (
	"fmt"
	"os"
	"rinha-go-2025/internal/worker"
	"runtime/debug"
)

func main() {
	debug.SetGCPercent(-1)
	debug.SetMemoryLimit(80 * 1024 * 1024) // 80MB limite

	fmt.Println("Iniciando worker...")

	w := worker.New()

	// Testar conexão com banco
	if err := w.TestDatabaseConnection(); err != nil {
		fmt.Printf("Erro de conexão com banco: %v\n", err)
		os.Exit(1)
	}

	// Iniciar worker
	if err := w.Start(); err != nil {
		fmt.Printf("Erro ao iniciar worker: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Worker iniciado com sucesso!")

	// Manter worker rodando
	select {}
}
