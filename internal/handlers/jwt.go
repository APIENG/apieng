package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/APIENG/apieng/internal/db"
	"github.com/APIENG/apieng/internal/models"
	jwt "github.com/dgrijalva/jwt-go"
)

// using asymmetric crypto/RSA keys
// location of the files used for signing and verification
const (
	privKeyPath = "keys/app.rsa"
	pubKeyPath  = "keys/app.rsa.pub" // openssl rsa -in app.rsa -pubout > app.rsa.pub
)

// verify key and sign key
var (
	verifyKey, signKey []byte
)

type Response struct {
	Text string `json:"text"`
}
type Token struct {
	Token string `json:"token"`
}

type CustomClaims struct {
	UserInfo models.Users `json:"CustomUserInfo"`
	jwt.StandardClaims
}

//var secretKey = []byte("your-secret-key")

func FetchEachUser(db *sql.DB, iid string) (models.Users, error) {
	// Prepare the SQL query with a placeholder for the iid
	query := `SELECT iid, email, password FROM users WHERE iid LIKE ?`

	// Execute the query with the iid
	row := db.QueryRow(query, iid)

	var user models.Users
	err := row.Scan(&user.Iid, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			// No user found with the given iid
			return models.Users{}, fmt.Errorf("no user found with iid %d", iid)
		}
		// Other errors
		return models.Users{}, err
	}

	return user, nil
}

// read the key files before starting http handlers
func init() {
	var err error
	signKey, err = ioutil.ReadFile(privKeyPath)
	if err != nil {
		log.Fatal("Error reading private key")
		return
	}
	verifyKey, err = ioutil.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatal("Error reading private key")
		return
	}
}

func createToken(userInfo models.Users) (string, error) {
	claims := CustomClaims{
		UserInfo: userInfo,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 100).Unix(), // Set expiration time
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// reads the login credentials, checks them and creates JWT the token
func GenerateKey(w http.ResponseWriter, r *http.Request) {
	db, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	cookie, err := r.Cookie("user_id")
	if err != nil || cookie.Value == "" {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			//http.Error(w, "Unauthorized: No session token", http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	//decode into User struct
	users, err := FetchEachUser(db, cookie.Value)
	if err != nil {
		http.Error(w, "Unable to fetch users", http.StatusInternalServerError)
		return
	}
	tokenString, err := createToken(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Sorry, error while Signing Token!")
		log.Printf("Token Signing error: %v\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(tokenString)

	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}

// only accessible with a valid token
func jsonResponse(response Response, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func AuthorizeAPI(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization") // Assume token is sent in the Authorization header
		log.Printf("%s", tokenString)

		if tokenString == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the token method and return the key
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return verifyKey, nil
		})

		if err != nil {
			if vErr, ok := err.(*jwt.ValidationError); ok {
				switch vErr.Errors {
				case jwt.ValidationErrorExpired:
					w.WriteHeader(http.StatusUnauthorized)
					fmt.Fprintln(w, "Token Expired, get a new one.")
				case jwt.ValidationErrorSignatureInvalid:
					w.WriteHeader(http.StatusUnauthorized)
					fmt.Fprintln(w, "Invalid token signature.")
				case jwt.ValidationErrorMalformed:
					w.WriteHeader(http.StatusUnauthorized)
					fmt.Fprintln(w, "Malformed token.")
				case jwt.ValidationErrorUnverifiable:
					w.WriteHeader(http.StatusUnauthorized)
					fmt.Fprintln(w, "Token could not be verified.")
				default:
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintln(w, "Unknown token validation error.")
				}
				log.Printf("ValidationError error: %+v\n", vErr.Errors)
				return
			}
			// General error handling for non-ValidationError cases
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error while Parsing Token!")
			log.Printf("Token parse error: %v\n", err)
			return
		}

		if token.Valid {
			next(w, r)
		} else {
			response := Response{"Invalid token"}
			jsonResponse(response, w)
		}
	}
}
