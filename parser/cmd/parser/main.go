package main

import (
	"bufio"
	"fmt"
	"parser/configs"
	"parser/internal/inmemory_cache"
	"parser/internal/manager"
	"parser/internal/model"
	"parser/internal/parser"

	"os"
	"strings"
	"time"
)

const (
	numOfShards = 7
)

func main() {
	fmt.Println("🚀 Multi-Source Vacancy Parser запущен!")
	fmt.Println("==========================")

	// создаём config
	conf, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	//создаём экземпляр inmemory cache
	cacheSh := inmemory_cache.NewInmemoryShardedCache(numOfShards, time.Minute)

	// Создаём парсеры
	hhParser := parser.NewHHParser()
	sjParser := parser.NewSuperJobParser(conf.Api_conf.SJ_api_key)

	// Создаём менеджер парсеров
	parserManager := manager.NewParserManager(conf, hhParser, sjParser)

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
			parserManager.MultiSearch(scanner, cacheSh)
		case "2":
			getVacancyDetails(hhParser, scanner)
		case "3":
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
	fmt.Println("2. Получить детали вакансии по ID")
	fmt.Println("3. Выход")
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

func printVacancies(vacancies []model.HHVacancy) {
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

func printVacancyDetails(vacancy *model.HHVacancy) {
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
