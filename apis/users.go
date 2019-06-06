package apis

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"

	"moowda/app"
	apiErrors "moowda/errors"
	"moowda/models"
)

type UserAPI struct {
	db *gorm.DB
}

func NewUserAPI(db *gorm.DB) *UserAPI {
	return &UserAPI{db: db}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (f RegisterRequest) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Username, validation.Required, validation.Length(1, 24)),
		validation.Field(&f.Email, validation.Required, is.Email),
		validation.Field(&f.Password, validation.Required),
	)
}

func (s *UserAPI) Me(c echo.Context) error {
	user := c.Get("user").(*models.User)
	return c.JSON(http.StatusOK, user)
}

func (s *UserAPI) Register(c echo.Context) error {
	registerRequest := new(RegisterRequest)
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

	if err := s.db.Where("username = ?", user.Username).Find(user).Error; err == nil {
		return apiErrors.BadRequest(errors.Errorf("username %s already taken", user.Username))
	}

	if err := s.db.Where("email = ?", user.Email).Find(user).Error; err == nil {
		return apiErrors.BadRequest(errors.Errorf("email %s already taken", user.Email))
	}

	if err := s.db.Create(user).Error; err != nil {
		return errors.Wrap(err, "create new user")
	}

	if err := tx.Commit().Error; err != nil {
		return errors.Wrap(err, "commit transaction")
	}

	// Set claims
	claims := jwt.MapClaims{
		"userID": user.ID,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	signedToken, err := token.SignedString([]byte(app.Config.JWTSigningKey))
	if err != nil {
		return errors.Wrap(err, "sign JWT token")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"type":  "Bearer",
		"token": signedToken,
	})
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (f LoginRequest) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Username, validation.Required, validation.Length(1, 24)),
		validation.Field(&f.Password, validation.Required),
	)
}

func (s *UserAPI) Login(c echo.Context) error {
	loginRequest := new(LoginRequest)
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

	return c.JSON(http.StatusOK, map[string]string{
		"type":  "Bearer",
		"token": t,
	})
}

func (s *UserAPI) RestoreRequest(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (s *UserAPI) Restore(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"token": "-",
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
