package model

import (
	"encoding/json"
	"fmt"
)

// Vacancy представляет структуру вакансии с HH.ru
type Vacancy struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Salary   Salary   `json:"salary"`
	Employer Employer `json:"employer"`
	Area     Area     `json:"area"`
	//	PublishedAt time.Time `json:"published_at"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// Salary представляет информацию о зарплате
type Salary struct {
	From     int    `json:"from"`
	To       int    `json:"to"`
	Currency string `json:"currency"`
	Gross    bool   `json:"gross"`
}

// Employer представляет информацию о работодателе
type Employer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Area представляет информацию о местоположении
type Area struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SearchResponse представляет ответ от API HH.ru
type SearchResponse struct {
	Items []Vacancy `json:"items"`
	Found int       `json:"found"`
	Pages int       `json:"pages"`
}

// ToJSON преобразует вакансию в JSON строку
func (v Vacancy) ToJSON() string {
	bytes, _ := json.MarshalIndent(v, "", "  ")
	return string(bytes)
}

// GetSalaryString возвращает форматированную строку зарплаты
func (v Vacancy) GetSalaryString() string {
	if v.Salary.From == 0 && v.Salary.To == 0 {
		return "не указана"
	}

	if v.Salary.From > 0 && v.Salary.To > 0 {
		return formatSalary(v.Salary.From, v.Salary.To, v.Salary.Currency)
	} else if v.Salary.From > 0 {
		return formatSalary(v.Salary.From, 0, v.Salary.Currency)
	} else {
		return formatSalary(0, v.Salary.To, v.Salary.Currency)
	}
}

func formatSalary(from, to int, currency string) string {
	if from > 0 && to > 0 {
		return formatNumber(from) + " - " + formatNumber(to) + " " + currency
	} else if from > 0 {
		return "от " + formatNumber(from) + " " + currency
	} else {
		return "до " + formatNumber(to) + " " + currency
	}
}

// эта функция разделяет пробелами тысячи
func formatNumber(num int) string {
	if num >= 1000 {
		// Рекурсивно обрабатываем тысячи и добавляем пробел
		return formatNumber(num/1000) + " " + fmt.Sprintf("%03d", num%1000)
	}
	// Базовый случай - возвращаем число как строку
	return fmt.Sprintf("%d", num)
}
