package models

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
