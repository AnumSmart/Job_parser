package parsers_manager

import (
	"context"
	"fmt"
	"parser/internal/domain/models"
	"time"
)

// метод менджера парсеров, который формирует джобу для поиска списка вакансий, добавляет эту джобу в очередь и получает результат поиска в канал
// возвращает результат поиска или ошибку
func (pm *ParsersManager) searchVacancies(ctx context.Context, params models.SearchParams) ([]models.SearchVacanciesResult, error) {
	// Создаем канал для получения результата
	//resultChan := make(chan *jobs.JobOutput, 1)

	// Создаем задание
	/*
		job := &models.SearchVacanciesJob{
			ID:         pkg.QuickUUID(),
			Params:     params,
			ResultChan: resultChan,
			CreatedAt:  time.Now(),
		}
	*/

	job := pm.NewSearchJob(params)

	success := pm.tryEnqueueJob(ctx, job, 5*time.Second)

	// Пытаемся добавить в очередь с таймаутом и повторными попытками

	/*
		start := time.Now()
		timeout := 5 * time.Second

		for {
			// Пробуем добавить в очередь
			if pm.jobSearchQueue.Enqueue(job) {
				break // успешно
			}

			// Проверяем таймаут
			if time.Since(start) > timeout {
				close(resultChan) // Закрываем канал, чтобы не было утечек
				return nil, fmt.Errorf("❌ Очередь заданий переполнена, попробуйте позже")
			}

			// Проверяем отмену контекста
			select {
			case <-ctx.Done():
				close(resultChan)
				return nil, ctx.Err()
			default:
				// Небольшая пауза перед следующей попыткой
				time.Sleep(50 * time.Millisecond)
			}
		}
	*/
	if success {
		// Ждем результата с таймаутом
		select {
		case result := <-job.ResultChan:
			resultCheked, ok := result.Data.([]models.SearchVacanciesResult)
			if ok {
				return resultCheked, result.Error
			}
			return nil, result.Error

		case <-time.After(30 * time.Second):
			return nil, fmt.Errorf("❌ Таймаут выполнения поиска")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, fmt.Errorf("❌ Джоба не была добавлена в очередь")
}

// Основная логика поиска списка вакансий по всем доступным парсерам
func (pm *ParsersManager) executeSearch(ctx context.Context, params models.SearchParams) ([]models.SearchVacanciesResult, error) {

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
func (pm *ParsersManager) filterSuccessfulResults(results []models.SearchVacanciesResult) []models.SearchVacanciesResult {
	var successful []models.SearchVacanciesResult
	for _, result := range results {
		if result.Error == nil && len(result.Vacancies) > 0 {
			successful = append(successful, result)
		}
	}
	return successful
}
