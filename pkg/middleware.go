package pkg

import (
	"log"
	"net/http"

	"github.com/APIENG/apieng/internal/db"
	"github.com/gorilla/context"
)

func Authorize(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "APIENG" {
			log.Println("Authorized to the system")
			context.Set(r, "user", "ADMIN USER")
			next(w, r)
		} else {
			http.Error(w, "Not Authorized", http.StatusUnauthorized)
		}
	}
}

func AuthorizeCookie(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value == "" {
			if err == http.ErrNoCookie {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Validate session token against database
		dr, err := db.InitializeDB()
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		userID, err := db.ValidateSession(dr, cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		context.Set(r, "user", userID)
		next(w, r)
	}
}
