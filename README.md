# remote-deploy

远程部署Java web项目

## 使用方法

1. 修改 main.go 中的服务器配置

    `localDir` 要上传的项目路径
    
    `targetDir` 服务器目标路径

    `hostStr` 服务器域名或IP

    `password` 密码

    `privateKey` 登录私钥

2. 编译运行 `go run main.go`