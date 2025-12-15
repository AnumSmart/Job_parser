// parser_executor.go
package manager

import (
	"context"
	"fmt"
	"parser/internal/domain/models"
	"sync"
	"time"
)

type ParserSearchExecutor struct {
	manager       *ParsersManager
	parserNames   []string
	maxConcurrent int
}

func NewParserSearchExecutor(manager *ParsersManager, parserNames []string, maxConcurrent int) *ParserSearchExecutor {
	return &ParserSearchExecutor{
		manager:       manager,
		parserNames:   parserNames,
		maxConcurrent: maxConcurrent,
	}
}

func (e *ParserSearchExecutor) Execute(ctx context.Context, params models.SearchParams) ([]models.SearchResult, []error) {
	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		results    []models.SearchResult
		errors     []error
		semaphore  = make(chan struct{}, e.maxConcurrent)
		resultChan = make(chan *parserResult, len(e.parserNames))
	)

	// Запускаем поиск
	for _, parserName := range e.parserNames {
		wg.Add(1)
		go e.searchWithParser(ctx, &wg, semaphore, resultChan, parserName, params)
	}

	// Ждем завершения
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Собираем результаты
	e.collectResults(resultChan, &results, &errors, mu)

	return results, errors
}

type parserResult struct {
	result *models.SearchResult
	error  error
	parser string
}

func (e *ParserSearchExecutor) searchWithParser(
	ctx context.Context,
	wg *sync.WaitGroup,
	semaphore chan struct{},
	resultChan chan<- *parserResult,
	parserName string,
	params models.SearchParams,
) {
	defer wg.Done()

	// Получаем семафор
	if !e.acquireSemaphore(ctx, semaphore, parserName) {
		resultChan <- &parserResult{
			error:  fmt.Errorf("таймаут при получении семафора"),
			parser: parserName,
		}
		return
	}
	defer e.releaseSemaphore(semaphore)

	// Находим парсер
	parser := e.manager.findParserByName(parserName)
	if parser == nil {
		resultChan <- &parserResult{
			error:  fmt.Errorf("парсер не найден"),
			parser: parserName,
		}
		return
	}

	// Выполняем поиск
	vacancies, err := parser.SearchVacancies(ctx, params)

	// Обновляем статус
	e.manager.updateParserStatus(parserName, err == nil, err)

	// Формируем результат
	if err != nil {
		resultChan <- &parserResult{
			error:  fmt.Errorf("%s: %v", parserName, err),
			parser: parserName,
		}
		return
	}

	resultChan <- &parserResult{
		result: &models.SearchResult{
			Source:    parserName,
			Vacancies: vacancies,
			Timestamp: time.Now(),
		},
		parser: parserName,
	}
}

func (e *ParserSearchExecutor) acquireSemaphore(ctx context.Context, semaphore chan struct{}, parserName string) bool {
	select {
	case semaphore <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

func (e *ParserSearchExecutor) releaseSemaphore(semaphore chan struct{}) {
	<-semaphore
}

func (e *ParserSearchExecutor) collectResults(
	resultChan <-chan *parserResult,
	results *[]models.SearchResult,
	errors *[]error,
	mu sync.Mutex,
) {
	for result := range resultChan {
		mu.Lock()
		if result.error != nil {
			*errors = append(*errors, result.error)
		} else if result.result != nil {
			*results = append(*results, *result.result)
		}
		mu.Unlock()
	}
}
