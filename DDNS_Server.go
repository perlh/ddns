package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

var mu sync.RWMutex

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

type A struct {
	domain string
	ip     [16]byte
	ipType int
	ttl    int
}

var DomainTypeA []A

func add(DomainTypeA *[]A, data A) bool {
	mu.Lock()
	defer mu.Unlock()
	*DomainTypeA = append(*DomainTypeA, A{domain: data.domain, ip: data.ip, ipType: data.ipType, ttl: data.ttl})
	return true
}
func del(DomainTypeA *[]A, domain string) bool {
	mu.Lock()
	defer mu.Unlock()
	var tmp1 []A
	for i := 0; i < len(*DomainTypeA); i++ {
		if (*DomainTypeA)[i].domain == domain {
			continue
		}
		tmp1 = append(tmp1, (*DomainTypeA)[i])
	}
	*DomainTypeA = tmp1
	return true
}
func updateTtl(DomainTypeA *[]A, domain string, isAdd bool) bool {
	mu.Lock()
	defer mu.Unlock()
	var tmp1 []A
	for i := 0; i < len(*DomainTypeA); i++ {
		if (*DomainTypeA)[i].domain == domain {

			if isAdd {
				(*DomainTypeA)[i].ttl++
			} else {
				(*DomainTypeA)[i].ttl--
			}

			tmp1 = append(tmp1, (*DomainTypeA)[i])
			continue
		}
		tmp1 = append(tmp1, (*DomainTypeA)[i])
	}
	*DomainTypeA = tmp1
	return true
}
func checkDNS() {
	d := time.NewTicker(time.Minute * 5)
	for {
		for _, data := range DomainTypeA {
			// 检查这条dns是否存活
			if data.ttl <= 0 {
				// 如果这条域名长期不更新，那么删除这条记录
				del(&DomainTypeA, data.domain)
			} else {
				// 存活的话，让其ttl减1
				updateTtl(&DomainTypeA, data.domain, false)
			}
		}
		<-d.C
	}

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

func updateDNSByLocalFile() {

}
