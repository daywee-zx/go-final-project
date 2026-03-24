package scheduler_db

import (
	"database/sql"
	"strconv"

	_ "modernc.org/sqlite"
)

type SchedulerDB struct {
	DB *sql.DB
}

const (
	schema = `
	CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
	title TEXT NOT NULL DEFAULT "",
	comment TEXT NOT NULL DEFAULT "",
	repeat CHAR(128) NOT NULL DEFAULT ""
	);
	`
)

var (
	ErrNotFound = sql.ErrNoRows
)

func New(db *sql.DB) *SchedulerDB {
	return &SchedulerDB{DB: db}
}

func (db *SchedulerDB) Init() error {
	_, err := db.DB.Exec(schema)
	if err != nil {
		return err
	}

	return nil
}

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func (db *SchedulerDB) AddTask(t Task) (int64, error) {
	res, err := db.DB.Exec("INSERT INTO scheduler (date, title, comment, repeat) "+
		"VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", t.Date),
		sql.Named("title", t.Title),
		sql.Named("comment", t.Comment),
		sql.Named("repeat", t.Repeat),
	)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (db *SchedulerDB) GetTask(id string) (*Task, error) {
	var task Task
	var taskID int64

	row := db.DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id",
		sql.Named("id", id))
	err := row.Scan(&taskID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return nil, err
	}
	task.ID = strconv.FormatInt(taskID, 10)

	return &task, nil
}

func (db *SchedulerDB) GetTasks(limit int) ([]*Task, error) {
	rows, err := db.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler "+
		"ORDER BY date LIMIT :limit", sql.Named("limit", limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var tasks []*Task = make([]*Task, 0, limit)
	for rows.Next() {
		var task Task
		var taskID int64
		if err := rows.Scan(&taskID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		task.ID = strconv.FormatInt(taskID, 10)
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (db *SchedulerDB) UpdateTask(t Task) error {
	res, err := db.DB.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
		sql.Named("date", t.Date),
		sql.Named("title", t.Title),
		sql.Named("comment", t.Comment),
		sql.Named("repeat", t.Repeat),
		sql.Named("id", t.ID))

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return err
}

func (db *SchedulerDB) DeleteTask(id string) error {
	res, err := db.DB.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", id))
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
