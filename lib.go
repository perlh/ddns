package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

// 检查域名是否在域名表中
func checkDomainExit(table []A, domain string) bool {
	for i := 0; i < len(table); i++ {
		if table[i].domain == domain {
			return true
		}
	}
	return false
}

// 检查域名和ip是否在域名表中
func checkDomainExitAndIpNotExit(table []A, domain string, Ip [16]byte) int {
	/*
		0 域名和ip都不在表中
		1 域名在表中，ip不在
		2 域名和ip都在表中
	*/
	for i := 0; i < len(table); i++ {
		if table[i].domain == domain {
			fmt.Println(table[i].domain, domain, table[i].ip, Ip)
			if table[i].ip == Ip {
				return 2
			}
			return 1
		}
	}
	return 0
}

func convertHex2String(domainHex []byte) string {
	//fmt.Println("---")
	length := len(domainHex)
	var returnDomain string
	//tmp_lable := ""
	for i := 1; i < length; i++ {
		if int(domainHex[i]) > 10 {
			returnDomain = returnDomain + string(domainHex[i])
		}
		if int(domainHex[i]) < 10 {
			returnDomain = returnDomain + "."
			//tmp_lable = ""
		}

	}
	//fmt.Println(returnDomain)
	return returnDomain
}

func getIpv4(ipv4Byte []byte) string {
	ipv4 := ""
	for index, char2 := range ipv4Byte[0:4] {
		ipv4 += strconv.Itoa(int(char2))
		if index < 3 {
			ipv4 = ipv4 + "."
		}

	}

	//fmt.Println(ipv4)
	return ipv4
}

func getIpv6(ipv6Byte []byte) string {
	ipv6 := ""
	for index, char1 := range hex.EncodeToString(ipv6Byte[0:16]) {
		ipv6 += string(char1)
		//fmt.Println(index)
		if index%4 == 3 && index != 31 {
			ipv6 += ":"
		}
	}

	return net.ParseIP(ipv6).To16().String()
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
	// 关流(不关流会长时间占用内存)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)
	return true
}
