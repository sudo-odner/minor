package models

type UserStatus int32

const (
	UserStatusUnspecified UserStatus = 0
	UserStatusOnline      UserStatus = 1
	UserStatusIdle        UserStatus = 2
	UserStatusDnd         UserStatus = 3
	UserStatusOffline     UserStatus = 4
)

type Presence struct {
	UserID       string     `json:"user_id"`
	Status       UserStatus `json:"status"`
	CustomStatus string     `json:"custom_status"`
	LastActiveAt int64      `json:"last_active_at"`
}
