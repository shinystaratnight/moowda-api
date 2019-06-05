package apis

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/pbkdf2"
	"moowda/app"
	"moowda/models"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"

	"net/http"
)

type UserAPI struct {
	db *gorm.DB
}

func NewUserAPI(db *gorm.DB) *UserAPI {
	return &UserAPI{db: db}
}

type RegisterForm struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *UserAPI) Me(c echo.Context) error {
	user := c.Get("user").(*models.User)
	return c.JSON(http.StatusOK, user)
}

func (s *UserAPI) Register(c echo.Context) error {
	registerForm := new(RegisterForm)
	if err := c.Bind(registerForm); err != nil {
		return err
	}

	salt := RandASCIIBytes(12)
	hashStr := EncodePassword(registerForm.Password, salt)

	user := &models.User{
		Username: registerForm.Username,
		Email:    registerForm.Email,
		Password: fmt.Sprintf("%s$%s$%s$%s", "pbkdf2_sha256", "150000", string(salt), hashStr),
	}

	if err := s.db.Create(user).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}

type LoginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *UserAPI) Login(c echo.Context) error {
	form := new(LoginForm)
	if err := c.Bind(form); err != nil {
		return err
	}

	var user models.User
	if err := s.db.Where("login = ?", form.Username).Find(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.NoContent(http.StatusUnauthorized)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	hashParts := strings.Split(user.Password, "$")
	if len(hashParts) != 4 {
		return c.NoContent(http.StatusInternalServerError)
	}

	encodedPassword := EncodePassword(form.Password, []byte(hashParts[2]))

	if encodedPassword != hashParts[3] {
		return c.NoContent(http.StatusUnauthorized)
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
