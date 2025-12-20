package main

import (
	"fmt"
)

func main() {

	// инициализируем приложение
	app, err := initApp()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize app: %v", err))
	}

	// запускаем приложение
	if err := app.Run(); err != nil {
		panic(fmt.Sprintf("App runtime error: %v", err))
	}
}
