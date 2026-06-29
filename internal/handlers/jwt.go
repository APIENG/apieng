package handlers

import (
	"crypto/rsa"
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
	"github.com/gorilla/context"
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

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
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


func FetchEachUser(db *sql.DB, iid string) (models.Users, error) {
	// Prepare the SQL query with a placeholder for the iid
	query := `SELECT iid, email, password, apikey FROM users WHERE iid LIKE ?`

	// Execute the query with the iid
	row := db.QueryRow(query, iid)

	var user models.Users
	err := row.Scan(&user.Iid, &user.Email, &user.Password, &user.Apikey)
	if err != nil {
		if err == sql.ErrNoRows {
			// No user found with the given iid
			return models.Users{}, fmt.Errorf("no user found with iid %s", iid)
		}
		// Other errors
		return models.Users{}, err
	}

	return user, nil
}

func FetchEachUserByToken(db *sql.DB, token string) (models.Users, error) {
	// Prepare the SQL query with a placeholder for the iid
	query := `SELECT iid, email, password, apikey FROM users WHERE apikey LIKE ?`

	// Execute the query with the iid
	row := db.QueryRow(query, token)

	var user models.Users
	err := row.Scan(&user.Iid, &user.Email, &user.Password, &user.Apikey)
	if err != nil {
		if err == sql.ErrNoRows {
			// No user found with the given iid
			return models.Users{}, fmt.Errorf("no user found with iid %s", token)
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
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(signKey)
	if err != nil {
		log.Fatalf("Error parsing private key: %v", err)
		return
	}
	verifyKey, err = ioutil.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatal("Error reading private key")
		return
	}
	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyKey)
	if err != nil {
		log.Fatalf("Error parsing public key: %v", err)
	}
}

func CreateToken() (string, error) {
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
		Issuer:    "your-app",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// reads the login credentials, checks them and creates JWT the token
func GenerateKey(w http.ResponseWriter, r *http.Request) {
	dr, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer dr.Close()

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

	tokenString, err := CreateToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Sorry, error while Signing Token!")
		log.Printf("Token Signing error: %v\n", err)
		return
	}
	token := context.Get(r, "user")
	strToken, _ := token.(string)

	err = db.UpdateUser(dr, strToken, tokenString)
	if err != nil {
		log.Printf("Error updating user: %v\n", err)
		http.Error(w, "Error Generating key", http.StatusInternalServerError)
		return
	}
	log.Printf("Token: here\n")
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
		dr, err := db.InitializeDB()
		if err != nil {
			http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
			return
		}
		defer dr.Close()

		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		// First, try to parse as JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		})

		if err != nil {
			// JWT parsing failed, try to validate as a simple API key from database
			user, err := FetchEachUserByToken(dr, tokenString)
			if err != nil {
				http.Error(w, "Invalid API key or token", http.StatusUnauthorized)
				log.Printf("API auth failed for token: %s\n", tokenString)
				return
			}
			// Valid API key found
			context.Set(r, "user", user.Iid)
			next(w, r)
			return
		}

		if token.Valid {
			user, err := FetchEachUserByToken(dr, tokenString)
			if err != nil {
				http.Error(w, "Unable to fetch user", http.StatusInternalServerError)
				return
			}
			context.Set(r, "user", user.Iid)
			next(w, r)
		} else {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}
	}
}
