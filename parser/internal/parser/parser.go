package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"parser/internal/model"
	"strconv"
	"time"
)

// HHParser представляет парсер для HH.ru API
type HHParser struct {
	baseURL    string
	httpClient *http.Client
}

// NewHHParser создаёт новый экземпляр парсера (конструктор для парсера)
func NewHHParser() *HHParser {
	return &HHParser{
		baseURL: "https://api.hh.ru/vacancies",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchParams параметры поиска вакансий
type SearchParams struct {
	Text    string // Поисковый запрос
	Area    string // Регион (например, "1" - Москва, "2" - СПб)
	PerPage int    // Количество вакансий на странице (max 100)
	Page    int    // Номер страницы
}

func (p *HHParser) SearchVacancies(params SearchParams) ([]model.Vacancy, error) {
	// Строим URL с параметрами
	apiURL, err := p.buildURL(params)
	if err != nil {
		return nil, fmt.Errorf("build URL failed: %w", err)
	}

	// Выполняем HTTP запрос
	resp, err := p.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Читаем и парсим ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var searchResponse model.SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %w", err)
	}

	return searchResponse.Items, nil
}

// buildURL строит URL для API запроса
func (p *HHParser) buildURL(params SearchParams) (string, error) {
	u, err := url.Parse(p.baseURL)
	if err != nil {
		return "", err
	}

	query := u.Query()

	if params.Text != "" {
		query.Set("text", params.Text)
	}
	if params.Area != "" {
		query.Set("area", params.Area)
	}

	// Устанавливаем количество вакансий на странице
	perPage := params.PerPage
	if perPage <= 0 || perPage > 100 {
		perPage = 20 // Значение по умолчанию
	}
	query.Set("per_page", strconv.Itoa(perPage))

	if params.Page > 0 {
		query.Set("page", strconv.Itoa(params.Page))
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}

// GetVacancyByID получает детальную информацию о вакансии по ID
func (p *HHParser) GetVacancyByID(vacancyID string) (*model.Vacancy, error) {
	if vacancyID == "" {
		return nil, fmt.Errorf("vacancy ID cannot be empty")
	}

	apiURL := p.baseURL + "/" + vacancyID
	resp, err := p.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var vacancy model.Vacancy
	if err := json.Unmarshal(body, &vacancy); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %w", err)
	}

	return &vacancy, nil
}

// SimpleSearch упрощённый поиск по тексту
func (p *HHParser) SimpleSearch(query string, limit int) ([]model.Vacancy, error) {
	params := SearchParams{
		Text:    query,
		PerPage: limit,
	}
	return p.SearchVacancies(params)
}
