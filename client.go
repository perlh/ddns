package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

func RegisterUser() (ok bool) {
	type Response struct {
		JsonResult
		User
	}
	var response Response
	url1 := "http://" + client.ServerAddr + "/register"
	//encodeToken(user,passwd)
	token := encodeToken(client.Root, client.RootPassword)
	resp, err := http.PostForm(url1,
		url.Values{"user": {client.Username}, "password": {client.Password}, "root_token": {token}})
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("http code : %d\n", resp.StatusCode)
		return false
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("err2")
	}
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		return false
	}
	//fmt.Println(string(body))
	if response.Code == 200 {
		return true
	}
	return false
}

func Update(domain string, dnsType string, value string) (ok bool) {
	type Response struct {
		JsonResult
		Domain  string `json:"domain"`
		DnsType string `json:"dns_type"`
		Value   string `json:"value"`
	}
	var response Response
	url1 := "http://" + client.ServerAddr + "/update"
	token := encodeToken(client.Username, client.Password)
	resp, err := http.PostForm(url1,
		url.Values{"domain": {domain}, "dnsType": {dnsType}, "value": {value}, "token": {token}})
	if err != nil {
		// handle error
		//log.Println("http请求发送失败！")
		ok = false
		return ok
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("http code : %d\n", resp.StatusCode)
		return false
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("err2")
	}
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		log.Println("反序列化失败")
		return false
	}
	//log.Println(response)
	//log.Println()
	if response.Code == 200 {
		return true
	}
	return false
}

func CreateRandDomain(dnsType string, value string, time string) (ok bool) {
	type Response struct {
		JsonResult
		Domain string `json:"domain"`
	}
	var response Response
	url1 := "http://" + client.ServerAddr + "/create_domain"
	token := encodeToken(client.Username, client.Password)
	resp, err := http.PostForm(url1,
		url.Values{"dns_type": {dnsType}, "time": {time}, "value": {value}, "token": {token}})
	if err != nil {
		// handle error
		//log.Println("http请求发送失败！")
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("http code : %d\n", resp.StatusCode)
		return false
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("err2")
	}
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		log.Println("反序列化失败")
		return false
	}
	//log.Println(response)
	if response.Code == 200 {
		client.LocalDomain = response.Domain
		return true
	}
	return false
}

func clientStart() {
	//log.Println(client.DnsType)

	if client.DnsType == "A" || client.DnsType == "AAAA" {

		var DnsIpType int
		if client.DnsType == "A" {
			DnsIpType = 4
		} else if client.DnsType == "AAAA" {
			DnsIpType = 6
		}
		domainIp, _ := getIPByNetworkCardName(client.NetworkCard, DnsIpType)
		client.ip = net.ParseIP(domainIp)
		ok := RegisterUser()
		if ok {
			if client.LocalDomain != "" {
				ok = Update(client.LocalDomain, client.DnsType, client.ip.String())
				if ok {
					log.Println("更新成功！", client.LocalDomain, client.DnsType, client.ip.String())

				} else {
					log.Println("无法添加域名！")
					return
					//log.Println("dsfaasdf")
				}
			} else {
				// 自动获取ip
				ok = CreateRandDomain(client.DnsType, client.ip.String(), "60")
				if ok {
					log.Println("获取域名成功！", client.LocalDomain, client.DnsType, client.ip.String())
				} else {
					//RegisterUser()
					log.Println("获取域名失败！")
					return
				}

			}

		}

	} else if client.DnsType == "CNAME" {
		//TODO
	}

	if client.DnsType == "A" || client.DnsType == "AAAA" {
		var DnsIpType int
		if client.DnsType == "A" {
			DnsIpType = 4
		} else if client.DnsType == "AAAA" {
			DnsIpType = 6
		}
		d := time.NewTicker(time.Duration(client.ScreenTime) * time.Second)
		firstError := false
		for {
			//log.Println(firstError)
			domainIp, _ := getIPByNetworkCardName(client.NetworkCard, DnsIpType)
			client.ip = net.ParseIP(domainIp)
			ok := Update(client.LocalDomain, client.DnsType, client.ip.String())
			if ok {
				//log.Println("更新成功！", client.LocalDomain, client.DnsType, client.ip.String())
				<-d.C
				continue
			}
			if firstError == false {
				log.Println("服务器断开，尝试重新连接！")
				firstError = true
			} else {
				ok = RegisterUser()
				if ok {
					log.Println("重新连接成功,本地域名：", client.LocalDomain, client.ip)
					firstError = false
					<-d.C
					continue
				}
			}

			<-d.C
		}

	}

}

func getIPByNetworkCardName(networkCard string, ipType int) (ipv4Address string, err error) {
	//获取网卡信息
	addrs, err := net.InterfaceByName(networkCard)
	ipNet, _ := addrs.Addrs()
	//var ip2 [16]byte
	for _, ipInfo := range ipNet {
		//fmt.Println(index)
		ip2, type2 := checkIPType(ipInfo.(*net.IPNet).IP.String())
		if type2 == ipType {

			return ip2.String(), nil
		}
	}
	// 如果网卡只有一个ip，那么不管是ipv4还是ipv6，返回这个ip
	if len(ipNet) == 1 {
		for _, ipInfo := range ipNet {
			//fmt.Println(index)
			ip2, _ := checkIPType(ipInfo.(*net.IPNet).IP.String())
			return ip2.String(), nil
		}
	}
	return
}

// 检查ip的合法性，并且给出ip类型
func checkIPType(s string) (net.IP, int) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, 0
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return ip, 4
		case ':':
			return ip, 6
		}
	}
	return nil, 0
}
