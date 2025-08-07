package main

import (
	"fmt"
	"net"
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

	// fmt.Printf("API iniciada na porta %s\n", port)

	// fmt.Println("Criando worker...")
	w := worker.New()

	// fmt.Println("Testando conexão com banco...")
	if err := w.TestDatabaseConnection(); err != nil {
		fmt.Printf("Erro de conexão com banco: %v\n", err)
		os.Exit(1)
	}

	// fmt.Println("Iniciando worker...")
	if err := w.Start(); err != nil {
		fmt.Printf("Erro ao iniciar worker: %v\n", err)
		os.Exit(1)
	}

	// fmt.Println("Worker iniciado com sucesso!")

	// Criar handlers da API com worker integrado
	// fmt.Println("Criando handlers da API...")
	apiHandlers := handlers.NewAPIHandlersWithWorker(w)

	// Criar Unix Socket para Nginx
	socketName := os.Getenv("SOCKET_NAME")

	socketPath := fmt.Sprintf("/var/run/%s", socketName)

	os.Remove(socketPath)

	// Criar listener Unix Socket
	unixListener, err := net.Listen("unix", socketPath)
	if err != nil {
		// fmt.Printf("Erro ao criar Unix Socket %s: %v\n", socketPath, err)
		// Continuar apenas com HTTP
		// fmt.Printf("Iniciando servidor HTTP na porta %s...\n", port)
		if err := fasthttp.ListenAndServe(":"+port, apiHandlers.Router); err != nil {
			panic(fmt.Sprintf("Erro ao iniciar servidor: %v", err))
		}
		return
	}

	// Configurar permissões do socket
	if err := os.Chmod(socketPath, 0666); err != nil {
		// fmt.Printf("Erro ao configurar permissões do socket: %v\n", err)
	}

	// fmt.Printf("Unix Socket criado: %s\n", socketPath)
	// fmt.Printf("Iniciando servidor HTTP na porta %s e Unix Socket...\n", port)

	// Iniciar servidor HTTP em goroutine separada
	go func() {
		if err := fasthttp.ListenAndServe(":"+port, apiHandlers.Router); err != nil {
			fmt.Printf("Erro no servidor HTTP: %v\n", err)
		}
	}()

	// Iniciar servidor Unix Socket
	if err := fasthttp.Serve(unixListener, apiHandlers.Router); err != nil {
		panic(fmt.Sprintf("Erro ao iniciar servidor Unix Socket: %v", err))
	}
}
