package main

import (
	"api-gateway/config"
	"api-gateway/discovery"
	"api-gateway/internal/service"
	"api-gateway/routers"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config.InitConfig()
	etcdAddrs := []string{viper.GetString("etcd.address")}
	etcdResolver := discovery.NewResolver(etcdAddrs, logrus.New())
	resolver.Register(etcdResolver)

	go startListen()
	{
		osSignal := make(chan os.Signal, 1)
		signal.Notify(osSignal, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
		s := <-osSignal
		fmt.Println("exit!", s)
	}
	fmt.Println("gateway listen on: 4000")
}

func startListen() {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	grpcConn, _ := grpc.Dial("127.0.0.1:10001", opts...)
	userService := service.NewUserServiceClient(grpcConn)

	ginRouter := routers.NewRouter(userService)

	server := &http.Server{
		Addr:           viper.GetString("server.port"),
		Handler:        ginRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("绑定失败，可能端口被占用", err)
	}

}
