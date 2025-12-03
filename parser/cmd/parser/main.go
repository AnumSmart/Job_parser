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
)

func main() {
	fmt.Println("🚀 Multi-Source Vacancy Parser запущен!")
	fmt.Println("==========================")

	// создаём config
	conf, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	// создаём конфиг для парсеров
	parsConf, err := (configs.LoadParseConfig(conf.ParsConfAddress))
	if err != nil {
		panic(err)
	}

	//создаём экземпляр inmemory cache для результатов поиска вакансий
	searchCache := inmemory_cache.NewInmemoryShardedCache(conf.Cache_conf.NumOfShards, conf.Cache_conf.SearchCacheTTL)

	//создаём экземпляр inmemory cache для обратного индекса для вакансий
	vacancyIndex := inmemory_cache.NewInmemoryShardedCache(conf.Cache_conf.NumOfShards, conf.Cache_conf.VacancyCacheTTL)

	// Создаём парсеры
	hhParser := parser.NewHHParser(parsConf.HH)
	sjParser := parser.NewSJParser(parsConf.SuperJob)

	/*
		//создаём фабрику парсеров
		ParserFactory := parser.NewParserFactory()

		// регистрируем парсеры в фабрике
		// НЕ ВЫЗЫВАЕМ функцию, а передаем ее как значение!
		ParserFactory.Register("hh", parsConf.HH, parser.NewHHParser)
		ParserFactory.Register("superjob", parsConf.SuperJob, parser.NewSJParser)

		enabledParsers := []parser.ParserType{"hh", "superjob"}

		parsers, err := ParserFactory.CreateEnabled(enabledParsers) // доработать использование фабрики-----------------------------
		if err != nil {
			panic(err)
		}

	*/
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
			err := parserManager.MultiSearch(scanner)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
		case "2":
			err := parserManager.GetVacancyDetails(scanner)
			if err != nil {
				fmt.Println(err.Error())
				continue
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
