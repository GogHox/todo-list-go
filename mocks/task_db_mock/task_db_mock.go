package task_db_mock

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"task-manager-go/model"

	_ "github.com/mattn/go-sqlite3"
)

type TaskStore interface {
	AddTask(task *model.Task)
	RemoveTask(task *model.Task)
	ModifyTask(task *model.Task, newTask *model.Task)
	ListTasks(user string) ([]*model.Task, error)
}

type TaskStoreBySqlite3 struct {
	sqllite_file string
}

var db_init_sql = `CREATE TABLE task(
		"task_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		"name" VARCHAR(255) NOT NULL,
		"username" VARCHAR(255) NOT NULL,
		"completed" boolean NOT NULL
	);`

func (_m TaskStoreBySqlite3) InitSqlite3(db_path string) error {
	fmt.Println("Initialing DB")
	db_file, err := os.Create(db_path)
	if err != nil {
		return fmt.Errorf("error in create file: " + db_path)
	}
	db_file.Close()
	db, err := sql.Open("sqlite3", db_path)
	if err != nil {
		return fmt.Errorf("error in init sqlite db: " + db_path)
	}
	defer db.Close()

	_, err = db.Exec(db_init_sql)
	if err != nil {
		return fmt.Errorf("error in init sqlite db, failed to create table: " + err.Error())
	}
	return nil
}

func NewClient() (*TaskStoreBySqlite3, error) {
	_m := TaskStoreBySqlite3{sqllite_file: "./task.db"}
	if _, err := os.Stat(_m.sqllite_file); errors.Is(err, os.ErrNotExist) {
		// can't find the db file, init one
		err = _m.InitSqlite3(_m.sqllite_file)
		if err != nil {
			return &TaskStoreBySqlite3{}, err
		}
	}
	return &_m, nil
}

func (_m *TaskStoreBySqlite3) GetSqliteClient() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", _m.sqllite_file)
	if err != nil {
		return nil, fmt.Errorf("error in getting sqlite3 client: ")
	}
	return db, nil
}

func (_m *TaskStoreBySqlite3) AddTask(task *model.Task, user string) (bool, error) {
	if len(user) == 0 {
		return false, fmt.Errorf("failed to add task due to missing username")
	}

	db, err := _m.GetSqliteClient()
	if err != nil {
		return false, err
	}
	defer db.Close()
	
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	
	// in case any error, will rollback the transaction
	defer tx.Rollback()
	
	// check task name if existing
	var length int
	if err := tx.QueryRow("SELECT count(*) FROM task WHERE name=? AND username=?",task.Name, user).Scan(&length); err != nil {
		return false, fmt.Errorf("failed to add task: " + err.Error())
	}
	if length != 0 {
		return false, fmt.Errorf("failed to add task due to there is task with name: " + task.Name)
	}

	// insert task to db
	_, err = tx.Exec("INSERT INTO task(name, username, completed) VALUES(?, ?, ?)", task.Name, user, task.Completed)
	if err != nil {
		return false, fmt.Errorf("error in insert task, id: " + user)
	}
	
	if tx.Commit() != nil {
		return false, err
	}

	return true, nil
}

func (_m *TaskStoreBySqlite3) RemoveTask(task *model.Task, user string) (bool, error) {
	if len(user) == 0 {
		return false, fmt.Errorf("failed to add task due to missing username")
	}

	db, err := _m.GetSqliteClient()
	if err != nil {
		return false, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	
	fmt.Println("checking task")
	// check task name if existing
	var id int
	if err := tx.QueryRow("SELECT task_id FROM task WHERE name=? AND username=?", task.Name, user).Scan(&id); err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, fmt.Errorf("there is not this task, name: " + task.Name)
		}
		return false, fmt.Errorf("error in query task, user: " + user + ", task name: " + task.Name + ", " + err.Error())
	}
	// delete task from db
	if _, err := tx.Exec("DELETE FROM task WHERE task_id=?", id); err != nil {
		return false, fmt.Errorf("error in delete task, user: " + user + ", " + err.Error())
	}
	
	if tx.Commit() != nil {
		return false, err
	}

	return true, nil
}

func (_m *TaskStoreBySqlite3) ModifyTask(newTask *model.Task, id int, username string) (bool, error) {

	db, err := _m.GetSqliteClient()
	if err != nil {
		return false, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	// check task name if existing
	var taskName string
	if err := tx.QueryRow("SELECT name FROM task WHERE username=? AND task_id=? ", username, id).Scan(&taskName); err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, fmt.Errorf("NOT task found")
		}
		return false, fmt.Errorf(fmt.Sprintf("error in query task, id: %d, task name: %s, %s", id, newTask.Name, err.Error()))
	}
	
	if _, err := tx.Exec("Update task SET name=?, completed=? WHERE task_id=?", newTask.Name, newTask.Completed, id); err != nil {
		return false, fmt.Errorf("error in update task")
	}
	
	if tx.Commit() != nil {
		return false, err
	}
	
	return true, nil
}

func (_m *TaskStoreBySqlite3) ListTasks(user string) ([]*model.TaskResponse, error) {
	if len(user) == 0 {
		return []*model.TaskResponse{}, fmt.Errorf("failed to add task due to missing username")
	}

	db, err := _m.GetSqliteClient()
	if err != nil {
		return []*model.TaskResponse{}, err
	}
	defer db.Close()

	// get list
	row, err := db.Query(fmt.Sprintf("SELECT task_id, name, username, completed FROM task WHERE username='%s'", user))
	if err != nil {
		return []*model.TaskResponse{}, fmt.Errorf("error in list task, id: " + user)
	}

	var res = []*model.TaskResponse{}
	for row.Next() {
		var id int
		var taskName string
		var username string
		var completed bool
		row.Scan(&id, &taskName, &username, &completed)
		res = append(res, &model.TaskResponse{id, taskName, completed})
	}

	return res, nil
}
