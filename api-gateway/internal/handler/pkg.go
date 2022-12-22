package handler

import (
	"api-gateway/pkg/util"
	"errors"
)

// 包装错误
func PanicIfUserError(err error) {
	if err != nil {
		err = errors.New("user Service--" + err.Error())
		util.LogrusObj.Info(err)
		panic(err)
	}
}

// 包装错误
func PanicIfTaskError(err error) {
	if err != nil {
		err = errors.New("task Service--" + err.Error())
		util.LogrusObj.Info(err)
		panic(err)
	}
}
