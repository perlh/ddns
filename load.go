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

func loadConfig(fileName string) bool {
	// 读取配置文件
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
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
	for s.Scan() {

		//dns_file=dns.txt
		//listen_ip=0.0.0.0
		//listen_port=8050
		config.Port = -1
		recode := strings.Split(s.Text(), "=")
		switch recode[0] {
		case "dns_file":
			config.DnsFile = recode[1]
		case "listen_port":
			{
				config.Port, err = strconv.Atoi(recode[1])
			}
		case "listen_ip":
			config.LinstenIP = recode[1]

		}
		sum += 1
	}
	if config.LinstenIP == "" || config.DnsFile == "" || config.Port == -1 || sum != 3 {
		fmt.Println("配置文件不完整！")
		return false
	}
	return true
}

func loadLocalDnsFile(fileName string) bool {
	f, err := os.Open(fileName)
	if err != nil {
		CreateDnsFile(fileName)
		log.Fatal(err)

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
		var recodeO A
		recodeO.dnsType = 1
		recode := strings.Split(s.Text(), " ")
		fmt.Println(recode, recode[0], recode[1], recode[2])
		switch recode[0] {
		case "A":
			{
				recodeO.ipType = 4
				recodeO.domain = recode[1]
				var tmpByte16 [16]byte
				ipv4 := net.ParseIP(recode[2])
				//tem1 := strings.Split(recode[2], ".")
				//ip1 ,_:=
				for index1, data := range ipv4.To4() {
					//t1, _ := strconv.Atoi(data)
					tmpByte16[index1] = data
				}
				recodeO.ip = tmpByte16

				add(&DomainTypeA, recodeO)
				//recodeO.ip = [16]byte{byte(ip1)}
				//recodeO.ip = recodeO[2]
				//fmt.Printf("%T", recode[1])
				//fmt.Println(recode[1], recode[2])
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
