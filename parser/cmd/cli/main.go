package main

import (
	"fmt"
	"log"
	"parser/internal/cli"
	"parser/internal/core"
)

func main() {

	// Инициализируем зависимости
	deps, err := core.InitDependencies()
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	// инициализируем консольную реализацию логики
	cliApp := cli.NewCLIApp(deps.Config, deps.ParserManager)

	// запускаем консольное приложение
	if err := cliApp.Run(); err != nil {
		panic(fmt.Sprintf("App runtime error: %v", err))
	}
}
