// result_handler.go
package manager

import (
	"errors"
	"fmt"
	"parser/internal/domain/models"
)

func (pm *ParsersManager) handleSearchResult(results []models.SearchResult, err error, params models.SearchParams) ([]models.SearchResult, error) {
	if err == nil {
		return results, nil
	}

	// Проверяем, это ошибка Circuit Breaker или другая ошибка
	var cbErr pm.circuitBreaker.ErrCircuitBreakerOpen
	
	if errors.As(err, &cbErr) {
		return pm.handleCircuitBreakerOpen(params, cbErr)
	}

	// Другие ошибки
	if len(results) > 0 {
		// Частичный успех - логируем ошибку, но возвращаем результаты
		fmt.Printf("⚠️  Частичный успех: %v\n", err)
		return results, nil
	}

	// Полный сбой
	return nil, fmt.Errorf("❌ Ошибка поиска: %v", err)
}

func (pm *ParsersManager) handleCircuitBreakerOpen(params models.SearchParams, cbErr error) ([]models.SearchResult, error) {
	fmt.Println("🚨 Circuit Breaker открыт - используем кэшированные данные")

	// Пробуем разные стратегии fallback
	if results, ok := pm.tryFallbackStrategies(params); ok {
		return results, nil
	}

	// Fallback не сработал
	return nil, fmt.Errorf("❌ Сервис временно недоступен. Попробуйте позже. (Circuit Breaker открыт)")
}

func (pm *ParsersManager) tryFallbackStrategies(params models.SearchParams) ([]models.SearchResult, bool) {
	// Стратегия 1: Поиск по более общему ключу
	if results, ok := pm.tryGeneralCacheKey(params); ok {
		return results, true
	}

	// Стратегия 2: Поиск похожих запросов
	if results, ok := pm.trySimilarQueries(params); ok {
		return results, true
	}

	// Стратегия 3: Статические/дефолтные данные
	if results, ok := pm.tryStaticFallback(params); ok {
		return results, true
	}

	return nil, false
}

func (pm *ParsersManager) tryGeneralCacheKey(params models.SearchParams) ([]models.SearchResult, bool) {
	cacheKey := fmt.Sprintf("fallback:%s", params.Text)
	if cached, ok := pm.searchCache.GetItem(cacheKey); ok {
		if results, ok := cached.([]models.SearchResult); ok {
			fmt.Println("✅ Найдены кэшированные данные для fallback")
			return results, true
		}
	}
	return nil, false
}
