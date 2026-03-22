package server

import (
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
	jsonParser := json.NewDecoder(r.Body)
	if err := jsonParser.Decode(&Task); err != nil {
		respondJSON(w, Response{Error: "JSON parsing error: " + err.Error()}, http.StatusBadRequest)
		return
	}

	if Task.Title == "" {
		respondJSON(w, Response{Error: "Title is required"}, http.StatusBadRequest)
		return
	}
	if err := checkDate(&Task); err != nil {
		respondJSON(w, Response{Error: "Date error: " + err.Error()}, http.StatusBadRequest)
		return
	}

	id, err := s.db.AddTask(Task)
	if err != nil {
		respondJSON(w, Response{Error: "Internal error"}, http.StatusInternalServerError)
		s.Server.ErrorLog.Println("Failed to add task:", err)
		return
	}

	respondJSON(w, Response{ID: id}, http.StatusOK)
}

func (s *Server) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		respondJSON(w, Response{Error: "ID is required"}, http.StatusBadRequest)
		return
	}

	task, err := s.db.GetTask(id)
	if err != nil {
		if err == scheduler_db.ErrNotFound {
			respondJSON(w, Response{Error: "Task not found"}, http.StatusNotFound)
			return
		}
		respondJSON(w, Response{Error: "Internal error"}, http.StatusInternalServerError)
		s.Server.ErrorLog.Println("Failed to get task:", err)
		return
	}

	respondJSON(w, task, http.StatusOK)
}

type TasksResponse struct {
	Tasks []*scheduler_db.Task `json:"tasks"`
}

func (s *Server) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.db.GetTasks(taskLimit)
	if err != nil {
		respondJSON(w, Response{Error: "Internal error"}, http.StatusInternalServerError)
		s.Server.ErrorLog.Println("Failed to get tasks:", err)
		return
	}
	respondJSON(w, TasksResponse{Tasks: tasks}, http.StatusOK)
}

func (s *Server) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var Task scheduler_db.Task
	jsonParser := json.NewDecoder(r.Body)
	if err := jsonParser.Decode(&Task); err != nil {
		respondJSON(w, Response{Error: "JSON parsing error: " + err.Error()}, http.StatusBadRequest)
		return
	}

	if Task.ID == "" {
		respondJSON(w, Response{Error: "ID is required"}, http.StatusBadRequest)
		return
	}
	if Task.Title == "" {
		respondJSON(w, Response{Error: "Title is required"}, http.StatusBadRequest)
		return
	}
	err := checkDate(&Task)
	if err != nil {
		respondJSON(w, Response{Error: "Date error: " + err.Error()}, http.StatusBadRequest)
		return
	}

	err = s.db.UpdateTask(Task)
	if err != nil {
		respondJSON(w, Response{Error: "Internal error"}, http.StatusInternalServerError)
		s.Server.ErrorLog.Println("Failed to update task:", err)
		return
	}

	respondJSON(w, Response{}, http.StatusOK)
}

func (s *Server) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		respondJSON(w, Response{Error: "ID is required"}, http.StatusBadRequest)
		return
	}

	err := s.db.DeleteTask(id)
	if err != nil {
		if err == scheduler_db.ErrNotFound {
			respondJSON(w, Response{Error: "Task not found"}, http.StatusNotFound)
			return
		}
		respondJSON(w, Response{Error: "Internal error"}, http.StatusInternalServerError)
		s.Server.ErrorLog.Println("Failed to delete task:", err)
		return
	}

	respondJSON(w, Response{}, http.StatusOK)
}

func (s *Server) MarkDoneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		respondJSON(w, Response{Error: "ID is required"}, http.StatusBadRequest)
		return
	}

	task, err := s.db.GetTask(id)
	if err != nil {
		if err == scheduler_db.ErrNotFound {
			respondJSON(w, Response{Error: "Task not found"}, http.StatusNotFound)
			return
		}
		respondJSON(w, Response{Error: "Internal error"}, http.StatusInternalServerError)
		s.Server.ErrorLog.Println("Failed to get task:", err)
		return
	}

	if task.Repeat == "" {
		err = s.db.DeleteTask(id)
		if err != nil {
			respondJSON(w, Response{Error: "Internal error"}, http.StatusInternalServerError)
			s.Server.ErrorLog.Println("Failed to delete task:", err)
			return
		}
		respondJSON(w, Response{}, http.StatusOK)
		return
	}

	nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		respondJSON(w, Response{Error: "Date error: " + err.Error()}, http.StatusBadRequest)
		return
	}

	task.Date = nextDate
	err = s.db.UpdateTask(*task)
	if err != nil {
		respondJSON(w, Response{Error: "Internal error"}, http.StatusInternalServerError)
		s.Server.ErrorLog.Println("Failed to update task:", err)
		return
	}

	respondJSON(w, Response{}, http.StatusOK)
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var res map[string]string
	jsonParser := json.NewDecoder(r.Body)
	if err := jsonParser.Decode(&res); err != nil {
		respondJSON(w, Response{Error: "JSON parsing error: " + err.Error()}, http.StatusBadRequest)
		return
	}

	password := res["password"]

	hash := sha256.Sum256([]byte(password))
	if hex.EncodeToString(hash[:]) != s.password {
		respondJSON(w, Response{Error: "Wrong password"}, http.StatusUnauthorized)
		return
	}

	token := generateToken(s.password)
	respondJSON(w, Response{Token: token}, http.StatusOK)
}
