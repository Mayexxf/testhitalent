package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"organizational-api/internal/logger"
	"organizational-api/internal/models"
	"organizational-api/internal/repository/mocks"
	"organizational-api/internal/service"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockHandler(t *testing.T) (*DepartmentHandler, *mocks.MockDepartmentRepository, *mocks.MockEmployeeRepository) {
	logger.Init()

	ctrl := gomock.NewController(t)

	deptRepo := mocks.NewMockDepartmentRepository(ctrl)
	empRepo := mocks.NewMockEmployeeRepository(ctrl)

	svc := service.NewDepartmentService(deptRepo, empRepo)
	handler := NewDepartmentHandler(svc)

	t.Cleanup(ctrl.Finish)

	return handler, deptRepo, empRepo
}

func TestCreateDepartment(t *testing.T) {
	handler, deptRepo, _ := setupMockHandler(t)

	payload := CreateDepartmentRequest{
		Name:     "Engineering",
		ParentID: nil,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Ожидаем вызов Create с нужными полями
	deptRepo.EXPECT().
		Create(gomock.Any(), gomock.AssignableToTypeOf(&models.Department{})).
		DoAndReturn(func(ctx context.Context, dept *models.Department) error {
			dept.ID = 42 // Симулируем присвоение ID от БД
			return nil
		}).
		Times(1)

	handler.CreateDepartment(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var dept models.Department
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &dept))
	assert.Equal(t, 42, dept.ID)
	assert.Equal(t, "Engineering", dept.Name)
	assert.Nil(t, dept.ParentID)
}

func TestGetDepartment(t *testing.T) {
	handler, deptRepo, empRepo := setupMockHandler(t)

	companyID := 1
	engID := 2

	// Мокаем получение компании
	deptRepo.EXPECT().
		GetByID(gomock.Any(), companyID).
		Return(&models.Department{
			ID:   companyID,
			Name: "Company",
		}, nil)

	// Мокаем сотрудников компании
	empRepo.EXPECT().
		GetByDepartmentID(gomock.Any(), companyID).
		Return([]*models.Employee{
			{ID: 100, FullName: "CEO", Position: "Chief Executive", DepartmentID: companyID},
		}, nil)

	// Мокаем получение детей (depth=2, но передается 1 как оставшаяся глубина)
	deptRepo.EXPECT().
		GetChildren(gomock.Any(), companyID, 1).
		Return([]*models.Department{
			{ID: engID, Name: "Engineering", ParentID: &companyID},
		}, nil)

	// Мокаем сотрудников для Engineering
	empRepo.EXPECT().
		GetByDepartmentID(gomock.Any(), engID).
		Return([]*models.Employee{}, nil)

	url := fmt.Sprintf("/departments/%d?depth=2", companyID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	handler.GetDepartment(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))

	assert.Equal(t, float64(companyID), result["id"])
	assert.Equal(t, "Company", result["name"])

	employees, ok := result["employees"].([]interface{})
	require.True(t, ok)
	assert.Len(t, employees, 1)

	children, ok := result["children"].([]interface{})
	require.True(t, ok)
	assert.Len(t, children, 1)
}

func TestUpdateDepartment(t *testing.T) {
	handler, deptRepo, _ := setupMockHandler(t)

	engID := 2
	oldName := "Engineering"
	newName := "Product Engineering"

	// Сначала GetByID
	deptRepo.EXPECT().
		GetByID(gomock.Any(), engID).
		Return(&models.Department{ID: engID, Name: oldName}, nil)

	// Затем Update
	deptRepo.EXPECT().
		Update(gomock.Any(), gomock.AssignableToTypeOf(&models.Department{})).
		DoAndReturn(func(ctx context.Context, dept *models.Department) error {
			assert.Equal(t, newName, dept.Name)
			return nil
		}).
		Times(1)

	payload := UpdateDepartmentRequest{
		Name: &newName,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("/departments/%d", engID)
	req := httptest.NewRequest(http.MethodPatch, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateDepartment(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var dept models.Department
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &dept))
	assert.Equal(t, newName, dept.Name)
}

func TestDeleteDepartment(t *testing.T) {
	handler, deptRepo, _ := setupMockHandler(t)

	engID := 2

	// Сначала GetByID для проверки существования
	deptRepo.EXPECT().
		GetByID(gomock.Any(), engID).
		Return(&models.Department{ID: engID, Name: "Engineering"}, nil)

	// Затем Delete
	deptRepo.EXPECT().
		Delete(gomock.Any(), engID, "cascade").
		Return(nil).
		Times(1)

	url := fmt.Sprintf("/departments/%d?mode=cascade", engID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	handler.DeleteDepartment(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCreateEmployee(t *testing.T) {
	handler, deptRepo, empRepo := setupMockHandler(t)

	companyID := 1

	// Сначала проверка существования департамента
	deptRepo.EXPECT().
		GetByID(gomock.Any(), companyID).
		Return(&models.Department{ID: companyID, Name: "Company"}, nil)

	// Затем создание сотрудника
	empRepo.EXPECT().
		Create(gomock.Any(), gomock.AssignableToTypeOf(&models.Employee{})).
		DoAndReturn(func(ctx context.Context, emp *models.Employee) error {
			emp.ID = 101
			return nil
		}).
		Times(1)

	payload := CreateEmployeeRequest{
		FullName: "John Doe",
		Position: "Developer",
		HiredAt:  nil,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("/departments/%d/employees/", companyID)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateEmployee(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var emp models.Employee
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &emp))
	assert.Equal(t, 101, emp.ID)
	assert.Equal(t, "John Doe", emp.FullName)
	assert.False(t, emp.HiredAt.Valid)
}

// Убери ненужные хелперы, если они больше не используются
// func intPtr(i int) *int { return &i }
// func stringPtr(s string) *string { return &s }
