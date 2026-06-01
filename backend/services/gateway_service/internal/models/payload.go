package models

import (
	"encoding/json"
)

type WSPayload struct {
    Op int             `json:"op"` // 1-Dispatch (Событие), 2-Heartbeat, 3-Typing
    T  string          `json:"t"`  // Название события (MESSAGE_CREATE, PRESENCE_UPDATE)
    D  json.RawMessage `json:"d"`  // Данные события
}