package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/token"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type UserSignupRequest struct {
	Name     string `json:"name" binding:"required"`
	LastName string `json:"last_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResponse struct {
	Token string `json:"token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ForgotPasswordResponse struct {
	Message    string `json:"message"`
	ResetToken string `json:"reset_token,omitempty"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// Signup godoc
// @Summary      Register a new user
// @Description  Creates a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      UserSignupRequest  true  "User signup info"
// @Success      200   {object}  SimpleMessageResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /signup [post]
func Signup(c *gin.Context) {
	var req UserSignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if req.Name == "" || req.LastName == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user information shouldn't be empty"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to server"})
		return
	}

	var existingUser model.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "email already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to check email"})
		return
	}

	salt, err := utils.GenerateSalt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate salt"})
		return
	}

	hashedPassword := utils.HashPassword(req.Password, salt)

	user := model.User{
		Name:     req.Name,
		LastName: req.LastName,
		Email:    req.Email,
		Password: hashedPassword,
		Salt:     salt,
	}

	err = db.Create(&user).Error
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create user"})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: "user registered successfully"})
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      UserLoginRequest  true  "Login credentials"
// @Success      200          {object}  UserLoginResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /login [post]
func Login(c *gin.Context) {
	var req UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to db"})
		return
	}

	var user model.User
	if err = db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid email or password"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve user"})
			return
		}
	}

	enteredPassword := utils.HashPassword(req.Password, user.Salt)

	if user.Password != enteredPassword {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid email or password"})
		return
	}

	tokenStr, err := token.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, UserLoginResponse{Token: tokenStr})
}

// ForgotPasswordHandler godoc
// @Summary      Initiate password reset
// @Description  Sends password reset instructions if email exists
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        email  body      ForgotPasswordRequest  true  "User email"
// @Success      200    {object}  ForgotPasswordResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /forgot-password [post]
func ForgotPasswordHandler(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid email"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to server"})
		return
	}

	var user model.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "if email exists, reset instructions have been sent"})
		return
	}

	resetToken, err := token.GenerateResetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate reset token"})
		return
	}

	user.PasswordResetToken = resetToken
	expiresAt := time.Now().Add(15 * time.Minute)
	user.PasswordResetExpiresAt = &expiresAt

	err = db.Model(&user).Updates(map[string]any{
		"password_reset_token":      resetToken,
		"password_reset_expires_at": expiresAt,
		"updated_at":                time.Now()}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to store reset token"})
		return
	}

	c.JSON(http.StatusOK, ForgotPasswordResponse{
		Message:    "reset link sent",
		ResetToken: resetToken,
	})
}

// ResetPasswordHandler godoc
// @Summary      Reset user password
// @Description  Resets password using reset token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        data  body      ResetPasswordRequest  true  "Reset password data"
// @Success      200   {object}  ResetPasswordResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /reset-password [post]
func ResetPasswordHandler(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid input"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to server"})
		return
	}

	var user model.User
	if err := db.Where("password_reset_token = ?", req.Token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid or expired token"})
		return
	}

	if user.PasswordResetExpiresAt == nil || user.PasswordResetExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "token has expired"})
		return
	}

	salt, err := utils.GenerateSalt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate salt"})
		return
	}
	hashed := utils.HashPassword(req.NewPassword, salt)

	user.Password = hashed
	user.Salt = salt
	user.PasswordResetToken = ""
	user.PasswordResetExpiresAt = nil

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, ResetPasswordResponse{Message: "password reset successfully"})
}
