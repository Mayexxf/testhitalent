package router

import (
	"net/http"
	"organizational-api/internal/handlers"
	"strings"
)

func SetupRoutes(deptHandler *handlers.DepartmentHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/departments/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/departments/" && r.Method == http.MethodPost {
			deptHandler.CreateDepartment(w, r)
			return
		}

		if len(path) > len("/departments/") {
			if strings.Contains(path, "/employees/") && r.Method == http.MethodPost {
				deptHandler.CreateEmployee(w, r)
				return
			}

			if r.Method == http.MethodGet {
				deptHandler.GetDepartment(w, r)
				return
			}

			if r.Method == http.MethodPatch {
				deptHandler.UpdateDepartment(w, r)
				return
			}

			if r.Method == http.MethodDelete {
				deptHandler.DeleteDepartment(w, r)
				return
			}
		}

		http.Error(w, "Not found", http.StatusNotFound)
	})

	return mux
}
