package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"todolist.go/db"
	"todolist.go/service"
)

const port = 8000

func main() {
	// initialize DB connection
	dsn := db.DefaultDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err := db.Connect(dsn); err != nil {
		log.Fatal(err)
	}

	// initialize Gin engine
	engine := gin.Default()
	engine.Static("/css", "views/css/")
	engine.LoadHTMLGlob("views/*.html")

	// routing
	engine.Static("/assets", "./assets")
	engine.GET("/", service.Home)
	engine.GET("/list", service.TaskList)
	engine.GET("/task/:id", service.ShowTask) // ":id" is a parameter

	//　タスクの新規登録
	engine.GET("/task/new", service.NotImplemented)
	engine.POST("/task/new", service.NotImplemented)

	// 既存のタスクの編集
	engine.GET("/task/edit/:id", service.NotImplemented)
	engine.POST("/task/edit/:id", service.NotImplemented)

	//　既存タスクの削除
	engine.GET("/task/delete/:id", service.NotImplemented)

	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}
