package repository

import (
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

func (r *DepartmentRepository) Create(dept *models.Department) error {
	// Check uniqueness of name within the same parent
	var count int64
	query := r.db.Model(&models.Department{}).Where("name = ?", dept.Name)

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

	return r.db.Create(dept).Error
}

func (r *DepartmentRepository) GetByID(id int) (*models.Department, error) {
	var dept models.Department
	if err := r.db.First(&dept, id).Error; err != nil {
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

func (r *DepartmentRepository) GetChildren(parentID int) ([]models.Department, error) {
	var children []models.Department
	if err := r.db.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return nil, err
	}
	return children, nil
}

func (r *DepartmentRepository) Update(dept *models.Department) error {
	// Check uniqueness of name within the same parent (excluding current department)
	var count int64
	query := r.db.Model(&models.Department{}).
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

	return r.db.Save(dept).Error
}

func (r *DepartmentRepository) Delete(id int) error {
	return r.db.Delete(&models.Department{}, id).Error
}

func (r *DepartmentRepository) GetAllDescendantIDs(id int) ([]int, error) {
	var ids []int
	ids = append(ids, id)

	children, err := r.GetChildren(id)
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		childIDs, err := r.GetAllDescendantIDs(child.ID)
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

func (r *DepartmentRepository) GetAllEmployeesInDepartmentAndChildren(id int) ([]models.Employee, error) {
	descendantIDs, err := r.GetAllDescendantIDs(id)
	if err != nil {
		return nil, err
	}

	var employees []models.Employee
	if err := r.db.Where("department_id IN ?", descendantIDs).Find(&employees).Error; err != nil {
		return nil, err
	}

	return employees, nil
}
