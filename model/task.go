package model

type Task struct {
	Name string `json:"name"`
	Completed bool `json:"completed"`
}

type TaskResponse struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Completed bool `json:"completed"`
}