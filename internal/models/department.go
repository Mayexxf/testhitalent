package models

import (
	"time"
)

type Department struct {
	ID        int          `json:"id" gorm:"primaryKey"`
	Name      string       `json:"name" gorm:"type:varchar(200);not null"`
	ParentID  *int         `json:"parent_id" gorm:"index"`
	CreatedAt time.Time    `json:"created_at" gorm:"autoCreateTime"`
	Employees []Employee   `json:"employees,omitempty" gorm:"foreignKey:DepartmentID;constraint:OnDelete:CASCADE"`
	Children  []Department `json:"children,omitempty" gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`
}

func (Department) TableName() string {
	return "departments"
}
