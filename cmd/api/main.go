package main

import (
	"fmt"
	"os"
	"rinha-go-2025/internal/handlers"
	"rinha-go-2025/internal/worker"
	"runtime/debug"

	"github.com/valyala/fasthttp"
)

func main() {
	debug.SetGCPercent(-1)
	debug.SetMemoryLimit(100 * 1024 * 1024) // 100MB limite

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("API iniciada na porta %s\n", port)

	fmt.Println("Criando worker...")
	w := worker.New()

	fmt.Println("Testando conexão com banco...")
	if err := w.TestDatabaseConnection(); err != nil {
		fmt.Printf("Erro de conexão com banco: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Iniciando worker...")
	if err := w.Start(); err != nil {
		fmt.Printf("Erro ao iniciar worker: %v\n", err)
		os.Exit(1)
	}

	// fmt.Println("Worker iniciado com sucesso!")

	// Criar handlers da API com worker integrado
	// fmt.Println("Criando handlers da API...")
	apiHandlers := handlers.NewAPIHandlersWithWorker(w)

	fmt.Printf("Iniciando servidor HTTP na porta %s...\n", port)
	if err := fasthttp.ListenAndServe(":"+port, apiHandlers.Router); err != nil {
		panic(fmt.Sprintf("Erro ao iniciar servidor: %v", err))
	}
}
