package model

// Request описывает запрос пользователя.
type Request struct {
    URL string `json:"url"`
}

// ResponsePayload описывает ответ
type Response struct {
    URL string `json:"result"`
} 