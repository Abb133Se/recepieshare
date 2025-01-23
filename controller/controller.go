package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/token"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func Signup(c *gin.Context) {
	var req model.User
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(req)
	if req.Name == "" || req.LastName == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user information shouldn't be empty"})
		return
	}

	salt, err := utils.GenerateSalt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate salt"})
		return
	}
	req.Salt = salt

	hashedPassword := utils.HashPassword(req.Password, salt)
	req.Password = hashedPassword

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to server"})
		return
	}

	err = db.Create(&req).Error
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user registered successfully"})

}

func Login(c *gin.Context) {
	var req struct {
		Email    string
		Password string
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to db"})
		return
	}

	var user model.User
	if err = db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password 1"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user"})
			return
		}
	}

	enteredPassword := utils.HashPassword(req.Password, user.Salt)

	if user.Password != enteredPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password 2"})
		return
	}

	token, err := token.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
