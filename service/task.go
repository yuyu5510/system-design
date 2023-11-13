package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	database "todolist.go/db"
)

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get query parameter
	kw := ctx.Query("kw")
	is_done := ctx.Query("is_done")
	is_not_done := ctx.Query("is_not_done")

	// Get tasks in DB
	var tasks []database.Task
	query := "SELECT id, title, created_at, is_done FROM tasks INNER JOIN ownership ON tasks.id = ownership.task_id WHERE user_id = ? "
	switch{
		// チェックボックスに両方つけるまたは両方つけない時タイトル検索だけ行う
		case kw != "" && ((is_done != "" && is_not_done != "") || (is_done == "" && is_not_done == "")):
			err = db.Select(&tasks, query + "AND title LIKE ?", userID, "%" + kw + "%")
		case kw != "" && (is_done != "" && is_not_done == ""):
			err = db.Select(&tasks, query + "AND title LIKE ? AND is_done=?", userID, "%" + kw + "%", true)
		case kw != "" && (is_done == "" && is_not_done != ""):
			err = db.Select(&tasks, query + "AND title LIKE ? AND is_done=?", userID, "%" + kw + "%", false)
		case kw == "" && (is_done != "" && is_not_done == ""):
			err = db.Select(&tasks, query + "AND is_done=?", userID, true)
		case kw == "" && (is_done == "" && is_not_done != ""):
			err = db.Select(&tasks, query + "AND is_done=?", userID, false)
		default:
			err = db.Select(&tasks, query, userID)
	}	
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks, "Kw": kw, "IsDone": is_done, "IsNotDone": is_not_done, "UserID": userID})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	// Get DB connection
	userID := sessions.Default(ctx).Get("user")

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

	err = db.Get(&task, "SELECT id, title, created_at, description, is_done FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id=? and id=?", userID, id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Render task
	ctx.HTML(http.StatusOK, "task.html", gin.H{"Tasks": task, "UserID": userID})
}

func NewTaskForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration"})
}

func RegisterTask(ctx *gin.Context){
	userID := sessions.Default(ctx).Get("user")
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
	tx := db.MustBegin()
	result, err := db.Exec("INSERT INTO tasks (title, description) VALUES (?, ?)", title, description)
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	
	taskID, err := result.LastInsertId()
	if err != nil{
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	_, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userID, taskID)
	if err != nil{
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx.Commit()
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
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