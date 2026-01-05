package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"parser/internal/core"
	"parser/internal/server"
	"syscall"
	"time"
)

func main() {
	// Инициализируем общие зависимости
	deps, err := core.InitDependencies()
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	// Создаем HTTP-сервер
	server, err := server.NewServer(context.Background(), deps.Config.Server)
	if err != nil {
		panic("Failed to create server!")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера
	go func() {
		fmt.Printf("🚀 HTTP сервер запускается на %s\n", deps.Config.Server.Addr())
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидание сигнала
	<-sigChan
	fmt.Println("\n🛑 Остановка сервера...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	/*
		// Остановка сервисов
		vacancyService.Shutdown()
	*/

	// Остановка сервера
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	fmt.Println("👋 Сервер остановлен")
}
