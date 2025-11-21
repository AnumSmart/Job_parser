package manager

import (
	"context"
	"fmt"
	"parser/internal/interfaces"
	"parser/internal/model"
	"sync"
	"time"
)

type ParserManager struct {
	parsers []interfaces.Parser
}

func NewParserManager(parsers ...interfaces.Parser) *ParserManager {
	return &ParserManager{
		parsers: parsers,
	}
}

type SearchResult struct {
	ParserName string
	Vacancies  []model.Vacancy
	Error      error
	Duration   time.Duration
}

// ConcurrentSearch выполняет поиск во всех парсерах одновременно
func (pm *ParserManager) ConcurrentSearch(params interfaces.SearchParams) ([]SearchResult, error) {
	var wg sync.WaitGroup
	results := make(chan SearchResult, len(pm.parsers))

	// Запускаем горутины для каждого парсера
	for _, parser := range pm.parsers {
		wg.Add(1)
		go func(p interfaces.Parser) {
			defer wg.Done()

			start := time.Now()
			vacancies, err := p.SearchVacancies(params)
			duration := time.Since(start)

			results <- SearchResult{
				ParserName: p.GetName(),
				Vacancies:  vacancies,
				Error:      err,
				Duration:   duration,
			}
		}(parser)
	}

	// Закрываем канал после завершения всех горутин
	go func() {
		wg.Wait()
		close(results)
	}()

	// Собираем результаты
	var searchResults []SearchResult
	for result := range results {
		searchResults = append(searchResults, result)
	}

	return searchResults, nil
}

// ConcurrentSearchWithTimeout с таймаутом
func (pm *ParserManager) ConcurrentSearchWithTimeout(ctx context.Context, params interfaces.SearchParams, timeout time.Duration) ([]SearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan SearchResult, len(pm.parsers))

	for _, parser := range pm.parsers {
		wg.Add(1)
		go func(p interfaces.Parser) {
			defer wg.Done()

			// Создаем канал для результата
			resultChan := make(chan SearchResult, 1)

			go func() {
				start := time.Now()
				vacancies, err := p.SearchVacancies(params)
				duration := time.Since(start)

				resultChan <- SearchResult{
					ParserName: p.GetName(),
					Vacancies:  vacancies,
					Error:      err,
					Duration:   duration,
				}
			}()

			select {
			case <-ctx.Done():
				// Таймаут или отмена
				results <- SearchResult{
					ParserName: p.GetName(),
					Error:      fmt.Errorf("timeout exceeded"),
				}
			case result := <-resultChan:
				results <- result
			}
		}(parser)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var searchResults []SearchResult
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
