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
    ID string `json:"uuid"`
    ShortURL string `json:"short_url"`
    OriginalURL string `json:"original_url"`
} 

// MassiveRequest описывает большой запрос пользователя.
type MassiveRequest struct {
    ID string `json:"correlation_id"`
    OriginalURL string `json:"original_url"`
    ShortURL string `json:"-"`
}

// MassiveResponse описывает большой ответ
type MassiveResponse struct {
    ID string `json:"correlation_id"`
    ShortURL string `json:"short_url"`
} 