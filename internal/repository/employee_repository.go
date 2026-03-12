package repository

import (
	"context"
	"organizational-api/internal/models"

	"gorm.io/gorm"
)

type EmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) Create(ctx context.Context, emp *models.Employee) error {
	return r.db.WithContext(ctx).Create(emp).Error
}

func (r *EmployeeRepository) GetByID(ctx context.Context, id int) (*models.Employee, error) {
	var emp models.Employee
	if err := r.db.WithContext(ctx).First(&emp, id).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *EmployeeRepository) GetByDepartmentID(ctx context.Context, deptID int) ([]*models.Employee, error) {
	var employees []*models.Employee
	if err := r.db.WithContext(ctx).Where("department_id = ?", deptID).
		Order("created_at ASC, full_name ASC").
		Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

func (r *EmployeeRepository) Update(ctx context.Context, emp *models.Employee) error {
	return r.db.WithContext(ctx).Save(emp).Error
}

func (r *EmployeeRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Employee{}, id).Error
}

func (r *EmployeeRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Employee, int64, error) {
	var employees []*models.Employee
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Employee{})

	for key, value := range filter {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&employees).Error; err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}
