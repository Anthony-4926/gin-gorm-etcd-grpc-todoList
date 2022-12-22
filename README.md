# 项目主要技术

- gin
- gorm
- etcd
- grpc
- jwt-go
- logrus
- viper
- protobuf

# 项目结构

## 1. gatewagy网关部分

```
api-gateway/
├── cmd                   // 启动入口
├── config                // 配置文件
├── discovery             // etcd服务注册、keep-alive、获取服务信息等等
├── internal              // 业务逻辑（不对外暴露）
│   ├── handler           // 视图层
│   └── service           // 服务层
│       └──pb             // 放置生成的pb文件
├── logs                  // 放置打印日志模块
├── middleware            // 中间件
├── pkg                   // 各种包
│   ├── e                 // 统一错误状态码
│   ├── res               // 统一response接口返回
│   └── util              // 各种工具、JWT、Logger等等..
├── routes                // http路由模块
└── wrappers              // 各个服务之间的熔断降级
```

## 2. user && task 用户与任务模块

```
user/
├── cmd                   // 启动入口
├── config                // 配置文件
├── discovery             // etcd服务注册、keep-alive、获取服务信息等等
├── internal              // 业务逻辑（不对外暴露）
│   ├── handler           // 视图层
│   ├── cache             // 缓存模块
│   ├── repository        // 持久层
│   └── service           // 服务层
│       └──pb             // 放置生成的pb文件
├── logs                  // 放置打印日志模块
└── pkg                   // 各种包
    ├── e                 // 统一错误状态码
    ├── res               // 统一response接口返回
    └── util              // 各种工具、JWT、Logger等等..
```



# 服务注册与发现

![image-20221222164658510](http://imgbed4926.oss-cn-hangzhou.aliyuncs.com/img/image-20221222164658510.png)

# [Etcd 服务注册](https://cnl25x1hkc.feishu.cn/docx/M90Fd8KJLoTE57xqTc9c3NDdn5f)
  - [服务注册准备](https://cnl25x1hkc.feishu.cn/docx/M90Fd8KJLoTE57xqTc9c3NDdn5f#HOCCd8QmOoyIAexIRRVcWOIinzf)
  - [服务注册整体流程](https://cnl25x1hkc.feishu.cn/docx/M90Fd8KJLoTE57xqTc9c3NDdn5f#YEOmdyAaCouGa0xL9YJcug3Znkp)
  - [客户端与etcd通信](https://cnl25x1hkc.feishu.cn/docx/M90Fd8KJLoTE57xqTc9c3NDdn5f#XKQkdcSw8ooOuOxE7B4cecIfnAc)

## 服务注册准备

- 服务注册需要的角色有：
  - **service**：真正执行业务逻辑，提供服务功能的。如`user service`
  - **grpc server**：代理service通信，进行远程过程调用。grpc server 会绑定服务接口。
  - **etcd**：服务注册中心
  - **register**：把grpc对外通信的地址端口，以及服务名注册到etcd身上。并且维护etcd与grpc之间的连接。

  <img src="http://imgbed4926.oss-cn-hangzhou.aliyuncs.com/img/image-20221220120011013.png" alt="image-20221220120011013" height="300dp" />

```Go
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

   // ....
   
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
```

## 服务注册整体流程

服务注册不仅仅是把服务名和grpc服务地址注册到etcd上，还需要维护etcd与grpc之间的连接。etcd的整体服务注册思路总体可以分为两个大的过程：

1. 首先，我们得向etcd申请一个位置，挂载我们的`服务名：服务地址`，这个过程叫做租赁。
   1. etcd会给我们返回一个**租赁凭证**，租赁凭证包含着`租赁ID`
   2. etcd隔一段时间会跟客户端通信一下（[具体如何通信的后边会解释](https://cnl25x1hkc.feishu.cn/docx/M90Fd8KJLoTE57xqTc9c3NDdn5f#XKQkdcSw8ooOuOxE7B4cecIfnAc)），保证客户端没挂。
   3. 如果客户端挂了，etcd会把这个服务包括租赁踢了。
2. 然后，客户端获取租赁后，就可以把自己的服务通过`租赁ID`挂载到etcd
   1. 因为etcd也可能挂，所以客户端要维持一个心跳，不断跟客户端通信
   2. 通信不成功时就需要重新注册服务

服务注册到etcd后，在etcd中是下边这个样子滴

![image-20221220121206313](http://imgbed4926.oss-cn-hangzhou.aliyuncs.com/img/image-20221220121206313.png)

## 客户端与etcd的通信

调用`keepAliveCh, _ = KeepAlive(ctx context.Context, id LeaseID) (<-chan *LeaseKeepAliveResponse, error)`申请etcd保活这个租赁，etcd就会返回一个通道。

etcd每500ms向这个通道中发送一个消息，客户端通过心跳不断消费这个通道。

- 如果客户端从通道中取出来`nil`，表示这个通道没有值了，说明etcd挂了，就需要重新注册服务。
- 当客户端挂了，这个通道就会被塞满，然后etcd再向这个通道中发消息就会被阻塞。阻塞一段时间后，etcd就认为客户端挂了，或者租期到了，etcd就会踢了这个租约。