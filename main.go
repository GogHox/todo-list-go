package main

import (
	"github.com/google/uuid"
	"time"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"task-manager-go/handler/task_handler"
	"github.com/gin-gonic/gin"
)

//var KEEPIE_SERVER="http://localhost:8000" // for local
var KEEPIE_SERVER="http://107.173.104.196:8000" // for internet

var CURREN_SERVER="localhost:8081"

var Session = make(map[string]int)

func main() {
	router := gin.New()
	
	router.POST("/givemesecret", func(context *gin.Context) {
		// store the secret to memory
		
		var body map[string]interface{}
		err := context.BindJSON(&body)
		if err != nil {
			return
		}
		
		token := context.Query("token")
		
		if body["username"].(string) == context.Query("username") &&
			body["password"].(string) == context.Query("password") {
			Session[token] = 1
			context.IndentedJSON(http.StatusOK, gin.H{})
		} else {
			Session[token] = 2
			context.IndentedJSON(http.StatusUnauthorized, gin.H{})
		}
		
	})
	authorized := router.Group("/")
	
	// Authorized
	authorized.Use(func(context *gin.Context) {
		iUsername := context.GetHeader("username")
		iPassword := context.GetHeader("Password")
		if len(iUsername) == 0 || len(iPassword) == 0 {
			context.JSON(http.StatusUnauthorized, gin.H{"msg": "pls input username and password"})
			context.Abort()
			return
		} 
		uuid := uuid.New().String()
		Session[uuid] = 0
		defer delete(Session, uuid)
		var RetryTime = 10
		context.Writer.Header().Set("X-Request-ID", uuid)
		for  i := 0; i < RetryTime; i++ {
			bytesData, _ := json.Marshal(map[string]string{"receive_url": fmt.Sprintf(
				"http://%s/givemesecret?token=%s&username=%s&password=%s", CURREN_SERVER, uuid, iUsername, iPassword)})
			resp, _ := http.Post(KEEPIE_SERVER + "/sendSecretToMe","application/json", bytes.NewReader(bytesData))
			
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(fmt.Sprintf("calling api to get credencial(%d): %s", i, string(body)))
			defer resp.Body.Close()

			switch Session[uuid] {
			case 1:  // pass
				i = RetryTime
				context.Set("username", iUsername)
				context.Next()
			case 2:  // fail
				i = RetryTime
				context.JSON(http.StatusUnauthorized, gin.H{"msg": "Unauthorized, pls check your username&password"})
				context.Abort()
			default: // retry
				time.Sleep(time.Duration(1) * time.Second)
			}
		}
		if Session[uuid] == 0 {
			context.JSON(http.StatusUnauthorized, gin.H{"msg": "Unauthorized after retry 10 times"})
			context.Abort()
		}
	})
	
	
	// router
	authorized.POST("/", task_handler.Index)
	authorized.POST("/add_task", task_handler.AddTask)
	authorized.POST("/list_task", task_handler.ListTask)
	authorized.POST("/remove_task", task_handler.RemoveTask)
	authorized.POST("/modify_task", task_handler.ModifyTask)
	authorized.POST("/get_task", task_handler.GetTask)
	

	router.Run(CURREN_SERVER)
}
