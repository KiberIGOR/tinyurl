package model

// Request описывает запрос пользователя.
type Request struct {
    URL string `json:"url"`
}

// ResponsePayload описывает ответ
type Response struct {
    URL string `json:"result"`
} 

type MemoryString struct {
    ID string `json:"uuid"`
    ShortURL string `json:"short_url"`
    OriginalURL string `json:"original_url"`
} 