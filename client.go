package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func checkError(err error) bool {
	if err != nil {
		return true
	}
	return false
}

func getDomainFromServer(serverIp string, port int, ip string, ipType int, dns string) string {
	urlValues := url.Values{}
	urlValues.Add("ip", ip)
	urlValues.Add("dns", dns)
	urlValues.Add("type", strconv.Itoa(ipType))
	httpParse := "http://" + serverIp + ":" + strconv.Itoa(port) + "/dns_discover"
	//fmt.Println(httpParse)
	resp, err := http.PostForm(httpParse, urlValues)
	if err != nil {
		log.Fatal("DNS服务器不可用！")
	}
	body1, _ := ioutil.ReadAll(resp.Body)

	//fmt.Println("request:", convertHex2String2(body1))
	return convertHex2String2(body1)
}
func client() {
	flag.Parse()
	// 配置dns默认53端口
	config.DNS_Port = 53
	if *configFilePath == "" {
		// 如果没有指定配置文件
		// 上面三个参数必须配置，不能默认
		log.Fatal("未指定配置文件!")
		return
	} else {
		// 检查到存在配置文件路径，那就读取配置文件
		loadStatus := loadClientConfig(*configFilePath)
		if loadStatus == false {
			return
		}
	}

	//var dnsUpdateInterval = flag.Int("t", 5, "dnsUpdateInterval")

	//fmt.Println("-b:", *b)
	domain_ip, _ := getIPByNetworkCardName(clientConfig.Interface, clientConfig.DnsIpType)
	if clientConfig.Domain == "" {
		//fmt.Println("未找到指定域名，正在从服务器请求子域名！")
		clientConfig.Domain = getDomainFromServer(clientConfig.Server, clientConfig.HttpPort, domain_ip, clientConfig.DnsIpType, clientConfig.DnsParentDomain)
	}
	port := clientConfig.ServerPort
	server := clientConfig.Server
	//dnsUpdateInterval1 := *dnsUpdateInterval
	serverIp, err := net.ResolveIPAddr("ip", server)
	fmt.Println("******* ddns client mode *******")
	fmt.Println("网卡名称：", clientConfig.Interface)
	fmt.Println("DNS服务器：", server)
	fmt.Println("DNS服务器IP：", serverIp)
	fmt.Println("DDNS父域名：", clientConfig.DnsParentDomain)
	fmt.Println("本地IP：", domain_ip)
	fmt.Println("本地DDNS域名：", clientConfig.Domain)
	fmt.Println("******* ddns client mode *******")
	fmt.Println("")
	fmt.Println("")
	//fmt.Println("DDNS同步间隔：", dnsUpdateInterval1)

	if domain_ip == "" {
		log.Fatal("无法获取此网卡IP,请检查接口设置！")
	}

	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println("初始化")
	// 初始化，没有这条记录就要添加记录
	result := sendUpdateDomain(clientConfig.Domain, domain_ip, &net.UDPAddr{IP: serverIp.IP, Port: port}, 2)
	if result == false {
		return
	}
	// 打印dns记录表
	//fmt.Println("打印dns记录表")
	//sendUpdateDomain(*domain, domain_ip, &net.UDPAddr{IP: serverIp.IP, Port: port}, 4)
	// 间隔事件
	d := time.NewTicker(time.Duration(config.DnsUpdateTime) * time.Second)
	for {

		domainIp, _ := getIPByNetworkCardName(clientConfig.Interface, clientConfig.DnsIpType)
		result = sendUpdateDomain(clientConfig.Domain, domainIp, &net.UDPAddr{IP: serverIp.IP, Port: port}, 1)
		if result == false {
			for {
				result = sendUpdateDomain(clientConfig.Domain, domain_ip, &net.UDPAddr{IP: serverIp.IP, Port: port}, 2)
				if result {
					break
				}
				<-d.C
			}
		}
		//sendUpdateDomain(clientConfig.Domain, domainIp, &net.UDPAddr{IP: serverIp.IP, Port: port}, 4)
		//if result == false {
		//	return
		//}

		<-d.C
	}

}

func sendUpdateDomain(domain string, ip string, addr *net.UDPAddr, controller int) bool {
	//domainName := "2.d.hsm.cool."
	//fmt.Println(domainSum, subDomain)
	//domainIp := "192.168.100.1."
	domain = domain + "."
	socket, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("err:", err)
		return false
	}
	defer func(socket *net.UDPConn) {
		err := socket.Close()
		if err != nil {

		}
	}(socket)

	// 	update
	// 构建DNS首部
	sendData1 := make([]byte, 0)
	// id
	sendData1 = append(sendData1, 0x00)
	sendData1 = append(sendData1, 0x00)
	// flag
	sendData1 = append(sendData1, 0x00)
	sendData1 = append(sendData1, 0x00)
	// questions
	sendData1 = append(sendData1, 0x00)
	sendData1 = append(sendData1, 0x00)
	// answers
	sendData1 = append(sendData1, 0x00)
	sendData1 = append(sendData1, 0x00)
	//  authority rrs
	sendData1 = append(sendData1, 0x00)
	sendData1 = append(sendData1, 0x00)
	// additional rrs
	sendData1 = append(sendData1, 0x00)
	sendData1 = append(sendData1, 0x00)

	// Queries
	domainSum, subDomain := getDomainNum(domain)
	for i1 := 0; i1 < domainSum; i1++ {
		sendData1 = append(sendData1, 0x01)
		sendData1 = append(sendData1, []byte(subDomain[i1])...)
	}
	// Queries 结尾一定要0x00
	sendData1 = append(sendData1, 0x00)

	/*
			type
			00,01 表示A
			00,1C 表示AAAA
			00,EE 表示私有的控制字段
		// 记录表
		类型	助记符	说明
		1	A	由域名获得IPv4地址
		2	NS	查询域名服务器
		5	CNAME	查询规范名称
		6	SOA	开始授权
		11	WKS	熟知服务
		12	PTR	把IP地址转换成域名
		13	HINFO	主机信息
		15	MX	邮件交换
		28	AAAA	由域名获得IPv6地址
		252	AXFR	传送整个区的请求
		255	ANY	对所有记录的请求
	*/
	sendData1 = append(sendData1, 0x00)
	sendData1 = append(sendData1, 0xee)
	//查询类：通常为1，表明是Internet数据
	sendData1 = append(sendData1, 0x00)
	sendData1 = append(sendData1, 0x01)

	// 后面就是可以添加自己的功能
	if controller == 1 {
		/*
			控制类型
			0x01 更新指定域名的dns
		*/
		sendData1 = append(sendData1, 0x01)
		// 保留字段
		sendData1 = append(sendData1, 0x00)
		// ip类型，如果是ipv4，就是0x04,ipv6就是ipv6
		ip1, ip1_type := checkIPType(ip)
		if ip1_type == 4 {
			sendData1 = append(sendData1, 0x04)
			sendData1 = append(sendData1, ip1.To4()[0])
			sendData1 = append(sendData1, ip1.To4()[1])
			sendData1 = append(sendData1, ip1.To4()[2])
			sendData1 = append(sendData1, ip1.To4()[3])
		} else if ip1_type == 6 {
			sendData1 = append(sendData1, 0x16)
			sendData1 = append(sendData1, ip1.To16()[0])
			sendData1 = append(sendData1, ip1.To16()[1])
			sendData1 = append(sendData1, ip1.To16()[2])
			sendData1 = append(sendData1, ip1.To16()[3])
			sendData1 = append(sendData1, ip1.To16()[4])
			sendData1 = append(sendData1, ip1.To16()[5])
			sendData1 = append(sendData1, ip1.To16()[6])
			sendData1 = append(sendData1, ip1.To16()[7])
			sendData1 = append(sendData1, ip1.To16()[8])
			sendData1 = append(sendData1, ip1.To16()[9])
			sendData1 = append(sendData1, ip1.To16()[10])
			sendData1 = append(sendData1, ip1.To16()[11])
			sendData1 = append(sendData1, ip1.To16()[12])
			sendData1 = append(sendData1, ip1.To16()[13])
			sendData1 = append(sendData1, ip1.To16()[14])
			sendData1 = append(sendData1, ip1.To16()[15])
		} else {
			return false
		}

		_, err = socket.Write(sendData1) // 发送数据
		if checkError(err) {
			//fmt.Println(time.Now().Format("2006/1/2 15:04:05"), "发送数据失败，err:", err)
			log.Println("发送数据失败，err:", err)
			return false
		}

		// 接收服务端消息
		var recive [512]byte
		_, _, err := socket.ReadFromUDP(recive[:])
		if checkError(err) {
			log.Printf("read failed,err:%v \n", err)
			return false

		}
		if recive[0] == 0x00 {
			//fmt.Println(time.Now().Format("2006/1/2 15:04:05"), "更新ip成功")
			log.Println("发送更新DNS请求成功", domain, ip)
		} else {
			//fmt.Println(time.Now().Format("2006/1/2 15:04:05"), "更新ip失败")
			log.Println("发送更新DNS请求失败", domain, ip)
			return false
		}
		return true

	}
	if controller == 2 {

		/*
			0x02 添加指定域名的dns
		*/
		sendData1 = append(sendData1, 0x02)
		// 保留字段
		sendData1 = append(sendData1, 0x00)
		// ip类型，如果是ipv4，就是0x04,ipv6就是ipv6
		ip1, ip1_type := checkIPType(ip)
		//fmt.Println(ip1_type)
		//fmt.Println("添加记录")
		if ip1_type == 4 {
			//fmt.Printf("%v", )
			sendData1 = append(sendData1, 0x04)
			sendData1 = append(sendData1, ip1.To4()[0])
			sendData1 = append(sendData1, ip1.To4()[1])
			sendData1 = append(sendData1, ip1.To4()[2])
			sendData1 = append(sendData1, ip1.To4()[3])
			//fmt.Println(sendData1)
		} else if ip1_type == 6 {
			fmt.Println("添加ipv6")
			sendData1 = append(sendData1, 0x16)
			sendData1 = append(sendData1, ip1.To16()[0])
			sendData1 = append(sendData1, ip1.To16()[1])
			sendData1 = append(sendData1, ip1.To16()[2])
			sendData1 = append(sendData1, ip1.To16()[3])
			sendData1 = append(sendData1, ip1.To16()[4])
			sendData1 = append(sendData1, ip1.To16()[5])
			sendData1 = append(sendData1, ip1.To16()[6])
			sendData1 = append(sendData1, ip1.To16()[7])
			sendData1 = append(sendData1, ip1.To16()[8])
			sendData1 = append(sendData1, ip1.To16()[9])
			sendData1 = append(sendData1, ip1.To16()[10])
			sendData1 = append(sendData1, ip1.To16()[11])
			sendData1 = append(sendData1, ip1.To16()[12])
			sendData1 = append(sendData1, ip1.To16()[13])
			sendData1 = append(sendData1, ip1.To16()[14])
			sendData1 = append(sendData1, ip1.To16()[15])
		} else {
			return false
		}
		//fmt.Printf("添加指定域名的dns")
		_, err = socket.Write(sendData1) // 发送数据
		if err != nil {
			fmt.Println(time.Now().Format("2006/1/2 15:04:05"), "发送数据失败，err:", err)
			return false
		}
		// 接收服务端消息
		var recive [512]byte
		_, _, err := socket.ReadFromUDP(recive[:])
		if err != nil {
			fmt.Printf("read failed,err:%v \n", err)
			return false

		}
		if recive[0] == 0x00 {

			//fmt.Println(time.Now().Format("2006/1/2 15:04:05"), )
			//break
			log.Println("域名绑定ip成功", domain, ip)
		} else {
			log.Println("域名绑定ip失败", domain, ip)
			//fmt.Println(time.Now().Format("2006/1/2 15:04:05"), "绑定ip失败")
			return false
		}
		return true
	}
	if controller == 3 {
		/*
			0x03删除指定的域名
		*/
		sendData1 = append(sendData1, 0x03)
		// 保留字段
		sendData1 = append(sendData1, 0x00)
		// 发送请求包
		_, err = socket.Write(sendData1) // 发送数据
		if err != nil {
			fmt.Println(time.Now().Format("2006/1/2 15:04:05"), "发送数据失败，err:", err)
			return false
		}

		// 接收服务端消息
		var recive [512]byte
		_, _, err := socket.ReadFromUDP(recive[:])
		if err != nil {
			fmt.Printf("read failed,err:%v", err)

		}
		if recive[0] == 0x00 {
			fmt.Println(time.Now().Format("2006/1/2 15:04:05"), "删除成功")
		} else {
			fmt.Println(time.Now().Format("2006/1/2 15:04:05"), "删除失败")
			return false
		}
		//fmt.Printf("status：%v\n", string(recive[:n]))
		return true

	}
	if controller == 4 {
		/*
			0x04 使服务器显示dns记录列表
		*/
		sendData1 = append(sendData1, 0x04)
		// 保留字段
		sendData1 = append(sendData1, 0x00)
		_, err = socket.Write(sendData1) // 发送数据
		if err != nil {
			fmt.Println(time.Now().Format("2006/1/2 15:04:05"), "发送数据失败，err:", err)
			return false
		}
		//	接收服务端消息
		var recive [512]byte
		_, _, err := socket.ReadFromUDP(recive[:])
		if err != nil {
			fmt.Printf("read failed,err:%v\n", err)

		}
		if recive[0] == 0x00 {
			//fmt.Println(time.Now().Format("2006/1/2 15:04:05"), "查看成功")
			//log.Println("发送DNS查看请求成功", addr.String())
		} else {
			fmt.Println(time.Now().Format("2006/1/2 15:04:05"), "查看失败")
			log.Println("发送DNS查看请求失败", addr.String())
			return false
		}
		return true
	}

	return false
}

func getDomainNum(domain string) (int, []string) {
	/*

	 */
	var subDomain []string
	tmp := ""
	sum := 0
	for i := 0; i < len(domain); i++ {
		if domain[i] == '.' {
			subDomain = append(subDomain, tmp)
			tmp = ""
			sum += 1
			continue
		}

		tmp = tmp + string(domain[i])
	}
	return sum, subDomain
}

func getIPByNetworkCardName(networkCard string, ipType int) (ipv4Address string, err error) {
	//获取网卡信息
	addrs, err := net.InterfaceByName(networkCard)
	//ipNet, isIpNet := addrs.Addrs()
	ipNet, _ := addrs.Addrs()
	//fmt.Println(isIpNet)
	//var ip2 [16]byte
	for _, ipInfo := range ipNet {
		//fmt.Println(index)
		ip2, type2 := checkIPType(ipInfo.(*net.IPNet).IP.String())
		if type2 == ipType {

			//fmt.Println(string(ip2.String()))
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
