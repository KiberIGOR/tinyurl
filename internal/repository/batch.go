package repository

// URLEntry — пара short_url / original_url для пакетной записи в хранилище.
type URLEntry struct {
	ShortURL    string
	OriginalURL string
}

// BatchSaveResult описывает результат пакетной вставки.
type BatchSaveResult struct {
	// Retries — записи, не вставленные из-за коллизии по short_url.
	Retries []URLEntry
	// Existing — original_url уже есть в хранилище, значение — существующий short_url.
	Existing map[string]string
}
