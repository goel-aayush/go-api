package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/goel-aayush/students-api/internal/config"
	"github.com/goel-aayush/students-api/internal/types"
	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	Db *sql.DB
}

func New(cfg *config.Config) (*Sqlite, error) {
	db, err := sql.Open("sqlite3", cfg.StoragePath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS students(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT,
	email 	TEXT,
	age INTEGER
	)`)

	if err != nil {
		return nil, err
	}

	return &Sqlite{
		Db: db,
	}, nil

}

func (s *Sqlite) CreateStudent(name string, email string, age int) (int64, error) {
	stmt, err := s.Db.Prepare("INSERT INTO students (name , email , age) VALUES(? , ? , ?)")
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(name, email, age)
	if err != nil {
		return 0, err
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastId, nil

}

func (s *Sqlite) GetStudentById(id int64) (types.Student, error) {
	stmt, err := s.Db.Prepare("SELECT * FROM students WHERE id = ? LIMIT 1")
	if err != nil {
		return types.Student{}, err
	}

	defer stmt.Close()

	var student types.Student

	err = stmt.QueryRow(id).Scan(&student.Id, &student.Name, &student.Email, &student.Age)
	if err != nil {
		if err == sql.ErrNoRows {
			return types.Student{}, fmt.Errorf("no student found with id: %s", fmt.Sprint(id))
		}
		return types.Student{}, fmt.Errorf("query error: %w", err)

	}

	return student, nil
}

func (s *Sqlite) GetStudents() ([]types.Student, error) {
	stmt, err := s.Db.Prepare("SELECT id , name, email ,age FROM students")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var students []types.Student

	for rows.Next() {
		var student types.Student

		err := rows.Scan(&student.Id, &student.Name, &student.Email, &student.Age)

		if err != nil {
			return nil, err
		}

		students = append(students, student)

	}

	return students, nil
}

func (s *Sqlite) UpdateStudent(student types.Student) error {
	// Prepare SQL statement
	query := "UPDATE students SET name = ?, email = ?, age = ? WHERE id = ?"

	stmt, err := s.Db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the update query
	result, err := stmt.Exec(student.Name, student.Email, student.Age, student.Id)
	if err != nil {
		return err
	}

	// Check if any row was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("student with ID %d not found", student.Id)
	}

	return nil
}

func (s *Sqlite) RemoveStudent(id int64) error {
	// Prepare DELETE statement
	stmt, err := s.Db.Prepare("DELETE FROM students WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	defer stmt.Close()

	// Execute DELETE query
	result, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
	}

	// Check if any row was deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no student found with id: %d", id)
	}

	return nil
}
