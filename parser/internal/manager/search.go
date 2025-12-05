package manager

import (
	"context"
	"fmt"
	"parser/internal/domain/models"
)

func (pm *ParsersManager) search2(ctx context.Context, params models.SearchParams) ([]models.SearchResult, error) {
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

// Основная логика поиска
func (pm *ParsersManager) executeSearch(ctx context.Context, params models.SearchParams) ([]models.SearchResult, error) {
	// Проверяем кэш
	if cachedResults, found := pm.tryGetFromCache(params); found {
		pm.updateAllParsersStatus(true)
		return cachedResults, nil
	}

	// Получаем список парсеров для использования
	parsersToUse := pm.selectParsersForSearch()
	if len(parsersToUse) == 0 {
		return nil, fmt.Errorf("❌ Нет доступных парсеров для поиска")
	}

	// Выполняем поиск через парсеры
	searchResults, err := pm.searchWithParsers(ctx, params, parsersToUse)
	if err != nil && len(searchResults) == 0 {
		return nil, fmt.Errorf("❌ Полный сбой!") // Полный сбой
	}

	// Кэшируем результаты
	pm.cacheSearchResults(params, searchResults)

	return searchResults, nil
}
