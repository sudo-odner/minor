package models

import (
	"encoding/json"
)

type Payload struct {
	Op int             `json:"op"` // Код операции (1-сообщение, 2-статус, 3-печать)
	T  string          `json:"t"`  // Тип события (MESSAGE_CREATE, etc)
	D  json.RawMessage `json:"d"`  // Данные
}