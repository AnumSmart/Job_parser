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

	fmt.Println(pm.GetParserNames())             //------для дэбага
	fmt.Println(pm.parsersStatusManager.parsers) //------для дэбага, пока тут пустая мапа, нужно фиксить логику
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

	// Успех = нет ошибок ИЛИ есть хоть какие-то результаты
	overallSuccess := err == nil || len(searchResults) > 0

	// Обновляем статусы парсеров
	pm.updateAllParsersStatus(overallSuccess)

	// Если нет результатов вообще - это полный сбой
	if len(searchResults) == 0 {
		return nil, fmt.Errorf("❌ Полный сбой!") // Полный сбой
	}

	// Кэшируем результаты
	if overallSuccess {
		pm.cacheSearchResults(params, searchResults)
	}

	return searchResults, nil
}
