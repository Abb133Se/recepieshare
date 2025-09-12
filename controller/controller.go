package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/messages"
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
	Token  string `json:"token"`
	UserID uint   `json:"user_id"`
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

type TimeSeriesData struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type AnalyticsRequest struct {
	Metric string `form:"metric" binding:"required"` // views|favorites|ratings|site
	Period string `form:"period" binding:"required"` // week|month|year
}

type ChartJSResponse struct {
	Labels   []string         `json:"labels"`
	Datasets []ChartJSDataset `json:"datasets"`
}

type ChartJSDataset struct {
	Label           string  `json:"label"`
	Data            []int64 `json:"data"`
	BorderColor     string  `json:"borderColor"`
	BackgroundColor string  `json:"backgroundColor"`
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: messages.User.EmptyInfoErr.String()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var existingUser model.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{Error: messages.User.UserAlreadyExists.String()})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.User.EmailCheckErr.String()})
		return
	}

	salt, err := utils.GenerateSalt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.User.GenerateSaltFail.String()})
		return
	}

	hashedPassword := utils.HashPassword(req.Password, salt)

	user := model.User{
		Name:     req.Name,
		LastName: req.LastName,
		Email:    req.Email,
		Password: hashedPassword,
		Salt:     salt,
		Role:     "user",
	}

	err = db.Create(&user).Error
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			c.JSON(http.StatusConflict, ErrorResponse{Error: messages.User.EmailExistsErr.String()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.User.UserCreateFailed.String()})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: messages.User.UserCreatedSuccess.String()})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var user model.User
	if err = db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: messages.User.LoginInvalidEmailPass.String()})
			return
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.User.UserFetchFail.String()})
			return
		}
	}

	enteredPassword := utils.HashPassword(req.Password, user.Salt)

	if user.Password != enteredPassword {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: messages.User.LoginInvalidEmailPass.String()})
		return
	}

	tokenStr, err := token.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.User.GeneratTokenFail.String()})
		return
	}

	c.JSON(http.StatusOK, UserLoginResponse{Token: tokenStr, UserID: user.ID})
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: messages.User.EmailCheckErr.String()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var user model.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, ForgotPasswordResponse{Message: messages.User.PasswordResetFailed.String()})
		return
	}

	resetToken, err := token.GenerateResetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.User.GeneratTokenFail.String()})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.User.PasswordResetCreateFailed.String()})
		return
	}

	c.JSON(http.StatusOK, ForgotPasswordResponse{
		Message:    messages.Common.Success.String(),
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var user model.User
	if err := db.Where("password_reset_token = ?", req.Token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: messages.User.TokenInvalid.String()})
		return
	}

	if user.PasswordResetExpiresAt == nil || user.PasswordResetExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: messages.User.TokenExpired.String()})
		return
	}

	salt, err := utils.GenerateSalt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.User.GenerateSaltFail.String()})
		return
	}
	hashed := utils.HashPassword(req.NewPassword, salt)

	user.Password = hashed
	user.Salt = salt
	user.PasswordResetToken = ""
	user.PasswordResetExpiresAt = nil

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.User.PasswordResetFailed.String()})
		return
	}

	c.JSON(http.StatusOK, ResetPasswordResponse{Message: messages.User.PasswordResetSuccess.String()})
}

// GetAnalytics godoc
// @Summary      Get analytics time-series data
// @Description  Returns aggregated counts of metrics (views, favorites, ratings, site visits, recipes created) grouped by day or month depending on the requested period.
// @Tags         analytics
// @Produce      json
// @Param        metric query string true "Metric to analyze" Enums(views, favorites, ratings, site, recipes)
// @Param        period query string true "Time period" Enums(week, month, year)
// @Success      200 {array} TimeSeriesData
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /admin/analytics [get]
func GetAnalytics(c *gin.Context) {
	var req AnalyticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	table, column, err := utils.GetTableAndColumn(req.Metric)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric"})
		return
	}

	var query string
	switch req.Period {
	case "week":
		query = fmt.Sprintf(`
			SELECT DATE_FORMAT(%s, '%%Y-%%m-%%d') as label, COUNT(*) as count
			FROM %s
			WHERE %s >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
			GROUP BY DATE_FORMAT(%s, '%%Y-%%m-%%d')
			ORDER BY DATE_FORMAT(%s, '%%Y-%%m-%%d');
		`, column, table, column, column, column)

	case "month":
		query = fmt.Sprintf(`
			SELECT DATE_FORMAT(%s, '%%Y-%%m-%%d') as label, COUNT(*) as count
			FROM %s
			WHERE %s >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
			GROUP BY DATE_FORMAT(%s, '%%Y-%%m-%%d')
			ORDER BY DATE_FORMAT(%s, '%%Y-%%m-%%d');
		`, column, table, column, column, column)

	case "year":
		query = fmt.Sprintf(`
            SELECT DATE_FORMAT(%s, '%%Y-%%m') as label, COUNT(*) as count
            FROM %s
            WHERE %s >= DATE_SUB(CURDATE(), INTERVAL 12 MONTH)
            GROUP BY DATE_FORMAT(%s, '%%Y-%%m')
            ORDER BY DATE_FORMAT(%s, '%%Y-%%m');
        `, column, table, column, column, column)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period"})
		return
	}

	// Fetch DB results
	var results []TimeSeriesData
	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}
	if err := db.Raw(query).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	// Transform into Chart.js format
	labels := make([]string, 0, len(results))
	data := make([]int64, 0, len(results))
	for _, r := range results {
		labels = append(labels, r.Label)
		data = append(data, r.Count)
	}

	chart := ChartJSResponse{
		Labels: labels,
		Datasets: []ChartJSDataset{
			{
				Label:           strings.Title(req.Metric),
				Data:            data,
				BorderColor:     "rgba(75,192,192,1)",
				BackgroundColor: "rgba(75,192,192,0.2)",
			},
		},
	}

	c.JSON(http.StatusOK, chart)
}
