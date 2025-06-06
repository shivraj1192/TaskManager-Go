package userController

import (
	"net/http"
	"os"
	"task-manager/config"
	"task-manager/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *gin.Context) {
	var logUser model.LoginUser
	var user model.User
	db := config.DB

	if err := c.ShouldBindJSON(&logUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to bind with user while login" + err.Error()})
		return
	}

	if logUser.Email != "" {
		if err := db.Where("email=?", logUser.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Credentials (email) for login" + err.Error()})
			return
		}
	} else {
		if err := db.Where("uname=?", logUser.Uname).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Credentials (username) for login" + err.Error()})
			return
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(logUser.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Credentials (password) for login" + err.Error()})
		return
	}
	tokenString, ok := GenerateToken(user, c)
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "Unable to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})

}

func GenerateToken(user model.User, c *gin.Context) (string, bool) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
		"user_id":    user.ID,
		"authorized": true,
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Unable to create token"})
		return "", false
	}
	c.Header("Authorization", "Bearer "+tokenString)

	return tokenString, true
}
