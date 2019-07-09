package apis

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"moowda/services"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"

	"moowda/app"
	apiErrors "moowda/errors"
	"moowda/models"
)

type UserAPI struct {
	db                  *gorm.DB
	notificationService *services.NotificationService
}

func NewUserAPI(db *gorm.DB, notificationService *services.NotificationService) *UserAPI {
	return &UserAPI{db: db, notificationService: notificationService}
}

func (s *UserAPI) Me(c echo.Context) error {
	user := c.Get("user").(*models.User)
	return c.JSON(http.StatusOK, user)
}

func (s *UserAPI) Register(c echo.Context) error {
	registerRequest := new(models.RegisterRequest)
	if err := c.Bind(registerRequest); err != nil {
		return err
	}

	if err := registerRequest.Validate(); err != nil {
		return apiErrors.InvalidData(err.(validation.Errors))
	}

	salt := RandASCIIBytes(12)
	hashStr := EncodePassword(registerRequest.Password, salt)

	user := &models.User{
		Username: registerRequest.Username,
		Email:    registerRequest.Email,
		Password: fmt.Sprintf("%s$%s$%s$%s", "pbkdf2_sha256", "150000", string(salt), hashStr),
	}

	tx := s.db.Begin()

	if err := s.db.Where("login = ?", user.Username).Find(user).Error; err == nil {
		tx.Rollback()
		return apiErrors.BadRequest(errors.Errorf("username %s already taken", user.Username))
	}

	if err := s.db.Where("email = ?", user.Email).Find(user).Error; err == nil {
		tx.Rollback()
		return apiErrors.BadRequest(errors.Errorf("email %s already taken", user.Email))
	}

	if err := s.db.Create(user).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "create new user")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "commit transaction")
	}

	signedToken, err := GenerateJWT(user)
	if err != nil {
		return apiErrors.InternalServerError(err)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"type":  "Bearer",
		"token": signedToken,
	})
}

func GenerateJWT(user *models.User) (string, error) {
	// Set claims
	claims := jwt.MapClaims{
		"userID": user.ID,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	return token.SignedString([]byte(app.Config.JWTSigningKey))
}

func (s *UserAPI) Login(c echo.Context) error {
	loginRequest := new(models.LoginRequest)
	if err := c.Bind(loginRequest); err != nil {
		return err
	}

	if err := loginRequest.Validate(); err != nil {
		return apiErrors.InvalidData(err.(validation.Errors))
	}

	var user models.User
	if err := s.db.Where("login = ?", loginRequest.Username).Find(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return apiErrors.Unauthorized("wrong username or password")
		}
		return err
	}

	hashParts := strings.Split(user.Password, "$")
	if len(hashParts) != 4 {
		return apiErrors.Unauthorized("wrong username or password")
	}

	encodedPassword := EncodePassword(loginRequest.Password, []byte(hashParts[2]))

	if encodedPassword != hashParts[3] {
		return apiErrors.Unauthorized("wrong username or password")
	}

	// Set claims
	claims := jwt.MapClaims{
		"userID": user.ID,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(app.Config.JWTSigningKey))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"type":  "Bearer",
		"token": t,
	})
}

func (s *UserAPI) RestoreRequest(c echo.Context) error {
	restoreRequest := new(models.PasswordRestoreRequest)
	if err := c.Bind(restoreRequest); err != nil {
		return err
	}

	if err := restoreRequest.Validate(); err != nil {
		return apiErrors.InvalidData(err.(validation.Errors))
	}

	var user models.User
	if err := s.db.Where("email = ?", restoreRequest.Email).Find(&user).Error; err != nil {
		return apiErrors.BadRequest(errors.Errorf("user with email %s doesn't exists", restoreRequest.Email))
	}

	hash := GenerateHash()

	if err := s.db.Model(&user).UpdateColumns(models.User{ResetPasswordHash: &hash}).Error; err != nil {
		return apiErrors.InternalServerError(errors.Errorf("update hash for %s", restoreRequest.Email))
	}

	fmt.Printf(">> %v", user)

	if err := s.notificationService.SendEmail(
		app.Config.DefaultEmailAddress,
		user.Email,
		"Reset Password Request",
		fmt.Sprintf("reset password request %s/restore/%s", app.Config.BaseURL, hash),
		fmt.Sprintf("reset password request %s/restore/%s", app.Config.BaseURL, hash),
	); err != nil {
		return apiErrors.InternalServerError(err)
	}

	return c.NoContent(http.StatusOK)
}

func GenerateHash() string {
	hash := md5.New()
	hash.Write(RandASCIIBytes(20))
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *UserAPI) Restore(c echo.Context) error {
	restoreRequest := new(models.RestoreRequest)
	if err := c.Bind(restoreRequest); err != nil {
		return err
	}

	if err := restoreRequest.Validate(); err != nil {
		return apiErrors.InvalidData(err.(validation.Errors))
	}

	var user models.User
	if err := s.db.Where("reset_password_hash = ?", restoreRequest.Hash).Find(&user).Error; err != nil {
		return apiErrors.BadRequest(errors.Errorf("user with hash %s doesn't exists", restoreRequest.Hash))
	}

	salt := RandASCIIBytes(12)
	hashStr := EncodePassword(restoreRequest.Password, salt)

	newPassword := fmt.Sprintf("%s$%s$%s$%s", "pbkdf2_sha256", "150000", string(salt), hashStr)
	if err := s.db.Model(&user).Updates(map[string]interface{}{"password": newPassword, "reset_password_hash": gorm.Expr("NULL")}).Error; err != nil {
		return apiErrors.InternalServerError(errors.Errorf("update password"))
	}

	signedToken, err := GenerateJWT(&user)
	if err != nil {
		return apiErrors.InternalServerError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": signedToken,
	})
}

func EncodePassword(password string, salt []byte) string {
	hash := pbkdf2.Key([]byte(password), salt, 150000, sha256.Size, sha256.New)
	hashStr := base64.StdEncoding.EncodeToString(hash)

	return hashStr
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandASCIIBytes(n int) []byte {
	output := make([]byte, n)
	// We will take n bytes, one byte for each character of output.
	randomness := make([]byte, n)
	// read all random
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}
	l := len(letterBytes)
	// fill output
	for pos := range output {
		// get random item
		random := uint8(randomness[pos])
		// random % 64
		randomPos := random % uint8(l)
		// put into output
		output[pos] = letterBytes[randomPos]
	}
	return output
}
