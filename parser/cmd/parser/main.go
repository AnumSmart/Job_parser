package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"parser/internal/interfaces"
	"parser/internal/manager"
	"parser/internal/model"
	"parser/internal/parser"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println("🚀 Multi-Source Vacancy Parser запущен!")
	fmt.Println("==========================")

	// Создаём парсеры
	hhParser := parser.NewHHParser()
	sjParser := parser.NewSuperJobParser(os.Getenv("SUPERJOB_API_KEY"))

	// Создаём менеджер парсеров
	parserManager := manager.NewParserManager(hhParser, sjParser)

	// Основной цикл приложения
	scanner := bufio.NewScanner(os.Stdin)

	for {
		printMenu()
		fmt.Print("Выберите действие: ")

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			multiSearch(parserManager, scanner)
		case "2":
			searchByQuery(hhParser, scanner)
		case "3":
			getVacancyDetails(hhParser, scanner)
		case "4":
			fmt.Println("👋 До свидания!")
			return
		default:
			fmt.Println("❌ Неверный выбор. Попробуйте снова.")
		}

		fmt.Println()
	}
}

func printMenu() {
	fmt.Println("📋 Меню:")
	fmt.Println("1. Поиск вакансий (расширенный)")
	fmt.Println("2. Быстрый поиск по запросу")
	fmt.Println("3. Получить детали вакансии по ID")
	fmt.Println("4. Выход")
}

// Функция для мульти-поиска
func multiSearch(parserManager *manager.ParserManager, scanner *bufio.Scanner) {
	fmt.Println("\n🌐 Мульти-поиск вакансий")

	var params interfaces.SearchParams

	fmt.Print("Введите поисковый запрос: ")
	if scanner.Scan() {
		params.Text = strings.TrimSpace(scanner.Text())
	}

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

	fmt.Println("⏳ Ищем вакансии во всех источниках...")

	ctx := context.Background()
	results, err := parserManager.ConcurrentSearchWithTimeout(ctx, params, 10*time.Second)

	if err != nil {
		fmt.Printf("❌ Ошибка при поиске: %v\n", err)
		return
	}

	printMultiSearchResults(results)
}

func printMultiSearchResults(results []manager.SearchResult) {
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
			if i >= 3 {
				break
			}
			fmt.Printf("      %d. %s - %s\n", i+1, vacancy.Name, vacancy.GetSalaryString())
		}

		if len(result.Vacancies) > 3 {
			fmt.Printf("      ... и ещё %d\n", len(result.Vacancies)-3)
		}
	}

	fmt.Printf("\n🎯 Всего найдено: %d вакансий\n", totalVacancies)
}

func searchByQuery(hhParser *parser.HHParser, scanner *bufio.Scanner) {
	fmt.Println("\n⚡ Быстрый поиск")

	fmt.Print("Введите поисковый запрос: ")
	if !scanner.Scan() {
		return
	}

	query := strings.TrimSpace(scanner.Text())
	if query == "" {
		fmt.Println("❌ Запрос не может быть пустым")
		return
	}

	fmt.Print("Количество вакансий (max 100): ")
	var limit int = 10
	if scanner.Scan() {
		limitStr := strings.TrimSpace(scanner.Text())
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}
	}

	fmt.Println("⏳ Ищем вакансии...")

	vacancies, err := hhParser.SimpleSearch(query, limit)
	if err != nil {
		fmt.Printf("❌ Ошибка при поиске: %v\n", err)
		return
	}

	fmt.Printf("✅ Найдено %d вакансий по запросу '%s'\n", len(vacancies), query)
	printVacancies(vacancies)
}

func getVacancyDetails(hhParser *parser.HHParser, scanner *bufio.Scanner) {
	fmt.Println("\n📄 Детали вакансии")

	fmt.Print("Введите ID вакансии: ")
	if !scanner.Scan() {
		return
	}

	vacancyID := strings.TrimSpace(scanner.Text())
	if vacancyID == "" {
		fmt.Println("❌ ID вакансии не может быть пустым")
		return
	}

	fmt.Println("⏳ Загружаем информацию...")

	vacancy, err := hhParser.GetVacancyByID(vacancyID)
	if err != nil {
		fmt.Printf("❌ Ошибка при загрузке вакансии: %v\n", err)
		return
	}

	printVacancyDetails(vacancy)
}

func printVacancies(vacancies []model.Vacancy) {
	if len(vacancies) == 0 {
		fmt.Println("😞 Вакансии не найдены")
		return
	}

	for i, vacancy := range vacancies {
		fmt.Printf("\n%d. %s\n", i+1, vacancy.Name)
		fmt.Printf("   💼 %s\n", vacancy.Employer.Name)
		fmt.Printf("   💰 %s\n", vacancy.GetSalaryString())
		fmt.Printf("   📍 %s\n", vacancy.Area.Name)
		//fmt.Printf("   🕐 %s\n", formatDate(vacancy.PublishedAt))
		fmt.Printf("   🔗 %s\n", vacancy.URL)
		fmt.Printf("   🆔 %s\n", vacancy.ID)
	}
}

func printVacancyDetails(vacancy *model.Vacancy) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("🏢 %s\n", vacancy.Name)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("💼 Работодатель: %s\n", vacancy.Employer.Name)
	fmt.Printf("💰 Зарплата: %s\n", vacancy.GetSalaryString())
	fmt.Printf("📍 Местоположение: %s\n", vacancy.Area.Name)
	//fmt.Printf("🕐 Опубликовано: %s\n", formatDate(vacancy.PublishedAt))
	fmt.Printf("🔗 Ссылка: %s\n", vacancy.URL)
	fmt.Printf("🆔 ID: %s\n", vacancy.ID)

	// Обрезаем описание для читаемости
	description := vacancy.Description
	if len(description) > 500 {
		description = description[:500] + "..."
	}

	if description != "" {
		fmt.Println("\n📝 Описание:")
		fmt.Println(cleanHTML(description))
	}

	fmt.Println(strings.Repeat("=", 50))
}

func formatDate(t time.Time) string {
	return t.Format("02.01.2006 15:04")
}

func cleanHTML(text string) string {
	// Простая очистка HTML тегов
	text = strings.ReplaceAll(text, "<p>", "\n")
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<li>", "• ")

	// Удаляем HTML теги
	var result strings.Builder
	var inTag bool

	for _, ch := range text {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(ch)
		}
	}

	return strings.TrimSpace(result.String())
}
