package models

type User struct {
	BaseModel

	Username    string `gorm:"column:login"`
	Email       string `gorm:"column:email"`
	Password    string `gorm:"column:password" json:"-"`
	IsSuperuser bool   `gorm:"column:is_superuser"`
	IsStaff     bool   `gorm:"column:is_staff"`
	IsActive    bool   `gorm:"column:is_active"`
}

func (User) TableName() string {
	return "users_user"
}
