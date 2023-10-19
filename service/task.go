package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get tasks in DB
	var tasks []database.Task
	err = db.Select(&tasks, "SELECT * FROM tasks") // Use DB#Select for multiple entries
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// parse ID given as a parameter
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get a task with given ID
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Render task
	ctx.HTML(http.StatusOK, "task.html", task)
}

func NewTaskForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration"})
}

func RegisterTask(ctx *gin.Context){
	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
		Error(http.StatusBadRequest, "title is not given")(ctx)
		return
	}
	
	description, exist := ctx.GetPostForm("description")
	if !exist {
		Error(http.StatusBadRequest, "description is not given")(ctx)
		return
	}

	db, err := database.GetConnection()
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	result, err := db.Exec("INSERT INTO tasks (title, description) VALUES (?, ?)", title, description)
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	path := "/list"
	if id, err := result.LastInsertId(); err == nil {
		path = fmt.Sprintf("/task/%d", id)
	}
	ctx.Redirect(http.StatusFound, path)
}

func EditTaskForm(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil{
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	db, err := database.GetConnection()
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	ctx.HTML(http.StatusOK, "form_edit_task.html", 
			gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task})
}

func UpdateTask(ctx *gin.Context){
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil{
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
		Error(http.StatusBadRequest, "title is not given")(ctx)
		return
	}
	
	description, exist := ctx.GetPostForm("description")
	if !exist {
		Error(http.StatusBadRequest, "description is not given")(ctx)
		return
	}
	
	is_done_str, exist := ctx.GetPostForm("is_done")
	if !exist {
		Error(http.StatusBadRequest, "is_done is not given")(ctx)
		return
	}

	is_done, err := strconv.ParseBool(is_done_str)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	

	db, err := database.GetConnection()
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	_, err = db.Exec("UPDATE tasks SET title=?, is_done=?, description=? WHERE id=?", title, is_done, description, id)
	
	
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	path := "/list"
	if err == nil {
		path = fmt.Sprintf("/task/%d", id)
	}
	ctx.Redirect(http.StatusFound, path)
}

func DeleteTask(ctx *gin.Context){
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil{
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	_, err = db.Exec("DELETE FROM tasks WHERE id=?", id)

	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.Redirect(http.StatusFound, "/list")

}