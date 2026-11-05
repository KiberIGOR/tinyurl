package audit

import "strconv"

const (
	ActionShorten = "shorten"
	ActionFollow  = "follow"
)

// Event — сообщение аудита запроса.
type Event struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id,omitempty"`
	URL    string `json:"url"`
}

func NewShortenEvent(ts int64, userID int, originalURL string) Event {
	return Event{
		TS:     ts,
		Action: ActionShorten,
		UserID: formatUserID(userID),
		URL:    originalURL,
	}
}

func NewFollowEvent(ts int64, userID int, originalURL string) Event {
	return Event{
		TS:     ts,
		Action: ActionFollow,
		UserID: formatUserID(userID),
		URL:    originalURL,
	}
}

func formatUserID(userID int) string {
	if userID <= 0 {
		return ""
	}
	return strconv.Itoa(userID)
}
