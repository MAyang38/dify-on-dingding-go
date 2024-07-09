package models

import (
	"time"
)

type Question struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Query     string    `json:"query"`
	ChatType  int       `json:"chat_type"`
	GroupName string    `json:"group_name"`
	Reply     string    `json:"reply"`
	UserId    string    `json:"user_id"`
	SessionId string    `json:"session_id"`
	CreatedAt time.Time `json:"created_at"`
}
