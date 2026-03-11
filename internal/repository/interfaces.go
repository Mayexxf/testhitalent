package repository

import (
	"context"
	"organizational-api/internal/models"
)

type DepartmentRepo interface {
	Create(ctx context.Context, dept *models.Department) error
	GetByID(ctx context.Context, id int) (*models.Department, error)
	GetChildren(ctx context.Context, parentID int, depth int) ([]*models.Department, error)
	Update(ctx context.Context, dept *models.Department) error
	Delete(ctx context.Context, id int, mode string) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Department, int64, error)
}

type EmployeeRepo interface {
	Create(ctx context.Context, emp *models.Employee) error
	GetByID(ctx context.Context, id int) (*models.Employee, error)
	GetByDepartmentID(ctx context.Context, deptID int) ([]*models.Employee, error)
	Update(ctx context.Context, emp *models.Employee) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Employee, int64, error)
}
