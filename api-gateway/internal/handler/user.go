package handler

import (
	"api-gateway/internal/service"
	"api-gateway/pkg/e"
	"api-gateway/pkg/res"
	"api-gateway/pkg/util"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

// UserRegister 用户登录
func UserRegister(ginCtx *gin.Context) {
	var userReq service.UserRequest
	PanicIfUserError(ginCtx.Bind(&userReq))
	// gin.Keys中获取服务实例
	userService := ginCtx.Keys["user"].(service.UserServiceClient)
	// 执行远程过程调用
	userResp, err := userService.UserRegister(context.Background(), &userReq)
	PanicIfUserError(err)
	r := res.Response{
		Data:   userResp,
		Status: uint(userResp.Code),
		Msg:    e.GetMsg(uint(userResp.GetCode())),
		Error:  "",
	}
	ginCtx.JSON(http.StatusOK, r)
}

// UserLogin 用户登录
func UserLogin(ginCtx *gin.Context) {
	var userReq service.UserRequest
	PanicIfUserError(ginCtx.Bind(&userReq))
	// gin.Keys中获取服务实例
	userService := ginCtx.Keys["user"].(service.UserServiceClient)
	// 执行远程过程调用
	userResp, err := userService.UserLogin(context.Background(), &userReq)
	PanicIfUserError(err)
	token, err := util.GenerateToken(uint(userResp.UserDetail.UserID))

	r := res.Response{
		Data: res.TokenData{
			User:  userResp.UserDetail,
			Token: token,
		},
		Status: uint(userResp.Code),
		Msg:    e.GetMsg(uint(userResp.GetCode())),
	}
	ginCtx.JSON(http.StatusOK, r)
}
