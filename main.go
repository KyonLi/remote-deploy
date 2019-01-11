package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"remote-deploy/ssh"
	"remote-deploy/util"
	"strconv"
	"strings"
	"syscall"
)

var mossSep = "-----------------------------------------------------------------\n"

var localDir = "LiveBroadCast"
var targetDir = "/var/lib/tomcat8/webapps/"
var tmpDir = "/tmp/remote-deploy"
var hostStr = "root@tc.shencai.net.cn"
var password = ""
var privateKey = ""

func printWelcomeMessage(server ssh.Server) {
	fmt.Println(mossSep)
	serverInfo := fmt.Sprintf("Target: %s@%s:%s", server.User, server.Hostname, targetDir)
	fmt.Println(serverInfo)
	fmt.Printf("\n%s\n", mossSep)
}

func parseConfig() ssh.Client {
	hostInfo := strings.Split(strings.Replace(hostStr, ":", "@", -1), "@")
	var port = 22
	if len(hostInfo) > 2 {
		port, _ = strconv.Atoi(hostInfo[2])
	}

	server := ssh.Server{
		Hostname:   hostInfo[1],
		User:       hostInfo[0],
		Port:       port,
		Password:   password,
		PrivateKey: privateKey,
	}

	printWelcomeMessage(server)
	if !strings.Contains(server.Hostname, ":") {
		server.Hostname = server.Hostname + ":" + strconv.Itoa(server.Port)
	}
	client := ssh.Client{
		Host:       server.Hostname,
		User:       server.User,
		Password:   server.Password,
		PrivateKey: server.PrivateKey,
	}

	return client
}

func begin(client *ssh.Client) {
	if !util.IsExist(localDir) || !util.IsDir(localDir) {
		fmt.Println("本地目录不合法")
		return
	}

	fmt.Println("连接服务器...")
	if err := client.Connect(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("完成\n")

	fmt.Println("清理临时目录...")
	baseName := path.Base(localDir)
	if err := client.Execute(fmt.Sprintf("rm -rf %s; mkdir -p %s", tmpDir, tmpDir)); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("完成\n")

	fmt.Println("上传文件...")
	fileName, err := util.Compress(localDir)
	if err != nil {
		fmt.Printf("压缩失败: %s", err)
		return
	}
	if err := ssh.UploadFile(client, fileName, tmpDir); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("完成\n")

	fmt.Println("重启服务")
	if err := client.Execute(fmt.Sprintf("cd %s && tar zxf %s && systemctl stop tomcat8 && rm -rf %s && mv %s %s && systemctl start tomcat8", tmpDir, fileName, path.Join(targetDir, baseName), baseName, targetDir)); err != nil {
		fmt.Println(err)
		return
	}
	_ = os.Remove(fileName)
	fmt.Println("部署完成")
	client.Close()
}

func main() {
	client := parseConfig()
	begin(&client)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
}
