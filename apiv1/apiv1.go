package apiv1

import (
	"encoding/hex"
	"log"
	"net/http"
	"pronghorn-app/auth"
	"pronghorn-app/db"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const apiPrefix = "/apiv1"

func testRoute(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func loginRoute(writer http.ResponseWriter, request *http.Request) {
	dbConnection := db.GetDBConnection()

	username := strings.TrimSpace(request.Header.Get("username"))
	password := request.Header.Get("password")

	// checking for blank usernames or only whitespace
	if username == "" {
		http.Error(writer, "Incorrect username or password", http.StatusUnauthorized)
		return
	}

	// get the existing user out of the database to check the attmepted password against the hash
	var passwordHash []byte
	var userId int
	hashQueryError := dbConnection.QueryRow("SELECT password_hash,id FROM users where username = $1", username).Scan(&passwordHash, &userId)

	if hashQueryError != nil {
		// user probably doesn't exist, as the hash was not found by the username
		// we don't want to say user doens't exist, because that is a finding
		http.Error(writer, "Incorrect username or password", http.StatusUnauthorized)
		log.Println("User probably doesn't exist: ", hashQueryError)
		return
	}

	// make sure that the password matches
	hashComparisonError := bcrypt.CompareHashAndPassword(passwordHash, []byte(password))

	if hashComparisonError != nil {
		http.Error(writer, "Incorrect username or password", http.StatusUnauthorized)
		log.Println("Error comparing hashes: ", hashComparisonError)
		return
	}

	// create the user session now that authentication has succeeded
	rawKey, userSession, sessionCreateError := auth.CreateNewUserSession(userId, 24*60*60)
	// encode the key in hex to send down to the client as a cookie
	encodedKey := hex.EncodeToString(rawKey)

	if sessionCreateError != nil {
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error logging in user: ", sessionCreateError)
		return
	}

	// store the user session in the database
	_, sessionInsertError := dbConnection.Exec("INSERT INTO sessions (id, user_id, created_at, expires_at) VALUES ($1, $2, $3, $4)", userSession.HashedSessionToken,
		userSession.UserID,
		userSession.TimeCreated,
		userSession.TimeExpired)

	if sessionInsertError != nil {
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error logging in user: ", sessionInsertError)
		return
	}

	// create and send the raw session key in the session cookie
	loginSessionCookie := http.Cookie{
		Name:    "Login",
		Value:   encodedKey,
		Expires: userSession.TimeExpired,
	}

	http.SetCookie(writer, &loginSessionCookie)
}

func StartRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("GET "+apiPrefix+"/login", loginRoute)
	serveMux.Handle("GET "+apiPrefix+"/test", auth.AuthHandler(http.HandlerFunc(testRoute)))
}
