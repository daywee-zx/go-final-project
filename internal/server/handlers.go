package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"go_final/internal/scheduler_db"
	"net/http"
	"time"
)

const taskLimit = 25

type Response struct {
	ID    int64  `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
	Token string `json:"token,omitempty"`
}

func (s *Server) NextDayHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dstart := r.FormValue("date")
	repeat := r.FormValue("repeat")

	now, err := time.Parse(timeFormat, nowStr)
	if err != nil {
		http.Error(w, "Invalid 'now' date format", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dstart, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = w.Write([]byte(nextDate))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	var Task scheduler_db.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &Task); err != nil {
		respondJSON(w, Response{Error: "JSON parsing error: " + err.Error()})
		return
	}

	if Task.Title == "" {
		respondJSON(w, Response{Error: "Title is required"})
		return
	}
	if err = checkDate(&Task); err != nil {
		respondJSON(w, Response{Error: "Date error: " + err.Error()})
		return
	}

	id, err := s.db.AddTask(Task)
	if err != nil {
		respondJSON(w, Response{Error: "Internal error"})
		s.Server.ErrorLog.Println("Failed to add task:", err)
		return
	}

	respondJSON(w, Response{ID: id})
}

func (s *Server) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		respondJSON(w, Response{Error: "ID is required"})
		return
	}

	task, err := s.db.GetTask(id)
	if err != nil {
		if err == scheduler_db.ErrNotFound {
			respondJSON(w, Response{Error: "Task not found"})
			return
		}
		respondJSON(w, Response{Error: "Internal error"})
		s.Server.ErrorLog.Println("Failed to get task:", err)
		return
	}

	respondJSON(w, task)
}

type TasksResponse struct {
	Tasks []*scheduler_db.Task `json:"tasks"`
}

func (s *Server) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.db.GetTasks(taskLimit)
	if err != nil {
		respondJSON(w, Response{Error: "Internal error"})
		s.Server.ErrorLog.Println("Failed to get tasks:", err)
		return
	}
	respondJSON(w, TasksResponse{Tasks: tasks})
}

func (s *Server) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var Task scheduler_db.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &Task); err != nil {
		respondJSON(w, Response{Error: "JSON parsing error: " + err.Error()})
		return
	}

	if Task.ID == "" {
		respondJSON(w, Response{Error: "ID is required"})
		return
	}
	if Task.Title == "" {
		respondJSON(w, Response{Error: "Title is required"})
		return
	}
	err = checkDate(&Task)
	if err != nil {
		respondJSON(w, Response{Error: "Date error: " + err.Error()})
		return
	}

	err = s.db.UpdateTask(Task)
	if err != nil {
		respondJSON(w, Response{Error: "Internal error"})
		s.Server.ErrorLog.Println("Failed to update task:", err)
		return
	}

	respondJSON(w, Response{})
}

func (s *Server) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		respondJSON(w, Response{Error: "ID is required"})
		return
	}

	err := s.db.DeleteTask(id)
	if err != nil {
		if err == scheduler_db.ErrNotFound {
			respondJSON(w, Response{Error: "Task not found"})
			return
		}
		respondJSON(w, Response{Error: "Internal error"})
		s.Server.ErrorLog.Println("Failed to delete task:", err)
		return
	}

	respondJSON(w, Response{})
}

func (s *Server) MarkDoneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		respondJSON(w, Response{Error: "ID is required"})
		return
	}

	task, err := s.db.GetTask(id)
	if err != nil {
		if err == scheduler_db.ErrNotFound {
			respondJSON(w, Response{Error: "Task not found"})
			return
		}
		respondJSON(w, Response{Error: "Internal error"})
		s.Server.ErrorLog.Println("Failed to get task:", err)
		return
	}

	if task.Repeat == "" {
		err = s.db.DeleteTask(id)
		if err != nil {
			respondJSON(w, Response{Error: "Internal error"})
			s.Server.ErrorLog.Println("Failed to delete task:", err)
			return
		}
		respondJSON(w, Response{})
		return
	}

	nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		respondJSON(w, Response{Error: "Date error: " + err.Error()})
		return
	}

	task.Date = nextDate
	err = s.db.UpdateTask(*task)
	if err != nil {
		respondJSON(w, Response{Error: "Internal error"})
		s.Server.ErrorLog.Println("Failed to update task:", err)
		return
	}

	respondJSON(w, Response{})
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var res map[string]string
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &res); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		s.Server.ErrorLog.Println("Failed to parse login request:", err)
		return
	}

	password := res["password"]

	hash := sha256.Sum256([]byte(password))
	if hex.EncodeToString(hash[:]) != s.password {
		respondJSON(w, Response{Error: "Wrong password"})
		return
	}

	token := generateToken(s.password)
	respondJSON(w, Response{Token: token})
}
