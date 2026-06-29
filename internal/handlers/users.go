package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/APIENG/apieng/internal/db"
	"github.com/APIENG/apieng/internal/models"
	"github.com/APIENG/apieng/internal/services"
)

// usersTemplateData represents the data to be passed to the HTML template.
type UsersTemplateData struct {
	Users []models.Users
}

func FetchUsers(db *sql.DB) ([]models.Users, error) {
	rows, err := db.Query(`SELECT iid, email, firstname, lastname, password, apikey FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usersList []models.Users
	for rows.Next() {
		var m models.Users
		err := rows.Scan(&m.Iid, &m.Email, &m.FirstName, &m.LastName, &m.Password, &m.Apikey)
		if err != nil {
			return nil, err
		}
		usersList = append(usersList, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return usersList, nil
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/signup.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// Handler to display the users Information page
func UsersHandler(w http.ResponseWriter, r *http.Request) {
	dbConn, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer dbConn.Close()

	users, err := FetchUsers(dbConn)
	if err != nil {
		log.Println("error:", err)
		http.Error(w, "Unable to fetch users", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/users.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, UsersTemplateData{Users: users})
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// Handler to Create a new user
func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	firstname := r.FormValue("firstName")
	lastname := r.FormValue("lastName")
	iid := services.GenerateSessionID()
	user := models.Users{
		Email:     email,
		Password:  password,
		Iid:       iid,
		FirstName: firstname,
		LastName:  lastname,
	}
	if email == "" || password == "" {
		http.Error(w, "email and password is required", http.StatusBadRequest)
		return
	}
	err := user.SetPassword(password)
	if err != nil {
		http.Error(w, "password cannot be set", http.StatusBadRequest)
		return
	}

	dr, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}

	err = db.StoreUsers(dr, user)
	if err != nil {
		log.Println("error:", err)
		http.Error(w, "Error storing users", http.StatusInternalServerError)
		return
	}

	// Redirect back to the login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Handle the login of users
func LoginusersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		http.Error(w, "email and password is required to Login", http.StatusBadRequest)
		return
	}
	dbConn, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}

	users, err := FetchUsers(dbConn)
	if err != nil {
		log.Println("error:", err)
		http.Error(w, "Unable to fetch users", http.StatusInternalServerError)
		return
	}
	var foundUser *models.Users
	for _, user := range users {
		if user.Email == email {
			foundUser = &user
			break
		}
	}
	if foundUser == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if !foundUser.CheckPassword(password) {
		http.Error(w, "Unable to login, password does not match", http.StatusUnauthorized)
		return
	}

	sessionID := services.GenerateSessionID()
	expiresAt := time.Now().Add(24 * time.Hour)

	// Store the session in the database
	err = db.StoreSession(dbConn, foundUser.Iid, sessionID, expiresAt)
	if err != nil {
		log.Println("error storing session:", err)
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	// Set the cookie with the session ID
	isSecure := r.URL.Scheme == "https" || r.Header.Get("X-Forwarded-Proto") == "https"
	cookie := http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecure,
		Expires:  expiresAt,
	}
	cookie1 := http.Cookie{
		Name:     "user_id",
		Value:    foundUser.Iid,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecure,
		Expires:  expiresAt,
	}
	http.SetCookie(w, &cookie)
	http.SetCookie(w, &cookie1)
	http.Redirect(w, r, "/metrics", http.StatusSeeOther)
}

// API handler to return current user info (safe data only)
func APIusersHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Unauthorized: No session", http.StatusUnauthorized)
		return
	}

	dbConn, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}

	users, err := FetchUsers(dbConn)
	if err != nil {
		http.Error(w, "Unable to fetch users", http.StatusInternalServerError)
		return
	}

	// Find and return only current user's safe data
	for _, user := range users {
		if user.Iid == cookie.Value {
			safeUser := map[string]string{
				"id":       user.Iid,
				"email":    user.Email,
				"firstName": user.FirstName,
				"lastName": user.LastName,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(safeUser)
			return
		}
	}

	http.Error(w, "User not found", http.StatusNotFound)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		dbConn, err := db.InitializeDB()
		if err == nil {
			db.InvalidateSession(dbConn, cookie.Value)
		}
	}

	cookie = &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
