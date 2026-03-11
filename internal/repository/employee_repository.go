package repository

import (
	"organizational-api/internal/models"

	"gorm.io/gorm"
)

type EmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) Create(emp *models.Employee) error {
	return r.db.Create(emp).Error
}

func (r *EmployeeRepository) GetByDepartmentID(deptID int) ([]models.Employee, error) {
	var employees []models.Employee
	if err := r.db.Where("department_id = ?", deptID).
		Order("created_at ASC, full_name ASC").
		Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}
