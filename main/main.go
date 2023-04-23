package main

import (
	"fmt"
	"log"
	"net/http"
	"pronghorn-app/apiv1"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func generateAuthToken(username string, signingKey []byte) (string, error) {

	expires := time.Now().Add(24 * time.Hour).Unix()

	claims := jwt.StandardClaims{}
	claims.Subject = username
	claims.ExpiresAt = expires

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// make sure to sign with the key
	signedKey, err := token.SignedString(signingKey)

	if err != nil {
		return "", err
	}

	return signedKey, nil
}

func main() {

	fmt.Println("Starting http server....")

	signingKey := []byte("testingkey")

	router := gin.New()

	router.POST("/login", func(context *gin.Context) {
		// handle the auth here

		username := context.Request.Header.Get("username")
		password := context.Request.Header.Get("password")

		if !(username == "admin" && password == "password1") {
			context.JSON(200, gin.H{"Error": "Incorrect username or password"})
			return
		}

		// generate the token
		token, err := generateAuthToken(username, signingKey)

		if err != nil {
			log.Println("Error logging in user: ", err)
			return
		}

		// finally pass the token to the user
		context.JSON(200, gin.H{"token": token})
	})

	// setup all of the routes
	apiv1.Start(router)

	http.ListenAndServe(":8080", router)
}
