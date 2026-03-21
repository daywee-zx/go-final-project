package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go_final/internal/scheduler_db"
	"net/http"
	"time"
)

const timeFormat = "20060102"

// checks the date to correspond to the repeat pattern and not to be in the past, updates it if needed
func checkDate(task *scheduler_db.Task) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if task.Date == "" {
		task.Date = today.Format(timeFormat)
		return nil
	}

	t, err := time.Parse(timeFormat, task.Date)
	if err != nil {
		return err
	}

	next := today.Format(timeFormat)
	if task.Repeat != "" {
		next, err = NextDate(today, task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	if t.Before(today) {
		task.Date = next
	}
	return nil
}

func respondJSON(w http.ResponseWriter, item any) {
	itemJSON, err := json.Marshal(item)
	if err != nil {
		fmt.Println("Failed to marshal JSON:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	fmt.Fprint(w, string(itemJSON))
}

// very simple token generation, updates every day
func generateToken(password string) string {
	now := time.Now().Format(timeFormat)
	tokenData := password + now
	tokenHash := sha256.Sum256([]byte(tokenData))
	return hex.EncodeToString(tokenHash[:])
}
