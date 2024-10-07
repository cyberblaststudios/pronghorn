package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"pronghorn-app/db"
	"time"
)

type UserSession struct {
	HashedSessionToken []byte
	UserID             int
	TimeCreated        time.Time
	TimeExpired        time.Time
}

type UserIDKey struct{}

func generateRandomBytes(length int) ([]byte, error) {
	randomBytes := make([]byte, length)

	_, randomGenerationError := rand.Read(randomBytes)

	if randomGenerationError != nil {
		return []byte{}, randomGenerationError
	}

	return randomBytes, nil
}

func CreateNewUserSession(userId int, expirationTimeDelta time.Duration) ([]byte, UserSession, error) {
	generatedKey, generationError := generateRandomBytes(100)

	if generationError != nil {
		return []byte{}, UserSession{}, generationError
	}

	// for added protection, we want to hash the key
	sha256Hash := sha256.New()
	sha256Hash.Write(generatedKey)
	hashedKey := sha256Hash.Sum(nil)

	// we want to add expiration and generation times to put in the database
	timeNow := time.Now()
	timeExpired := timeNow.Add(expirationTimeDelta)

	return generatedKey, UserSession{HashedSessionToken: hashedKey,
		UserID:      userId,
		TimeCreated: timeNow,
		TimeExpired: timeExpired}, nil
}

// authorization handler to wrap protected api routes and resources
func AuthHandler(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// handle username and password authentication

		sessionCookie, sessionCookieError := request.Cookie("Login")

		if sessionCookieError != nil {
			http.Error(writer, "Unauthorized", http.StatusUnauthorized)
			log.Println("Session cookie for Login is nil", sessionCookieError)
			return
		}

		sessionCookieValue := sessionCookie.Value

		// check for a blank cookie value and throw unauthorized
		if sessionCookieValue == "" {
			http.Error(writer, "Unauthorized", http.StatusUnauthorized)
			log.Println("Session cookie for Login is blank")
			return
		}

		decodedSessionCookie, sessionCookieDecodeError := hex.DecodeString(sessionCookieValue)

		if sessionCookieDecodeError != nil {
			http.Error(writer, "Unauthorized", http.StatusUnauthorized)
			log.Println("Hex decode failed", sessionCookieDecodeError)
			return
		}

		// hash the session cookie coming in so that we can match it to an existing session
		sha256Hash := sha256.New()
		sha256Hash.Write(decodedSessionCookie)
		hashedKey := sha256Hash.Sum(nil)

		dbConnection := db.GetDBConnection()

		var sessionId []byte
		var userID int
		var createdAt time.Time
		var expiresAt time.Time

		findSessionQueryError := dbConnection.QueryRow("SELECT id,user_id,created_at,expires_at FROM session where id = $1", hashedKey).Scan(&sessionId,
			&userID, &createdAt, &expiresAt)

		if findSessionQueryError != nil {
			http.Error(writer, "Unauthorized", http.StatusUnauthorized)
			log.Println("Failed to find session in DB", findSessionQueryError)
			return
		}

		// write the authenticated user id
		ctx := request.Context()
		ctx = context.WithValue(ctx, UserIDKey{}, userID)
		nextHandler.ServeHTTP(writer, request.WithContext(ctx))
	})
}
