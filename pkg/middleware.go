package pkg

import (
	"log"
	"net/http"

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
		if err != nil {
			if err == http.ErrNoCookie {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				//http.Error(w, "Unauthorized: No session token", http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		context.Set(r, "token", cookie)
		next(w, r)
	}
}
