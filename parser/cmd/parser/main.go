package main

import (
	"bufio"
	"fmt"
	"parser/configs"
	"parser/internal/inmemory_cache"
	"parser/internal/manager"
	"parser/internal/parser"

	"os"
	"strings"
	"time"
)

const (
	numOfShards = 7
)

func main() {
	fmt.Println("🚀 Multi-Source Vacancy Parser запущен!")
	fmt.Println("==========================")

	// создаём config
	conf, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	//создаём экземпляр inmemory cache
	cacheSh := inmemory_cache.NewInmemoryShardedCache(numOfShards, 10*time.Minute)

	// Создаём парсеры
	hhParser := parser.NewHHParser()
	sjParser := parser.NewSuperJobParser(conf.Api_conf.SJAPIKey)

	// Создаём менеджер парсеров
	parserManager := manager.NewParserManager(conf, cacheSh, hhParser, sjParser)

	// Основной цикл приложения
	scanner := bufio.NewScanner(os.Stdin)

	for {
		printMenu()
		fmt.Print("Выберите действие: ")

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			parserManager.MultiSearch(scanner)
		case "2":
			parserManager.GetVacancyDetails(scanner)
		case "3":
			fmt.Println("👋 До свидания!")
			return
		default:
			fmt.Println("❌ Неверный выбор. Попробуйте снова.")
		}

		fmt.Println()
	}
}

func printMenu() {
	fmt.Println("📋 Меню:")
	fmt.Println("1. Поиск вакансий (расширенный)")
	fmt.Println("2. Получить детали вакансии по ID ")
	fmt.Println("3. Выход")
}
