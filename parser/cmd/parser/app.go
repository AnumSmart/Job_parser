package main

import (
	"bufio"
	"fmt"
	"os"
	"parser/configs"
	"parser/internal/inmemory_cache"
	"parser/internal/parser"
	"parser/internal/parsers_manager"
	"parser/internal/parsers_status_manager"
	"runtime"
	"strings"
)

// App содержит все зависимости приложения
type App struct {
	config              *configs.Config
	searchCache         *inmemory_cache.InmemoryShardedCache
	vacancyIndex        *inmemory_cache.InmemoryShardedCache
	vacancyDetails      *inmemory_cache.InmemoryShardedCache
	parserFactory       *parser.ParserFactory
	parserStatusManager *parsers_status_manager.ParserStatusManager
	parserManager       *parsers_manager.ParsersManager
	scanner             *bufio.Scanner
}

// initApp инициализирует все зависимости приложения
func initApp() (*App, error) {
	// Получаем количество CPU
	currentMaxProcs := runtime.GOMAXPROCS(-1)
	fmt.Printf("Текущее значение GOMAXPROCS: %d\n", currentMaxProcs)

	// Получаем конфигурацию
	conf, err := configs.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	//создаём экземпляр inmemory cache для результатов поиска вакансий
	searchCache := inmemory_cache.NewInmemoryShardedCache(conf.Cache.NumOfShards, conf.Cache.SearchCacheConfig.SearchCacheCleanUp)

	//создаём экземпляр inmemory cache для обратного индекса для вакансий
	vacancyIndex := inmemory_cache.NewInmemoryShardedCache(conf.Cache.NumOfShards, conf.Cache.VacancyCacheConfig.VacancyCacheCleanUp)

	// создаём экземпляр inmemory cache для деталей конкретной вакансии (ключ: ID вакансии)
	vacancyDetails := inmemory_cache.NewInmemoryShardedCache(conf.Cache.NumOfShards, conf.Cache.VacancyCacheConfig.VacancyCacheCleanUp)

	//создаём фабрику парсеров
	parserFactory := parser.NewParserFactory()

	// регистрируем парсеры в фабрике
	// НЕ ВЫЗЫВАЕМ функцию, а передаем ее как значение!
	parserFactory.Register("hh", conf.Parsers.HH, parser.NewHHParser)
	parserFactory.Register("superjob", conf.Parsers.SuperJob, parser.NewSJParser)

	// создаём список парсеров для создания (пока хард-код, но в будущем это будут переменные)
	enabledParsers := []parser.ParserType{"hh", "superjob"}

	// создаём только те парсеры, у которых в конфиге указано Enabled
	parsers, err := parserFactory.CreateEnabled(enabledParsers)
	if err != nil {
		panic(err)
	}

	// создаём мэнеджера состояния парсеров и инициализируем начальными значениями
	parserStatusManager := parsers_status_manager.NewParserStatusManager(conf, parsers...)

	// Создаём менеджер парсеров
	parserManager, err := parsers_manager.NewParserManager(conf, currentMaxProcs, searchCache, vacancyIndex, vacancyDetails, parserStatusManager, parsers...)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser manager: %w", err)
	}

	// Создаем сканер для ввода
	scanner := bufio.NewScanner(os.Stdin)

	// воздвращаем экземпляр приложения
	return &App{
		config:              conf,
		searchCache:         searchCache,
		vacancyIndex:        vacancyIndex,
		vacancyDetails:      vacancyDetails,
		parserFactory:       parserFactory,
		parserStatusManager: parserStatusManager,
		parserManager:       parserManager,
		scanner:             scanner,
	}, nil
}

func (a *App) Run() error {
	fmt.Println("🚀 Multi-Source Vacancy Parser запущен!")
	fmt.Println("==========================")

	// Основной цикл приложения
	for {
		printMenu()
		fmt.Print("Выберите действие: ")

		if !a.scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(a.scanner.Text())

		switch choice {
		case "1":
			err := a.parserManager.MultiSearch(a.scanner)
			if err != nil {
				fmt.Printf("Ошибка поиска: %v\n", err)
				continue
			}
		case "2":
			err := a.parserManager.GetVacancyDetails(a.scanner)
			if err != nil {
				fmt.Printf("Ошибка получения деталей: %v\n", err)
				continue
			}
		case "3":
			err := a.parserManager.GetFullVacancyDetails(a.scanner)
			if err != nil {
				fmt.Printf("Ошибка получения полных деталей: %v\n", err)
				continue
			}
		case "4":
			a.parserManager.Shutdown()
			fmt.Println("👋 До свидания!")
			return nil
		default:
			fmt.Println("❌ Неверный выбор. Попробуйте снова.")
		}

		fmt.Println()
	}

	return nil
}

// вспомогательная функция
func printMenu() {
	fmt.Println("📋 Меню:")
	fmt.Println("1. Поиск вакансий (расширенный)")
	fmt.Println("2. Получить описание вакансии по ID ")
	fmt.Println("3. Получить полное описание вакансии по ID ")
	fmt.Println("4. Выход")
}
