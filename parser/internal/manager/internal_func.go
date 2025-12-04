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
				vacancies, err := p.SearchVacancies(ctx, params)
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

// метод для построения обратного индекса и хранения его в кэше №2 для индексов и ID вакансий
func (pm *ParserManager) buildReverseIndex(searchHash string, results []models.SearchResult) {
	for _, parserResult := range results {
		for i, vacancy := range parserResult.Vacancies {
			compositeID := fmt.Sprintf("%s_%s", vacancy.Seeker, vacancy.ID)

			indexEntry := models.VacancyIndex{
				SearchHash: searchHash,
				ParserName: parserResult.ParserName,
				Index:      i,
			}

			// Сохраняем в индексный кэш (ТОТ ЖЕ ТИП!), TTL такой же как для кэша поиска
			pm.vacancyIndex.AddItemWithTTL(compositeID, indexEntry, pm.config.Cache.VacancyCacheTTL)
		}
	}
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

/*
метод - обёртка над другими методами.

	Формируем хэш для поиска
	пытаемся поискать в кэше №1
	если не удалось - делаем конкурентный запрос во все доступные сервисы
	кэшируем данные в кэш №1
	строим обратный индекс и кэшируем данные в кэш №2
*/
func (pm *ParserManager) search(ctx context.Context, params models.SearchParams) ([]models.SearchResult, error) {
	// Создаем новый контекст с таймаутом, который будет отменен либо по таймауту,
	// либо когда отменится родительский контекст (что наступит раньше)
	searchCtx, cancel := context.WithTimeout(ctx, pm.config.API.ConcSearchTimeout)
	defer cancel()

	// получаем хэш для поиска
	searchHash, err := genHashFromSearchParam(params)
	if err != nil {
		return nil, fmt.Errorf("❌ Ошибка при генерации поискового хэша: %v\n", err)
	}
	// ---------------------------------------------------------------------------------------------------------------
	// пытаемся найти в кэше данные по заданному хэш ключу
	fmt.Println("⏳ Ищем вакансии в кэше...")

	searchRes, ok := pm.searchCache.GetItem(searchHash)
	if ok {
		// если можно получить данные из кэша №1, то получаем интерфейс.
		// проводим type assertion, проверяем нужный тип
		searchResChecked, ok := searchRes.([]models.SearchResult)
		if !ok {
			fmt.Println("Type assertion after multi-search ---> failed!")
			return nil, fmt.Errorf("❌ Type assertion getting data from search cache ---> failed!\n")
		}
		return searchResChecked, nil
	} else {
		fmt.Println("⏳ Не удалось найти данные в кэше! Ищем вакансии во всех источниках...")
		// передаём созданный контектс searchCtx, чтобы синхронизировать таймауты
		results, err := pm.concurrentSearchWithTimeout(searchCtx, searchHash, params, pm.config.API.ConcSearchTimeout)
		if err != nil {
			return nil, fmt.Errorf("❌ Ошибка при конкурентном поиске данных во внешних источниках: %v\n", err)
		}

		//записываем данные в поисковый кэш №1
		pm.searchCache.AddItemWithTTL(searchHash, results, pm.config.Cache.SearchCacheTTL)

		// Строим обратный индекс и сразу кэшируем его в кэше №2
		pm.buildReverseIndex(searchHash, results)

		return results, nil
	}
}
