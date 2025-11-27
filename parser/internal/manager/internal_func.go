package manager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"parser/internal/domain/models"
	"parser/internal/interfaces"
	"sync"
	"time"
)

// concurrentSearchWithTimeout выполняет поиск во всех парсерах одновременно с таймаутом
func (pm *ParserManager) concurrentSearchWithTimeout(ctx context.Context, searchHash string, params models.SearchParams, timeout time.Duration) ([]models.SearchResult, error) {
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
					SearchHash: searchHash,
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
func genHashFromSearchParam(params models.SearchParams) (string, error) {
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
