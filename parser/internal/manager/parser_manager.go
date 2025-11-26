package manager

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"job_parser/configs"
	"job_parser/internal/domain/models"
	"job_parser/internal/inmemory_cache"
	"job_parser/internal/interfaces"

	"strconv"
	"strings"

	"sync"
	"time"
)

type ParserManager struct {
	parsers []interfaces.Parser
	config  *configs.Config
}

// Конструктор для мэнеджера парсинга из разных источников
func NewParserManager(config *configs.Config, parsers ...interfaces.Parser) *ParserManager {
	return &ParserManager{
		parsers: parsers,
		config:  config,
	}
}

// Метод для мульти-поиска
func (pm *ParserManager) MultiSearch(scanner *bufio.Scanner, cash *inmemory_cache.InmemoryShardedCache) {
	fmt.Println("\n🌐 Мульти-поиск вакансий")

	var params models.SearchParams

	fmt.Print("Введите поисковый запрос: ")
	if scanner.Scan() {
		params.Text = strings.TrimSpace(scanner.Text())
	}

	fmt.Print("Количество вакансий на источник (max 50): ")
	if scanner.Scan() {
		countStr := strings.TrimSpace(scanner.Text())
		if countStr != "" {
			if count, err := strconv.Atoi(countStr); err == nil && count > 0 {
				params.PerPage = count
			}
		}
	}

	if params.PerPage == 0 {
		params.PerPage = 20
	}

	searchHash, _ := GenHashFromSearchParam(params) // ****!!!! нужно обработать ошибку
	fmt.Println("Сформированный хэш:", searchHash)

	fmt.Println("⏳ Ищем вакансии в кэше...")
	searchRes, ok := cash.GetItem(searchHash)
	if ok {
		pm.printMultiSearchResults(searchRes, params.PerPage)
		return
	}

	fmt.Println("Не удалось найти данные в кэше ⏳ Ищем вакансии во всех источниках...")
	//fmt.Println("⏳ Ищем вакансии во всех источниках...")
	ctx := context.Background()
	results, err := pm.concurrentSearchWithTimeout(ctx, params, 10*time.Second)
	if err != nil {
		fmt.Printf("❌ Ошибка при поиске: %v\n", err)
		return
	}
	//записываем данные в кэш
	cash.AddItemWithTTL(searchHash, results, time.Minute)

	pm.printMultiSearchResults(results, params.PerPage)
}

// concurrentSearchWithTimeout выполняет поиск во всех парсерах одновременно с таймаутом
func (pm *ParserManager) concurrentSearchWithTimeout(ctx context.Context, params models.SearchParams, timeout time.Duration) ([]models.SearchResult, error) {
	// создаём контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan models.SearchResult, len(pm.parsers))

	for _, parser := range pm.parsers {
		wg.Add(1)
		go func(p interfaces.Parser) {
			defer wg.Done()

			// Создаем канал для результата и создаём ещё одну горутину, где производим поиск
			// 2я - горутина нужна, чтобы потом использовать select и контролировать отмену контекста
			resultChan := make(chan models.SearchResult, 1)

			go func() {
				start := time.Now()
				vacancies, err := p.SearchVacancies(params)
				duration := time.Since(start)

				resultChan <- models.SearchResult{
					ParserName: p.GetName(),
					Vacancies:  vacancies,
					Error:      err,
					Duration:   duration,
				}
			}()

			select {
			case <-ctx.Done():
				// Таймаут или отмена
				results <- models.SearchResult{
					ParserName: p.GetName(),
					Error:      fmt.Errorf("timeout exceeded"),
				}
			case result := <-resultChan:
				results <- result
			}
		}(parser)
	}

	// в этой горутине дожидаемся окончания обработки от всех парсеров и закрываем канал результатов
	go func() {
		wg.Wait()
		close(results)
	}()

	// обьявляем переменную для выходных данных
	var searchResults []models.SearchResult

	for result := range results {
		searchResults = append(searchResults, result)
	}

	return searchResults, nil
}

// GetAllParsers возвращает список доступных парсеров
func (pm *ParserManager) GetParserNames() []string {
	names := make([]string, len(pm.parsers))
	for i, parser := range pm.parsers {
		names[i] = parser.GetName()
	}
	return names
}

// Метод для вывода в консоль результатов поиска (с нужными атрибутами)
func (pm *ParserManager) printMultiSearchResults(results []models.SearchResult, resultsPerPage int) {
	totalVacancies := 0

	for _, result := range results {
		fmt.Printf("\n📊 %s:\n", result.ParserName)
		fmt.Printf("   ⏱️  Время: %v\n", result.Duration)

		if result.Error != nil {
			fmt.Printf("   ❌ Ошибка: %v\n", result.Error)
			continue
		}

		fmt.Printf("   ✅ Найдено: %d вакансий\n", len(result.Vacancies))
		totalVacancies += len(result.Vacancies)

		// Показываем первые 3 вакансии из каждого источника
		for i, vacancy := range result.Vacancies {
			if i >= resultsPerPage {
				break
			}
			fmt.Printf("      %d. %s - %s, company:%s, URL:[ %s ], ID:%s\n", i+1, vacancy.Job, *vacancy.Salary, vacancy.Company, vacancy.URL, vacancy.ID)
		}

		if len(result.Vacancies) > resultsPerPage {
			fmt.Printf("      ... и ещё %d\n", len(result.Vacancies)-resultsPerPage)
		}
	}

	fmt.Printf("\n🎯 Всего найдено: %d вакансий\n", totalVacancies)
}

// функция генерирует хэш запроса поиска, чтобы кэшировать запросы по этому хэшу
func GenHashFromSearchParam(params models.SearchParams) (string, error) {
	// Учитываем ВСЕ параметры, которые влияют на результат
	keyData := struct {
		Text    string `json:"text"`
		Area    string `json:"area"`
		PerPage int    `json:"per_page"`
		Page    int    `json:"page"`
		// Добавьте другие поля из SearchParams
	}{
		Text:    params.Text,
		Area:    params.Area,
		PerPage: params.PerPage,
		Page:    params.Page,
	}

	data, err := json.Marshal(keyData)
	if err != nil {
		return "", fmt.Errorf("Error while marshaling of params: %w\n", err)
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%s", hex.EncodeToString(hash[:16])), nil
}
