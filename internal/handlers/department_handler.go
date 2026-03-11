package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"organizational-api/internal/logger"
	"organizational-api/internal/service"
	"strconv"
	"strings"
	"time"
)

type DepartmentHandler struct {
	service service.DepartmentService
}

func NewDepartmentHandler(service service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

type CreateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id"`
}

type UpdateDepartmentRequest struct {
	Name     *string `json:"name"`
	ParentID *int    `json:"parent_id"`
}

type CreateEmployeeRequest struct {
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	HiredAt  *string `json:"hired_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *DepartmentHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error.Printf("Failed to decode request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	dept, err := h.service.CreateDepartment(req.Name, req.ParentID)
	if err != nil {
		logger.Error.Printf("Failed to create department: %v", err)
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "already exists") {
			respondWithError(w, http.StatusConflict, err.Error())
		} else {
			respondWithError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	logger.Info.Printf("Created department: %d", dept.ID)
	respondWithJSON(w, http.StatusCreated, dept)
}

func (h *DepartmentHandler) GetDepartment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/departments/")
	idStr := strings.Split(path, "/")[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid department ID")
		return
	}

	// Parse query parameters
	depth := 1
	if depthStr := r.URL.Query().Get("depth"); depthStr != "" {
		depth, err = strconv.Atoi(depthStr)
		if err != nil || depth < 1 {
			depth = 1
		}
		if depth > 5 {
			depth = 5
		}
	}

	includeEmployees := true
	if includeStr := r.URL.Query().Get("include_employees"); includeStr != "" {
		includeEmployees = includeStr == "true"
	}

	result, err := h.service.GetDepartment(id, depth, includeEmployees)
	if err != nil {
		logger.Error.Printf("Failed to get department: %v", err)
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, result)
}

func (h *DepartmentHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/departments/")
	idStr := strings.Split(path, "/")[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid department ID")
		return
	}

	var req UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error.Printf("Failed to decode request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	dept, err := h.service.UpdateDepartment(id, req.Name, req.ParentID)
	if err != nil {
		logger.Error.Printf("Failed to update department: %v", err)
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "circular") || strings.Contains(err.Error(), "already exists") {
			respondWithError(w, http.StatusConflict, err.Error())
		} else {
			respondWithError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	logger.Info.Printf("Updated department: %d", dept.ID)
	respondWithJSON(w, http.StatusOK, dept)
}

func (h *DepartmentHandler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/departments/")
	idStr := strings.Split(path, "/")[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid department ID")
		return
	}

	// Parse query parameters
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "cascade"
	}

	var reassignToDeptID *int
	if reassignStr := r.URL.Query().Get("reassign_to_department_id"); reassignStr != "" {
		reassignID, err := strconv.Atoi(reassignStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid reassign_to_department_id")
			return
		}
		reassignToDeptID = &reassignID
	}

	err = h.service.DeleteDepartment(id, mode, reassignToDeptID)
	if err != nil {
		logger.Error.Printf("Failed to delete department: %v", err)
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	logger.Info.Printf("Deleted department: %d", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *DepartmentHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract department ID from path
	path := strings.TrimPrefix(r.URL.Path, "/departments/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		respondWithError(w, http.StatusBadRequest, "Invalid URL")
		return
	}

	deptID, err := strconv.Atoi(parts[0])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid department ID")
		return
	}

	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error.Printf("Failed to decode request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse hired_at if provided
	var hiredAt sql.NullTime
	if req.HiredAt != nil && *req.HiredAt != "" {
		t, err := time.Parse("2006-01-02", *req.HiredAt)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid hired_at format (expected YYYY-MM-DD)")
			return
		}
		hiredAt = sql.NullTime{Time: t, Valid: true}
	}

	emp, err := h.service.CreateEmployee(deptID, req.FullName, req.Position, hiredAt)
	if err != nil {
		logger.Error.Printf("Failed to create employee: %v", err)
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	logger.Info.Printf("Created employee: %d in department: %d", emp.ID, deptID)
	respondWithJSON(w, http.StatusCreated, emp)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		logger.Error.Printf("Failed to marshal response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
