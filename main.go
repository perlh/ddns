package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type Client struct {
	NetworkCard  string `json:"network_card"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Root         string `json:"root"`
	RootPassword string `json:"root_password"`
	ServerAddr   string `json:"server_addr"`
	LocalDomain  string `json:"local_domain"`
	ScreenTime   int    `json:"screen_time"` // Second
	ClientDomain
}
type ClientDomain struct {
	DnsType string
	ip      net.IP
	cname   string
}

type Server struct {
	dnsFile    string
	listenIP   string
	listenPort int
	dnsPort    int
	screenTime int
	debug      bool
	Domain     string
	password   string
	user       string
}

// Server属性个数
func (server *Server) len() int {
	return 5
}

//type Client struct {
//}

var (
	server Server
	client Client
)
var logger *log.Logger

func init() {
	//指定路径的文件，无则创建
	//fmt.Println("xxx")
	logFile, err := os.OpenFile("./logs/log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	logger = log.New(logFile, "[test]", log.Ltime)
}

func main() {
	logger.Println("start server")
	var serverPath string
	var clientPath string
	flag.StringVar(&serverPath, "s", "", "")
	flag.StringVar(&clientPath, "c", "", "")
	flag.Parse()
	log.Println(clientPath)
	// 检查参数合法性
	if serverPath == "" && clientPath == "" {
		return
	}
	if clientPath != "" {
		if serverPath != "" {
			return
		} else {
			// client
			checkFlagsConfig(false, clientPath)
			log.Println("client init success")
			clientStart()
			return
		}

	}
	log.Println("start dns server ")
	checkFlagsConfig(true, serverPath)
	serverStart()

}

// checkFlagsConfig 检查配置文件是否正确
func checkFlagsConfig(isServer bool, configPath string) {
	confs, err := loadConf(configPath)
	if err != nil {
		panic(err)
		return
	}
	if isServer == true {
		// 服务器
		// 加载配置文件
		err = copyConf2Server(&server, confs)
		if err != nil {
			panic(err)
			return
		}

	} else {
		// 客户端
		err = copyConf2Client(&client, confs)
		if err != nil {
			panic(err)
			return
		}
	}
}

func copyConf2Client(client *Client, confs map[string]string) (err error) {
	// 添加默认值
	client.Root = "root"
	client.RootPassword = "123456"
	client.Username = "test"
	client.Password = "123456"
	client.NetworkCard = "eth0"
	client.DnsType = "A"
	client.LocalDomain = ""
	client.ScreenTime = 20
	client.ServerAddr = "127.0.0.1:5555"

	for key, value := range confs {
		switch key {
		case "Root":
			{
				client.Root = value
				break
			}
		case "RootPassword":
			{
				client.RootPassword = value
				break
			}
		case "Username":
			{
				client.Username = value
				break
			}
		case "Password":
			{
				client.Password = value
				break
			}
		case "ScreenTime":
			{

				client.ScreenTime, _ = strconv.Atoi(value)
				break
			}
		case "NetworkCard":
			{
				client.NetworkCard = value
				break
			}
		case "DnsType":
			{
				client.DnsType = value
				break
			}
		case "LocalDomain":
			{
				client.LocalDomain = value
				break
			}
		case "ServerAddr":
			{
				client.ServerAddr = value
				break
			}
		default:
			fmt.Println(value)
			return errors.New("配置文件出错！")

		}
	}

	// 检查参数是否填满

	return nil
}

func copyConf2Server(server *Server, confs map[string]string) (err error) {
	// 添加默认值
	server.debug = false
	server.screenTime = 20
	server.dnsPort = 53
	server.listenIP = "0.0.0.0"
	server.listenPort = 5353
	server.Domain = "localhost"
	server.dnsFile = ""
	server.user = "root"
	server.password = "123456"

	for key, value := range confs {
		switch key {
		case "dnsFile":
			{
				server.dnsFile = value
				break
			}
		case "listenIP":
			{
				server.listenIP = value
				break
			}
		case "listenPort":
			{
				server.listenPort, _ = strconv.Atoi(value)
				break
			}
		case "dnsPort":
			{
				server.dnsPort, _ = strconv.Atoi(value)
				break
			}
		case "screenTime":
			{
				//if value  {
				//	// 设置默认值
				//	server.screenTime = 20
				//}
				server.screenTime, _ = strconv.Atoi(value)
				break
			}
		case "debug":
			{
				if value == "false" {
					server.debug = false
				} else {
					server.debug = true
				}
				break
			}
		case "serverDomain":
			{
				server.Domain = value
				break
			}
		case "password":
			{
				server.password = value
				break
			}
		case "username":
			{
				server.user = value
				break
			}
		default:
			return errors.New("配置文件出错！")

		}
	}

	// 检查参数是否填满

	return nil
}

func loadConf(path string) (confs map[string]string, err error) {
	confs = make(map[string]string, 1)
	// 读取配置文件
	f, err := os.Open(path)
	if err != nil {
		log.Fatal("loadConf() run error:", err)
		return confs, err
	}
	// 关闭File句柄
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)
	// 以这个文件为参数，创建一个 scanner
	s := bufio.NewScanner(f)
	// 按行扫描文件
	for s.Scan() {
		// 如果是空行或者是#开头的行都当作注释，跳过～
		if s.Text() == "" || s.Text()[0] == '#' {
			continue
		}
		// 将空格删掉
		confLine := strings.Replace(s.Text(), " ", "", -1)
		//logs.Println(confLine)
		confLineArray := strings.Split(confLine, "=")
		//logs.Println(confLineArray, len(confLineArray))
		if len(confLineArray) != 2 {

			return confs, errors.New("config file error!")
		}

		if confLineArray[1] == "" {
			if confLineArray[0] != "LocalDomain" {
				return confs, errors.New("config file error!")
			}

		}

		confs[confLineArray[0]] = confLineArray[1]
	}
	return confs, nil
}
