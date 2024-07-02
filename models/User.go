package models

import "time"

type CustomTime struct {
	time.Time
}

const customTimeFormat = "2006-01-02 15:04:05"

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	str := string(b)
	// 去掉引号
	str = str[1 : len(str)-1]
	ct.Time, err = time.Parse(customTimeFormat, str)
	return
}

type User struct {
	ID              int        `json:"id" gorm:"primaryKey"`
	Name            string     `json:"name"`
	PermissionLevel int        `json:"permission_level"`
	Type            int        `json:"type"`
	UserID          string     `json:"user_id"`
	CreatedAt       CustomTime `json:"created_at"`
	UpdatedAt       CustomTime `json:"updated_at"`
}
