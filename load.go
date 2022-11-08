package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func loadClientConfig(fileName string) bool {
	// 读取配置文件

	f, err := os.Open(fileName)
	if err != nil {

		//panic(err)
		log.Fatal("运行错误！找不到该配置文件：", fileName)
		return false

	}
	// 关闭File句柄
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)
	// 以这个文件为参数，创建一个 scanner
	s := bufio.NewScanner(f)
	sum := 0
	// 扫描每行文件，按行读取
	clientConfig.ServerPort = -1
	config.DnsUpdateTime = 5
	for s.Scan() {
		if s.Text() == "" || s.Text()[0] == '#' {
			continue
		}
		recode := strings.Split(s.Text(), "=")
		switch recode[0] {
		case "interface":
			clientConfig.Interface = recode[1]
		case "local-domain":
			clientConfig.Domain = recode[1]
		case "server":
			clientConfig.Server = recode[1]
		case "port":
			clientConfig.ServerPort, _ = strconv.Atoi(recode[1])
		case "http-port":
			clientConfig.HttpPort, _ = strconv.Atoi(recode[1])
		case "dns-parent-domain":
			clientConfig.DnsParentDomain = recode[1]
		case "update-time":
			config.DnsUpdateTime, err = strconv.Atoi(recode[1])
		case "ip-type":
			{
				if recode[1] == "ipv4" {
					clientConfig.DnsIpType = 4
				} else if recode[1] == "ipv6" {
					clientConfig.DnsIpType = 6
				} else {
					log.Fatal("ip类型输入错误！")
				}
			}

		}
		sum += 1
	}
	// 判断
	//fmt.Println(clientConfig.ServerPort, sum)
	if clientConfig.ServerPort == -1 {
		//fmt.Println()
		log.Fatal("配置文件不完整！")
		return false
	}
	return true
}
func loadConfig(fileName string) bool {
	// 读取配置文件
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal("运行错误！找不到该配置文件：", fileName)
		return false

	}
	// 关闭File句柄
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)
	// 以这个文件为参数，创建一个 scanner
	s := bufio.NewScanner(f)
	sum := 0
	// 扫描每行文件，按行读取
	config.HTTP_Port = -1
	config.DnsUpdateTime = 5
	for s.Scan() {

		if s.Text() == "" || s.Text()[0] == '#' {
			continue
		}
		recode := strings.Split(s.Text(), "=")
		switch recode[0] {
		case "dns_file":
			config.DnsFile = recode[1]
		case "http_port":
			{
				config.HTTP_Port, err = strconv.Atoi(recode[1])
			}
		case "listen_ip":
			config.LinstenIP = recode[1]
		case "debug":
			if recode[1] == "true" {
				debug = true
			} else {
				debug = false
			}
		case "dns_port":
			config.DNS_Port, err = strconv.Atoi(recode[1])
		case "update-time":
			config.DnsUpdateTime, err = strconv.Atoi(recode[1])
		}
		sum += 1
	}
	//fmt.Println(config, sum)
	if config.LinstenIP == "" || config.DnsFile == "" || config.HTTP_Port == -1 || sum != 6 {
		//fmt.Println()
		log.Fatal("配置文件不完整！")
		return false
	}
	return true
}

func loadLocalDnsFile(fileName string) bool {
	f, err := os.Open(fileName)
	if err != nil {
		//
		fmt.Println("未找到dns记录文件!")
		CreateDnsFile(fileName)
		fmt.Printf("自动创建dns记录文件：%v/%v\n", getCurrentAbPathByExecutable(), fileName)
		//f, _ = os.Open(fileName)
		//log.Fatal(err)
		return true
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	// 以这个文件为参数，创建一个 scanner
	s := bufio.NewScanner(f)
	// 扫描每行文件，按行读取
	for s.Scan() {
		// 过滤注释
		if s.Text() == "" || s.Text()[0] == '#' {
			continue
		}
		var recodeO A
		recodeO.dnsType = 1
		recode := strings.Split(s.Text(), " ")
		//fmt.Println(recode, recode[0], recode[1], recode[2])
		switch recode[0] {
		case "A":
			{
				recodeO.ipType = 4
				recodeO.domain = recode[1]
				var tmpByte16 [16]byte
				ipv4 := net.ParseIP(recode[2])
				for index1, data := range ipv4.To4() {
					tmpByte16[index1] = data
				}
				recodeO.ip = tmpByte16
				//fmt.Println(recodeO)
				add(&DomainTypeA, recodeO)
				//fmt.Println(DomainTypeA)
			}
		case "AAAA":
			{
				recodeO.ipType = 6
				recodeO.domain = recode[1]
				var tmpByte16 [16]byte
				ipv6 := net.ParseIP(recode[2])
				for index1, data := range ipv6.To16() {
					//t1, _ := strconv.Atoi(data)
					tmpByte16[index1] = data
				}
				recodeO.ip = tmpByte16
				add(&DomainTypeA, recodeO)
			}
		}

	}
	//fmt.Println(DomainTypeA)
	err = s.Err()
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}

func CreateDnsFile(filePath string) bool {
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
	}
	//file.Close()
	// 关流(不关流会长时间占用内存)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)
	return true
}
