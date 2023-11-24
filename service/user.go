package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	database "todolist.go/db"
)

const userkey = "user"


func LoginForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"})
}

func Login(ctx *gin.Context){
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var user database.User
	err = db.Get(&user, "SELECT id, name, password, is_valid FROM users WHERE name = ?", username)
	if err != nil || !user.IsValid{
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
		return
	}

	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title":"Login", "Username": username, "Error": "Incorrect password"})
		return
	}

	session := sessions.Default(ctx)
	session.Set(userkey, user.ID)
	session.Save()

	ctx.Redirect(http.StatusFound, "/list")
}

func LoginCheck(ctx *gin.Context) {
	if sessions.Default(ctx).Get(userkey) == nil {
		ctx.Redirect(http.StatusFound, "/login")
		ctx.Abort()
	} else{
		ctx.Next()
	}
}

func Logout(ctx *gin.Context){
	session := sessions.Default(ctx)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1})
	session.Save()
	ctx.Redirect(http.StatusFound, "/")
}

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

func checkPassword (password string) string{
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

func RegisterUser (ctx *gin.Context){
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	password_confirm := ctx.PostForm("password_confirm")
	password_check := checkPassword(password)
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
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
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
	//ctx.JSON(http.StatusOK, user)
	ctx.Redirect(http.StatusFound, "/login")
}

func ChangeUserNameAndPasswordForm(ctx *gin.Context){
	userID := sessions.Default(ctx).Get("user")
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var user database.User
	err = db.Get(&user, "SELECT id, name, password, is_valid FROM users WHERE id = ?", userID);
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
	}
	ctx.HTML(http.StatusOK, "change_user_name_and_password.html", gin.H{"Title": "Change user name and password", "Username": user.Name});
}

func ChangeUserNameAndPassword (ctx *gin.Context){
	username := ctx.PostForm("old_username")
	new_username := ctx.PostForm("new_username")
	old_password := ctx.PostForm("old_password")
	new_password := ctx.PostForm("new_password")
	new_password_confirm := ctx.PostForm("new_password_confirm")
	password_check := checkPassword(new_password)
	switch{
		case new_username == "":
			ctx.HTML(http.StatusBadRequest, "change_user_name_and_password.html", gin.H{"Title": "Change user name and password", "Error": "Username is not provided", "Username": username})
			return
		case new_password == "":
			ctx.HTML(http.StatusBadRequest, "change_user_name_and_password.html", gin.H{"Title": "Change user name and password", "Error": "Password is not provided", "Username": username})
			return
		case new_password_confirm == "":
			ctx.HTML(http.StatusBadRequest, "change_user_name_and_password.html", gin.H{"Title": "Change user name and password", "Error": "Password confirm is not provided", "Username": username})
			return
		case new_password != new_password_confirm:
			ctx.HTML(http.StatusBadRequest, "change_user_name_and_password.html", gin.H{"Title": "Change user name and password", "Error": "Passwords do not match", "Username": username})
			return
		case len(new_password) < 8:
			ctx.HTML(http.StatusBadRequest, "change_user_name_and_password.html", gin.H{"Title": "Change user name and password", "Error": "Password is too short", "Username": username})
			return
		case password_check != "":
			ctx.HTML(http.StatusBadRequest, "change_user_name_and_password.html", gin.H{"Title": "Change user name and password", "Error": password_check, "Username": username})
			return
	}

	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// check duplication of new_username
	if username != new_username{
		var duplicate int
		err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", new_username)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return 
		}
		if duplicate > 0 {
			ctx.HTML(http.StatusBadRequest, "change_user_name_and_password.html", gin.H{"Title": "Change user name and password", "Error": "Username is already taken", "Username": username})
			return
		}
	}

	userID := sessions.Default(ctx).Get("user")
	if userID == nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var user database.User
	err = db.Get(&user, "SELECT id, name, password, is_valid FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(old_password)) {
		// password is not correct
		ctx.HTML(http.StatusBadRequest, "change_user_name_and_password.html", gin.H{"Title": "Change user name and password", "Error": "Incorrect Password", "Username": username})
		return
	}

	_, err = db.Exec("UPDATE users SET name=?, password=? WHERE id=?", new_username, hash(new_password), user.ID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.Redirect(http.StatusFound, "/list")
}


func DeleteUser(ctx *gin.Context){
	userID := sessions.Default(ctx).Get("user")
	db, err := database.GetConnection()
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	_, err = db.Exec("UPDATE users SET is_valid=? WHERE id=?", false, userID)
	
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	Logout(ctx)
}