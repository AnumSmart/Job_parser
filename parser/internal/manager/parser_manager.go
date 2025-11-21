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
	Vacancies  []model.HHVacancy
	Error      error
	Duration   time.Duration
}

// ConcurrentSearchWithTimeout выполняет поиск во всех парсерах одновременно с таймаутом
func (pm *ParserManager) ConcurrentSearchWithTimeout(ctx context.Context, params interfaces.SearchParams, timeout time.Duration) ([]SearchResult, error) {
	// создаём контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan SearchResult, len(pm.parsers))

	for _, parser := range pm.parsers {
		wg.Add(1)
		go func(p interfaces.Parser) {
			defer wg.Done()

			// Создаем канал для результата и создаём ещё одну горутину, где производим поиск
			// 2я - горутина нужна, чтобы потом использовать select и контролировать отмену контекста
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

	// в этой горутине дожидаемся окончания обработки от всех парсеров и закрываем канал результатов
	go func() {
		wg.Wait()
		close(results)
	}()

	// обьявляем переменную для выходных данных
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
