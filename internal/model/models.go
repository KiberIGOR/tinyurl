// Package model содержит DTO HTTP API и формата файлового хранилища.
package model

// Request — тело POST /api/shorten.
type Request struct {
	URL string `json:"url"`
}

// Response — ответ POST /api/shorten.
type Response struct {
	URL string `json:"result"`
}

// MemoryString описывает запись в файл
type MemoryString struct {
	ID          string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
}

// BatchRequest описывает элемент пакетного запроса.
type BatchRequest struct {
	ID          string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
}

// BatchResponse описывает элемент пакетного ответа.
type BatchResponse struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

// UserResponse — элемент списка GET /api/user/urls.
type UserResponse struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

// DeleteRequest — элемент JSON-массива в DELETE /api/user/urls (short id).
type DeleteRequest string
