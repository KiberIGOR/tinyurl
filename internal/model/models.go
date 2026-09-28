package model

// Request описывает запрос пользователя.
type Request struct {
	URL string `json:"url"`
}

// ResponsePayload описывает ответ
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

type UserResponse struct {
	OriginalURL string `json:"original_url"`
	ShortURL string `json:"short_url"`
}

