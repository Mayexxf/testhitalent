package service

import (
	"database/sql"
	"fmt"
	"organizational-api/internal/models"
	"organizational-api/internal/repository"
	"strings"

	"gorm.io/gorm"
)

type DepartmentService struct {
	deptRepo repository.DepartmentRepository
	empRepo  repository.EmployeeRepository
}

func NewDepartmentService(deptRepo repository.DepartmentRepository, empRepo repository.EmployeeRepository) DepartmentService {
	return DepartmentService{
		deptRepo: deptRepo,
		empRepo:  empRepo,
	}
}

func (s *DepartmentService) CreateDepartment(name string, parentID *int) (*models.Department, error) {
	// Validate name
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 200 {
		return nil, fmt.Errorf("name must be between 1 and 200 characters")
	}

	// Check if parent exists
	if parentID != nil {
		_, err := s.deptRepo.GetByID(*parentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("parent department not found")
			}
			return nil, err
		}
	}

	dept := &models.Department{
		Name:     name,
		ParentID: parentID,
	}

	if err := s.deptRepo.Create(dept); err != nil {
		return nil, err
	}

	return dept, nil
}

func (s *DepartmentService) GetDepartment(id int, depth int, includeEmployees bool) (map[string]interface{}, error) {
	if depth < 1 {
		depth = 1
	}
	if depth > 5 {
		depth = 5
	}

	var dept *models.Department
	var err error

	if includeEmployees {
		dept, err = s.deptRepo.GetByIDWithEmployees(id)
	} else {
		dept, err = s.deptRepo.GetByID(id)
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department not found")
		}
		return nil, err
	}

	result := s.buildDepartmentTree(dept, depth, includeEmployees, 1)
	return result, nil
}

func (s *DepartmentService) buildDepartmentTree(dept *models.Department, maxDepth int, includeEmployees bool, currentDepth int) map[string]interface{} {
	result := map[string]interface{}{
		"id":         dept.ID,
		"name":       dept.Name,
		"parent_id":  dept.ParentID,
		"created_at": dept.CreatedAt,
	}

	if includeEmployees {
		employees := make([]map[string]interface{}, 0)
		for _, emp := range dept.Employees {
			empMap := map[string]interface{}{
				"id":            emp.ID,
				"department_id": emp.DepartmentID,
				"full_name":     emp.FullName,
				"position":      emp.Position,
				"created_at":    emp.CreatedAt,
			}
			if emp.HiredAt.Valid {
				empMap["hired_at"] = emp.HiredAt.Time
			} else {
				empMap["hired_at"] = nil
			}
			employees = append(employees, empMap)
		}
		result["employees"] = employees
	}

	if currentDepth < maxDepth {
		children, err := s.deptRepo.GetChildren(dept.ID)
		if err == nil && len(children) > 0 {
			childrenData := make([]map[string]interface{}, 0)
			for _, child := range children {
				if includeEmployees {
					childWithEmployees, _ := s.deptRepo.GetByIDWithEmployees(child.ID)
					if childWithEmployees != nil {
						child = *childWithEmployees
					}
				}
				childData := s.buildDepartmentTree(&child, maxDepth, includeEmployees, currentDepth+1)
				childrenData = append(childrenData, childData)
			}
			result["children"] = childrenData
		} else {
			result["children"] = []interface{}{}
		}
	}

	return result
}

func (s *DepartmentService) UpdateDepartment(id int, name *string, parentID *int) (*models.Department, error) {
	dept, err := s.deptRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department not found")
		}
		return nil, err
	}

	if name != nil {
		*name = strings.TrimSpace(*name)
		if *name == "" || len(*name) > 200 {
			return nil, fmt.Errorf("name must be between 1 and 200 characters")
		}
		dept.Name = *name
	}

	if parentID != nil {
		// Check if trying to set self as parent
		if *parentID == id {
			return nil, fmt.Errorf("cannot set department as its own parent")
		}

		// Check if new parent exists
		if *parentID != 0 {
			_, err := s.deptRepo.GetByID(*parentID)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, fmt.Errorf("parent department not found")
				}
				return nil, err
			}

			// Check for circular dependency
			if err := s.checkCircularDependency(id, *parentID); err != nil {
				return nil, err
			}
		}

		if *parentID == 0 {
			dept.ParentID = nil
		} else {
			dept.ParentID = parentID
		}
	}

	if err := s.deptRepo.Update(dept); err != nil {
		return nil, err
	}

	return dept, nil
}

func (s *DepartmentService) checkCircularDependency(deptID int, newParentID int) error {
	// Get all ancestors of the new parent
	current := newParentID
	visited := make(map[int]bool)

	for current != 0 {
		if current == deptID {
			return fmt.Errorf("circular dependency detected: cannot move department into its own subtree")
		}

		if visited[current] {
			break
		}
		visited[current] = true

		dept, err := s.deptRepo.GetByID(current)
		if err != nil {
			return err
		}

		if dept.ParentID == nil {
			break
		}
		current = *dept.ParentID
	}

	return nil
}

func (s *DepartmentService) DeleteDepartment(id int, mode string, reassignToDeptID *int) error {
	dept, err := s.deptRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("department not found")
		}
		return err
	}

	switch mode {
	case "cascade":
		// Cascade delete is handled by database constraints
		return s.deptRepo.Delete(dept.ID)

	case "reassign":
		if reassignToDeptID == nil {
			return fmt.Errorf("reassign_to_department_id is required for reassign mode")
		}

		// Check if target department exists
		_, err := s.deptRepo.GetByID(*reassignToDeptID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("target department not found")
			}
			return err
		}

		// Check if trying to reassign to self
		if *reassignToDeptID == id {
			return fmt.Errorf("cannot reassign to the same department")
		}

		// Get all employees in this department and all child departments
		employees, err := s.deptRepo.GetAllEmployeesInDepartmentAndChildren(id)
		if err != nil {
			return err
		}

		// Reassign all employees to target department
		for _, emp := range employees {
			emp.DepartmentID = *reassignToDeptID
		}

		// Update employees in a transaction would be better, but for simplicity:
		if len(employees) > 0 {
			// Get all descendant department IDs
			descendantIDs, err := s.deptRepo.GetAllDescendantIDs(id)
			if err != nil {
				return err
			}

			// Reassign employees from all descendants
			for _, deptID := range descendantIDs {
				if err := s.deptRepo.ReassignEmployees(deptID, *reassignToDeptID); err != nil {
					return err
				}
			}
		}

		// Delete the department (cascade will delete child departments)
		return s.deptRepo.Delete(dept.ID)

	default:
		return fmt.Errorf("invalid mode: must be 'cascade' or 'reassign'")
	}
}

func (s *DepartmentService) CreateEmployee(deptID int, fullName, position string, hiredAt sql.NullTime) (*models.Employee, error) {
	// Check if department exists
	_, err := s.deptRepo.GetByID(deptID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department not found")
		}
		return nil, err
	}

	// Validate fields
	fullName = strings.TrimSpace(fullName)
	if fullName == "" || len(fullName) > 200 {
		return nil, fmt.Errorf("full_name must be between 1 and 200 characters")
	}

	position = strings.TrimSpace(position)
	if position == "" || len(position) > 200 {
		return nil, fmt.Errorf("position must be between 1 and 200 characters")
	}

	emp := &models.Employee{
		DepartmentID: deptID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      hiredAt,
	}

	if err := s.empRepo.Create(emp); err != nil {
		return nil, err
	}

	return emp, nil
}
