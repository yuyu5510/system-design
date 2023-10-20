package service

import (
	"crypto/sha256"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

func NewUserForm(ctx *gin.Context){
	ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Rgister user"})
}

func hash(pw string) []byte {
	const salt = "todolist.go#"
	h := sha256.New()
	h.Write([] byte(salt))
	h.Write([] byte(pw))
	return h.Sum(nil)
}

func CheckPassword (password string) string{
	ret := ""
	re_num := regexp.MustCompile(`\d+`)
	re_small := regexp.MustCompile(`[a-z]+`)
	re_big := regexp.MustCompile(`[A-Z]+`)
	if !re_num.MatchString(password){
		ret += "Password must include at least one number.\n"
	}
	if !re_small.MatchString(password){
		ret += "Password must include at least one small letter.\n"
	}
	if !re_big.MatchString(password){
		ret += "Password must include at least one big letter.\n"
	}
	return ret
}

func RegisetrUser (ctx *gin.Context){
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	password_confirm := ctx.PostForm("password_confirm")
	password_check := CheckPassword(password)
	switch{
		case username == "":
			ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is not provided", "Password": password})
			return
		case password == "":
			ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Username": username})
			return
		case password_confirm == "":
			ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password confirm is not provided", "Username": username, "Password": password})
			return
		case password != password_confirm:
			ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Passwords do not match", "Username": username})
			return
		case len(password) < 8:
			ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is too short", "Username": username})
			return
		case password_check != "":
			ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": password_check, "Username": username})
			return
	}


	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name = ?", username)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return 
	}
	if duplicate > 0 {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password})
		return
	}

	result, err := db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	id, _ := result.LastInsertId()
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.JSON(http.StatusOK, user)
}