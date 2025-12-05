package manager

import (
	"bufio"
	"context"
	"fmt"
	"parser/internal/domain/models"
	"strconv"
	"strings"
)

// Метод для мульти-поиска
func (pm *ParsersManager) MultiSearch(scanner *bufio.Scanner) error {
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
	ctx := context.Background()

	// запускаем комплексный метод поиска
	results, err := pm.search(ctx, params)
	if err != nil {
		return err
	}

	// вызываем функцию вывода в консоль информации о результатах поиска
	pm.printMultiSearchResults(results, params.PerPage)
	return nil
}
