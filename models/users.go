package models

type User struct {
	BaseModel

	Name        string `gorm:"name"`
	Username    string `gorm:"login"`
	Email       string `gorm:"email"`
	Password    string `gorm:"password" json:"-"`
	IsSuperuser bool   `gorm:"is_superuser"`
	IsStaff     bool   `gorm:"is_staff"`
	IsActive    bool   `gorm:"is_active"`
}

func (User) TableName() string {
	return "users_user"
}
