package repository

import (
	"context"
	"fmt"
	"organizational-api/internal/models"

	"gorm.io/gorm"
)

type DepartmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) Create(ctx context.Context, dept *models.Department) error {
	// Check uniqueness of name within the same parent
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Department{}).Where("name = ?", dept.Name)

	if dept.ParentID != nil {
		query = query.Where("parent_id = ?", *dept.ParentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	if err := query.Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("department with name '%s' already exists in this parent", dept.Name)
	}

	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *DepartmentRepository) GetByID(ctx context.Context, id int) (*models.Department, error) {
	var dept models.Department
	if err := r.db.WithContext(ctx).First(&dept, id).Error; err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *DepartmentRepository) GetByIDWithEmployees(id int) (*models.Department, error) {
	var dept models.Department
	if err := r.db.Preload("Employees", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC, full_name ASC")
	}).First(&dept, id).Error; err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *DepartmentRepository) GetChildren(ctx context.Context, parentID int, depth int) ([]*models.Department, error) {
	var children []*models.Department
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return nil, err
	}
	return children, nil
}

func (r *DepartmentRepository) Update(ctx context.Context, dept *models.Department) error {
	// Check uniqueness of name within the same parent (excluding current department)
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Department{}).
		Where("name = ? AND id != ?", dept.Name, dept.ID)

	if dept.ParentID != nil {
		query = query.Where("parent_id = ?", *dept.ParentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	if err := query.Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("department with name '%s' already exists in this parent", dept.Name)
	}

	return r.db.WithContext(ctx).Save(dept).Error
}

func (r *DepartmentRepository) Delete(ctx context.Context, id int, mode string) error {
	return r.db.WithContext(ctx).Delete(&models.Department{}, id).Error
}

func (r *DepartmentRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Department, int64, error) {
	var departments []*models.Department
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Department{})

	for key, value := range filter {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&departments).Error; err != nil {
		return nil, 0, err
	}

	return departments, total, nil
}

func (r *DepartmentRepository) GetAllDescendantIDs(ctx context.Context, id int) ([]int, error) {
	var ids []int
	ids = append(ids, id)

	children, err := r.GetChildren(ctx, id, 0)
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		childIDs, err := r.GetAllDescendantIDs(ctx, child.ID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, childIDs...)
	}

	return ids, nil
}

func (r *DepartmentRepository) ReassignEmployees(fromDeptID, toDeptID int) error {
	return r.db.Model(&models.Employee{}).
		Where("department_id = ?", fromDeptID).
		Update("department_id", toDeptID).Error
}

func (r *DepartmentRepository) GetAllEmployeesInDepartmentAndChildren(ctx context.Context, id int) ([]models.Employee, error) {
	descendantIDs, err := r.GetAllDescendantIDs(ctx, id)
	if err != nil {
		return nil, err
	}

	var employees []models.Employee
	if err := r.db.WithContext(ctx).Where("department_id IN ?", descendantIDs).Find(&employees).Error; err != nil {
		return nil, err
	}

	return employees, nil
}
