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
	numOfShards     = 7
	searchCacheTTL  = 10 * time.Minute
	vacancyCacheTTL = 60 * time.Minute
)

func main() {
	fmt.Println("🚀 Multi-Source Vacancy Parser запущен!")
	fmt.Println("==========================")

	// создаём config
	conf, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	//создаём экземпляр inmemory cache для результатов поиска вакансий
	searchCache := inmemory_cache.NewInmemoryShardedCache(numOfShards, searchCacheTTL)

	//создаём экземпляр inmemory cache для обратного индекса для вакансий
	vacancyIndex := inmemory_cache.NewInmemoryShardedCache(numOfShards, vacancyCacheTTL)

	// Создаём парсеры
	hhParser := parser.NewHHParser()
	sjParser := parser.NewSuperJobParser(conf.Api_conf.SJAPIKey)

	// Создаём менеджер парсеров
	parserManager := manager.NewParserManager(conf, searchCache, vacancyIndex, hhParser, sjParser)

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
			err := parserManager.GetVacancyDetails(scanner)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
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
