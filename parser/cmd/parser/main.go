package main

import (
	"bufio"
	"fmt"
	"parser/configs"
	"parser/internal/inmemory_cache"
	"parser/internal/parsers_manager"
	"parser/internal/parsers_status_manager"
	"runtime"

	"parser/internal/parser"

	"os"
	"strings"
)

func main() {

	// Получить количество CPU (то же, что runtime.NumCPU())
	currentMaxProcs := runtime.GOMAXPROCS(-1)
	fmt.Printf("Текущее значение GOMAXPROCS: %d\n", currentMaxProcs)

	fmt.Println("🚀 Multi-Source Vacancy Parser запущен!")
	fmt.Println("==========================")

	// создаём config
	conf, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	//создаём экземпляр inmemory cache для результатов поиска вакансий
	searchCache := inmemory_cache.NewInmemoryShardedCache(conf.Cache.NumOfShards, conf.Cache.SearchCacheConfig.SearchCacheCleanUp)

	//создаём экземпляр inmemory cache для обратного индекса для вакансий
	vacancyIndex := inmemory_cache.NewInmemoryShardedCache(conf.Cache.NumOfShards, conf.Cache.VacancyCacheConfig.VacancyCacheCleanUp)

	//создаём фабрику парсеров
	ParserFactory := parser.NewParserFactory()

	// регистрируем парсеры в фабрике
	// НЕ ВЫЗЫВАЕМ функцию, а передаем ее как значение!
	ParserFactory.Register("hh", conf.Parsers.HH, parser.NewHHParser)
	ParserFactory.Register("superjob", conf.Parsers.SuperJob, parser.NewSJParser)

	// создаём список парсеров для создания (пока хард-код, но в будущем это будут переменные)
	enabledParsers := []parser.ParserType{"hh", "superjob"}

	// создаём только те парсеры, у которых в конфиге указано Enabled
	parsers, err := ParserFactory.CreateEnabled(enabledParsers)
	if err != nil {
		panic(err)
	}

	// создаём мэнеджера состояния парсеров и инициализируем начальными значениями
	parserStatusManager := parsers_status_manager.NewParserStatusManager(conf, parsers...)

	// Создаём менеджер парсеров
	parserManager, err := parsers_manager.NewParserManager(conf, currentMaxProcs, searchCache, vacancyIndex, parserStatusManager, parsers...)
	if err != nil {
		panic(err)
	}

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
			parserManager.Shutdown() // останавливает работу всех запущенных воркеров
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
