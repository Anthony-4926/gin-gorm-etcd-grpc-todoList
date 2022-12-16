package repository

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"user/internal/service"
)

type User struct {
	UserID         uint   `gorm:"primarykey"`
	UserName       string `gorm:"unique"`
	NickName       string
	PasswordDigest string
}

const (
	PASSWORD_COST = 12
)

func (user *User) BindUser(req *service.UserRequest) (err error) {
	if exist := user.IsUserExist(req); exist {
		return nil
	}
	return errors.New("UserName Not Exist")
}

// IsUserExist 判断用户是否存在
func (user *User) IsUserExist(req *service.UserRequest) bool {
	return !(DB.Where("name = ?", req.UserName).First(user).Error == gorm.ErrRecordNotFound)
}

func (*User) Create(req *service.UserRequest) error {
	var user User
	var count int64
	DB.Where("user_name=?", req.UserName).Count(&count)
	if count != 0 {
		return errors.New("UserName Exist")
	}
	user = User{
		UserName: req.UserName,
		NickName: req.NickName,
	}
	_ = user.SetPassword(req.Password)
	if err := DB.Create(&user).Error; err != nil {
		//util.LogrusObj.Error("Insert User Error:" + err.Error())
		return err
	}
	return nil
}

// 加密密码
func (user *User) SetPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), PASSWORD_COST)
	if err != nil {
		return err
	}
	user.PasswordDigest = string(bytes)
	return nil
}

// 检验密码
func (user *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(password))
	return err == nil
}
