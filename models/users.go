package models

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type User struct {
	BaseModel

	Username    string `gorm:"column:login" json:"username"`
	Email       string `gorm:"column:email" json:"email"`
	Password    string `gorm:"column:password" json:"-"`
	IsSuperuser bool   `gorm:"column:is_superuser" json:"-"`
	IsStaff     bool   `gorm:"column:is_staff" json:"-"`
	IsActive    bool   `gorm:"column:is_active" json:"-"`
}

func (User) TableName() string {
	return "users_user"
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
