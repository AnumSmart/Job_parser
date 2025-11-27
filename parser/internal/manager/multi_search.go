package manager

import (
	"bufio"
	"context"
	"fmt"
	"parser/internal/domain/models"
	"parser/internal/inmemory_cache"
	"strconv"
	"strings"
	"time"
)

const (
	ctxTimeout = 2 * time.Second
)

// Метод для мульти-поиска
func (pm *ParserManager) MultiSearch(scanner *bufio.Scanner, cash *inmemory_cache.InmemoryShardedCache) {
	fmt.Println("\n🌐 Мульти-поиск вакансий")

	var params models.SearchParams

	// читаем запрос пользователя и сохраняем его в параметрах
	fmt.Print("Введите поисковый запрос: ")
	if scanner.Scan() {
		params.Text = strings.TrimSpace(scanner.Text())
	}

	// читаем необходимое количество вакансий в поиске, но не более 50, сохраняем в параметрах
	fmt.Print("Количество вакансий на источник (max 50): ")
	if scanner.Scan() {
		countStr := strings.TrimSpace(scanner.Text())
		if countStr != "" {
			if count, err := strconv.Atoi(countStr); err == nil && count > 0 {
				params.PerPage = count
			}
		}
	}

	if params.PerPage == 0 {
		params.PerPage = 20
	}

	searchHash, _ := genHashFromSearchParam(params) // ****!!!! нужно обработать ошибку

	// пытаемся найти в кэше данные по заданному хэш ключу
	fmt.Println("⏳ Ищем вакансии в кэше...")
	searchRes, ok := cash.GetItem(searchHash)
	if ok {
		pm.printMultiSearchResults(searchRes, params.PerPage)
		return
	}

	fmt.Println("⏳ Не удалось найти данные в кэше! Ищем вакансии во всех источниках...")

	ctx := context.Background()
	// Запускаем конкурентный поиск по всем источникам, таймаут для отмены контекста в переменной ctxTimeout
	results, err := pm.concurrentSearchWithTimeout(ctx, searchHash, params, ctxTimeout)
	if err != nil {
		fmt.Printf("❌ Ошибка при поиске: %v\n", err)
		return
	}

	//записываем данные в кэш
	cash.AddItemWithTTL(searchHash, results, time.Minute)

	// вызываем функцию вывода в консоль информации о результатах поиска
	pm.printMultiSearchResults(results, params.PerPage)
}
