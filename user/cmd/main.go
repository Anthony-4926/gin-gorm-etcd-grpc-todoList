package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
	"user/config"
	"user/discovery"
	"user/internal/handler"
	"user/internal/repository"
	"user/internal/service"
)

func main() {
	config.InitConfig()
	repository.InitDB()
	// gprc 与外界通信的地址端口  127.0.0.1: 10001
	grpcAddress := viper.GetString("grpcServer.grpcAddress")

	// 准备一个grpc server
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	// 服务绑定：服务动作 --> grpc
	// 把我们的服务动作绑定在grpc server中，这样，grpc server就是一个代理
	service.RegisterUserServiceServer(grpcServer, handler.NewUserService())

	// 服务注册：grpc --> etcd
	// 准备一个etcd
	etcdAddress := []string{viper.GetString("etcd.address")}
	register := discovery.NewRegister(etcdAddress, logrus.New())

	userServiceInfor := discovery.ServiceInfo{
		Name: viper.GetString("grpcServer.domain"),
		Addr: grpcAddress,
	}
	if _, err := register.Register(userServiceInfor, 10); err != nil {
		panic(err)
	}

	// 对外暴露grpc，监听对grpc的请求
	listen, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		panic(err)
	}
	if err := grpcServer.Serve(listen); err != nil {
		panic(err)
	}
}
