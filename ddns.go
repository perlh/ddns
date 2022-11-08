package main

import (
	"sync"
	"time"
)

// 互斥变量
var mu sync.RWMutex

type A struct {
	domain  string
	ip      [16]byte
	ipType  int // 4 or 6
	ttl     int // >0
	dnsType int //1 本地dns记录，0 ddns
}

var DomainTypeA []A

func add(DomainTypeA *[]A, data A) bool {
	mu.Lock()
	defer mu.Unlock()
	//fmt.Println("rss:", data)
	*DomainTypeA = append(*DomainTypeA, A{domain: data.domain, ip: data.ip, ipType: data.ipType, ttl: data.ttl, dnsType: data.dnsType})
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
	//var tmp1 []A

	for i := 0; i < len(*DomainTypeA); i++ {
		if (*DomainTypeA)[i].domain == domain && (*DomainTypeA)[i].dnsType == 0 {
			//fmt.Println("xxxxx")
			if isAdd {

				if (*DomainTypeA)[i].ttl < 5 {
					/*
						limit add many
					*/
					(*DomainTypeA)[i].ttl++
				}
			} else {
				(*DomainTypeA)[i].ttl--
			}
			return true
		}
	}

	return true
}

func checkDNS() {
	//a := 4
	//config.DnsUpdateTime = 1
	//v1:= time.Minute * config.DnsUpdateTime

	d := time.NewTicker(time.Duration(config.DnsUpdateTime) * time.Second)

	for {

		recodeA := DomainTypeA
		//fmt.Println(recodeA)
		for _, data := range recodeA {
			if data.dnsType == 1 {
				// 本地dns记录不检查
				continue
			}
			//fmt.Println("检查这条dns是否存活")
			// 检查这条dns是否存活
			if data.ttl <= 0 {
				// 如果这条域名长期不更新，那么删除这条记录
				//fmt.Println("delete", data.domain)
				del(&DomainTypeA, data.domain)
			} else {
				// 存活的话，让其ttl减1
				updateTtl(&DomainTypeA, data.domain, false)
			}
		}
		<-d.C
	}

}
