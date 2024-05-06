package task_handler

import (
	"net/http"
	"fmt"
	"task-manager-go/model"

	"task-manager-go/mocks/task_db_mock"

	"github.com/gin-gonic/gin"
)


func Index(context *gin.Context) {
//	t1 := model.Task{Name: "t1"}
//	context.Query("name")

	context.IndentedJSON(http.StatusOK, gin.H{})
}

func ListTask(context *gin.Context) {
	username := context.GetString("username")
	taskStore, err := task_db_mock.NewClient()
	if err != nil {
		context.IndentedJSON(http.StatusServiceUnavailable, gin.H{})
		return
	}

	res, err := taskStore.ListTasks(username)
	if err != nil {
			context.IndentedJSON(http.StatusServiceUnavailable, gin.H{})
			return
		}
	context.IndentedJSON(http.StatusOK, res)
}

func AddTask(context *gin.Context) {
	username := context.GetString("username")
	taskStore, err := task_db_mock.NewClient()
	if err != nil {
		context.IndentedJSON(http.StatusServiceUnavailable, gin.H{"msg": err.Error()})
		return
	}
	var task model.Task
	if err := context.BindJSON(&task); err != nil {
		context.JSON(http.StatusForbidden, gin.H{"msg": "pls input correct value"})
		return 
	}
	if len(task.Name) == 0 {
		context.JSON(http.StatusServiceUnavailable, gin.H{"msg": "pls input correct name"})
		return 
	}
	res, err := taskStore.AddTask(&task, username)
	if err != nil {
		fmt.Println(err.Error())
		context.IndentedJSON(http.StatusServiceUnavailable, gin.H{"msg": err.Error()})
		return
	}
	context.IndentedJSON(http.StatusOK, gin.H{"success": res})
		
}

func RemoveTask(context *gin.Context) {
	username := context.GetString("username")
	taskStore, err := task_db_mock.NewClient()
	if err != nil {
		context.IndentedJSON(http.StatusServiceUnavailable, gin.H{"msg": err.Error()})
		return
	}
	var task model.Task
	if err := context.BindJSON(&task); err != nil {
		context.JSON(http.StatusForbidden, gin.H{"msg": "pls input correct value"})
		return 
	}
	
	if len(task.Name) == 0 {
		context.JSON(http.StatusServiceUnavailable, gin.H{"msg": "pls input correct name"})
		return 
	}
	res, err := taskStore.RemoveTask(&task, username)
	if err != nil {
		context.IndentedJSON(http.StatusServiceUnavailable, gin.H{"msg": err.Error()})
		return 
	}
	context.IndentedJSON(http.StatusOK, gin.H{"success": res})
}


func ModifyTask(context *gin.Context) {
	username := context.GetString("username")
	taskStore, err := task_db_mock.NewClient()
	if err != nil {
		context.JSON(http.StatusServiceUnavailable, gin.H{"msg": err.Error()})
		return
	}
	var task model.TaskResponse
	if err := context.BindJSON(&task); err != nil {
		context.JSON(http.StatusForbidden, gin.H{"msg": "pls input correct value"})
		return 
	}
	
	if len(task.Name) == 0 {
		context.JSON(http.StatusServiceUnavailable, gin.H{"msg": "pls input correct name"})
		return 
	}
	res, err := taskStore.ModifyTask(&model.Task{Name: task.Name, Completed: task.Completed}, task.Id, username)
	if err != nil {
		context.JSON(http.StatusServiceUnavailable, gin.H{"msg": err.Error()})
		return 
	}
	context.JSON(http.StatusOK, gin.H{"success": res})
}

func GetTask(context *gin.Context) {
	username := context.GetString("username")
	var task model.Task
	if err := context.BindJSON(&task); err != nil {
		context.JSON(http.StatusForbidden, gin.H{"msg": "pls input correct value"})
		return 
	}
	
	taskStore, err := task_db_mock.NewClient()
	if err != nil {
		context.IndentedJSON(http.StatusServiceUnavailable, gin.H{"msg": err.Error()})
		return
	}

	res, err := taskStore.ListTasks(username)
	if err != nil {
		context.IndentedJSON(http.StatusServiceUnavailable, gin.H{"msg": err.Error()})
		return
	}
	for _, v := range res {
		if v.Name == task.Name {
			context.IndentedJSON(http.StatusOK, v)
			return
		}
	}
	context.IndentedJSON(http.StatusOK, gin.H{})
}