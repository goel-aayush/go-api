package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	//"github.com/goel-aayush/students-api/internal/storage"
	"github.com/goel-aayush/students-api/internal/storage"
	"github.com/goel-aayush/students-api/internal/types"
	"github.com/goel-aayush/students-api/internal/utils/response"
	"github.com/gorilla/mux"
)

func New(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var student types.Student

		// Decode JSON request body
		err := json.NewDecoder(r.Body).Decode(&student)
		if err != nil {
			if errors.Is(err, io.EOF) {
				response.WriteJson(w, http.StatusBadRequest, response.GenralError(fmt.Errorf("empty body")))
			} else {
				response.WriteJson(w, http.StatusBadRequest, response.GenralError(err))
			}
			return
		}

		// Validate request body
		validate := validator.New()
		if err := validate.Struct(student); err != nil {
			if validateErrs, ok := err.(validator.ValidationErrors); ok {
				response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			} else {
				response.WriteJson(w, http.StatusBadRequest, response.GenralError(err))
			}
			return
		}

		// Create student
		lastId, err := storage.CreateStudent(student.Name, student.Email, student.Age)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GenralError(err))
			return
		}

		slog.Info("User created successfully", slog.String("userId", fmt.Sprint(lastId)))
		response.WriteJson(w, http.StatusCreated, map[string]int64{"id": lastId})
	}
}

func GetById(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		slog.Info("Getting a student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GenralError(fmt.Errorf("invalid student ID format")))
			return
		}

		student, err := storage.GetStudentById(intId)
		if err != nil {
			slog.Error("Error getting student", slog.String("id", id))
			response.WriteJson(w, http.StatusInternalServerError, response.GenralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, student)
	}
}

func GetList(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Getting all students")
		students, err := storage.GetStudents()
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GenralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, students)
	}
}

func UpdateStudent(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var updatedData types.Student

		// Decode JSON request body
		if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GenralError(fmt.Errorf("invalid request body")))
			return
		}

		// Get student ID from URL params
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GenralError(fmt.Errorf("invalid student ID format")))
			return
		}

		// Fetch existing student record
		existingStudent, err := storage.GetStudentById(id)
		if err != nil {
			http.Error(w, "Student not found", http.StatusNotFound)
			return
		}

		// Merge the provided fields with the existing student data
		if updatedData.Name != "" {
			existingStudent.Name = updatedData.Name
		}
		if updatedData.Email != "" {
			existingStudent.Email = updatedData.Email
		}
		if updatedData.Age > 0 {
			existingStudent.Age = updatedData.Age
		}

		// Update student in storage
		if err := storage.UpdateStudent(existingStudent); err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GenralError(fmt.Errorf("failed to update student: %v", err)))
			return
		}

		response.WriteJson(w, http.StatusOK, map[string]string{"message": "Student updated successfully"})
	}
}

func RemoveStudent(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		slog.Info("Deleting a student", slog.String("id", id))

		// Convert id to int64
		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GenralError(fmt.Errorf("invalid student ID format")))
			return
		}

		// Attempt to delete the student
		err = storage.RemoveStudent(intId)
		if err != nil {
			slog.Error("Error deleting student", slog.String("id", id))
			response.WriteJson(w, http.StatusNotFound, response.GenralError(fmt.Errorf("student not found or could not be deleted")))
			return
		}

		// Respond with success message
		response.WriteJson(w, http.StatusOK, map[string]string{"message": "Student deleted successfully"})
	}
}
