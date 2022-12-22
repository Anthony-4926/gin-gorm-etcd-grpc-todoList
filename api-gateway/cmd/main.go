package main

import (
	"api-gateway/config"
	"api-gateway/discovery"
	"api-gateway/internal/service"
	"api-gateway/routers"
	"context"
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
	etcdAddrs := []string{viper.GetString("etcd.address")}
	etcdResolver := discovery.NewResolver(etcdAddrs, logrus.New())
	resolver.Register(etcdResolver)
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	// 服务名
	userServiceName := viper.GetString("domain.user")
	taskServiceName := viper.GetString("domain.task")

	userConn, err := RPCConnect(ctx, userServiceName, etcdResolver)
	if err != nil {
		fmt.Println("连接服务器失败：", err)
		return
	}
	userService := service.NewUserServiceClient(userConn)

	taskConn, _ := RPCConnect(ctx, taskServiceName, etcdResolver)
	taskService := service.NewTaskServiceClient(taskConn)

	ginRouter := routers.NewRouter(userService, taskService)

	server := &http.Server{
		Addr:           viper.GetString("server.port"),
		Handler:        ginRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("绑定失败，可能端口被占用", err)
	}

}

func RPCConnect(ctx context.Context, serviceName string, etcdRegister *discovery.Resolver) (conn *grpc.ClientConn, err error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	addr := fmt.Sprintf("%s:///%s", etcdRegister.Scheme(), serviceName)
	conn, err = grpc.Dial(addr, opts...)
	return
}
