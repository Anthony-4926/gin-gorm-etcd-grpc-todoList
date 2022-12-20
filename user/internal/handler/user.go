package handler

import (
	"context"
	"user/internal/repository"
	"user/internal/service"
	"user/pkg/e"
)

// UserService 真正包含服务动作的对象
type UserService struct {
}

func NewUserService() *UserService {
	return new(UserService)
}

// UserLogin 用户登录
func (*UserService) UserLogin(ctx context.Context, req *service.UserRequest) (resp *service.UserDetailResponse, err error) {
	var user repository.User
	resp = new(service.UserDetailResponse)
	resp.Code = e.SUCCESS
	err = user.BindUser(req)
	if err != nil {
		resp.Code = e.ERROR
		return resp, err
	}
	resp.UserDetail = SerializeUser(user)
	return resp, nil
}

func (*UserService) UserRegister(ctx context.Context, req *service.UserRequest) (resp *service.UserDetailResponse, err error) {
	var user repository.User
	resp = new(service.UserDetailResponse)
	resp.Code = e.SUCCESS
	err = user.Create(req)
	if err != nil {
		resp.Code = e.ERROR
		return resp, err
	}
	resp.UserDetail = SerializeUser(user)
	return resp, nil
}

// SerializeUser 序列化
func SerializeUser(item repository.User) *service.UserModel {
	userModel := service.UserModel{
		UserID:   uint32(item.UserID),
		NickName: item.NickName,
		UserName: item.UserName,
	}
	return &userModel
}
