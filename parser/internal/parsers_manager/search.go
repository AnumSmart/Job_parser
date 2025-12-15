package parsers_manager

import (
	"context"
	"fmt"
	"parser/internal/domain/models"
	"time"
)

/*
func (pm *ParsersManager) search(ctx context.Context, params models.SearchParams) ([]models.SearchResult, error) {
	var results []models.SearchResult

	// Используем глобальный Circuit Breaker
	err := pm.circuitBreaker.Execute(func() error {
		var err error
		results, err = pm.executeSearch(ctx, params)
		return err
	})

	// Обрабатываем ошибки Circuit Breaker
	return pm.handleSearchResult(results, err, params)
}
*/

func (pm *ParsersManager) search(ctx context.Context, params models.SearchParams) ([]models.SearchResult, error) {
	// Создаем канал для получения результата
	resultChan := make(chan *models.JobResult, 1)

	// Создаем задание
	job := &models.SearchJob{
		ID:         "1",
		Params:     params,
		ResultChan: resultChan,
		CreatedAt:  time.Now(),
	}

	// Пытаемся добавить в очередь с таймаутом
	select {
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("❌ Очередь заданий переполнена, попробуйте позже")
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		pm.jobQueue.Enqueue(job)
	}

	// Ждем результата с таймаутом
	select {
	case result := <-resultChan:
		return result.Results, result.Error
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("❌ Таймаут выполнения поиска")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Основная логика поиска
func (pm *ParsersManager) executeSearch(ctx context.Context, params models.SearchParams) ([]models.SearchResult, error) {

	// Проверяем кэш
	if cachedResults, found := pm.tryGetFromCache(params); found {
		// Только возвращаем кэшированные данные
		// Статус парсеров не трогаем — они не участвовали
		return cachedResults, nil
	}

	// Получаем список парсеров для использования
	parsersToUse := pm.selectParsersForSearch()
	if len(parsersToUse) == 0 {
		return nil, fmt.Errorf("❌ Нет доступных парсеров для поиска")
	}

	// Выполняем поиск через парсеры
	searchResults, err := pm.searchWithParsers(ctx, params, parsersToUse)

	if err != nil {
		return nil, fmt.Errorf("❌ Конкурентный поиск по парсерам - не удался!")
	}

	// Фильтруем результаты: берем только успешные, т.е. те, у которых в models.SearchResult.Error == nil
	successfulResults := pm.filterSuccessfulResults(searchResults)

	// Кэшируем только если есть хотя бы один успешный результат
	if len(successfulResults) > 0 {
		pm.cacheSearchResults(params, successfulResults)
	} else {
		// Ни один парсер не вернул результатов
		// НЕ кэшируем, пробуем снова при следующем запросе
	}

	return searchResults, nil // успех для глобального CB и получение данных парсинга
}

// Формируем слайс стркутур, где поиск прошёл без ошибок
func (pm *ParsersManager) filterSuccessfulResults(results []models.SearchResult) []models.SearchResult {
	var successful []models.SearchResult
	for _, result := range results {
		if result.Error == nil && len(result.Vacancies) > 0 {
			successful = append(successful, result)
		}
	}
	return successful
}
