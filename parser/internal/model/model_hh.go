package model

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
	Items []HHVacancy `json:"items"`
	Found int         `json:"found"`
	Pages int         `json:"pages"`
}
