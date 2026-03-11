package models

import (
	"database/sql"
	"time"
)

type Employee struct {
	ID           int          `json:"id" gorm:"primaryKey"`
	DepartmentID int          `json:"department_id" gorm:"not null;index"`
	FullName     string       `json:"full_name" gorm:"type:varchar(200);not null;index"`
	Position     string       `json:"position" gorm:"type:varchar(200);not null"`
	HiredAt      sql.NullTime `json:"hired_at,omitempty" gorm:"type:date"`
	CreatedAt    time.Time    `json:"created_at" gorm:"autoCreateTime;index"`
}

func (Employee) TableName() string {
	return "employees"
}
