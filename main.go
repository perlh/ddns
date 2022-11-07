package main

import (
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

//startHttp

type Config struct {
	LinstenIP string
	Port      int
	DnsFile   string
}

var (
	listenIp       = flag.String("ip", "", "Dns server Listen Ip")
	listenPort     = flag.Int("port", -1, "DNS server Port")
	configFilePath = flag.String("c", "", "Config file  path")
	dnsFilePath    = flag.String("d", "", "DNS local File recode ")

	config Config
)

// UDP server端
func main() {
	flag.String("example", "", "./ddns -d dns.txt -ip 0.0.0.0 -port 8050 or ./ddns -c ddns.conf")
	flag.Parse()
	if *configFilePath == "" {
		// 如果没有指定配置文件
		if *listenIp == "" || *dnsFilePath == "" || *listenPort == -1 {
			log.Fatal("未指定配置文件!")
			return
		} else {
			config.DnsFile = *dnsFilePath
			config.Port = *listenPort
			config.LinstenIP = *listenIp
		}

	} else {
		// 检查到存在配置文件路径，那就读取配置文件
		loadStatus := loadConfig(*configFilePath)
		if loadStatus == false {
			return
		}
	}
	dnsFile := config.DnsFile
	ipAddr := net.ParseIP(config.LinstenIP)
	// 从本地文件中导入静态dns记录
	loadLocalDnsFile(dnsFile)
	listen, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   ipAddr.To4(),
		Port: config.Port,
	})
	if err != nil {
		fmt.Println("listen failed, please check Port or other issue!", err)
		return
	}

	defer func(listen *net.UDPConn) {
		err := listen.Close()
		if err != nil {

		}
	}(listen)
	var data [512]byte

	// 设置一个定时器
	// auth
	authByte := []byte{0xc0, 0x0f, 0x00, 0x06, 0x00, 0x01, 0x00, 0x00, 0x00, 0xb4, 0x00, 0x41, 0x07, 0x6d, 0x61, 0x72, 0x74, 0x69, 0x6e, 0x69, 0x06, 0x64, 0x6e, 0x73, 0x70, 0x6f, 0x64, 0x03, 0x6e, 0x65, 0x74, 0x00, 0x0c, 0x66, 0x72, 0x65, 0x65, 0x64, 0x6e, 0x73, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x06, 0x64, 0x6e, 0x73, 0x70, 0x6f, 0x64, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x63, 0x4b, 0x72, 0x54, 0x00, 0x00, 0x0e, 0x10, 0x00, 0x00, 0x00, 0xb4, 0x00, 0x12, 0x75, 0x00, 0x00, 0x00, 0x00, 0xb4}
	go checkDNS()
	go startHttp()

	log.Printf("ddns server:%v:%v", config.LinstenIP, config.Port)
	//log.Println("server Port:", config.Port)
	log.Println("server local dns file path:", config.DnsFile)
	for {

		//每个DNS数据包限制在512字节之内(防止IP包超过MTU被碎片化)，
		n, addr, err := listen.ReadFromUDP(data[:]) // 接收数据
		//fmt.Println("dns服务开启成功！")
		//fmt.Println("源：", hex.EncodeToString(data[:n]))
		var end int

		for end = 12; end < 512; end++ {
			if int(data[end]) == 0 {
				break
			}
		}
		Domain := data[12:end]
		// dnslog
		domainLable := convertHex2String(Domain)

		domainLable += "\n"

		dnslogRecode = append(dnslogRecode, domainLable)
		DnsType := binary.BigEndian.Uint16(data[end+1 : end+3])
		if DnsType == 0x0001 {
			// A类型
			isSuccess := false
			domainLable := convertHex2String(Domain)
			for j := 0; j < len(DomainTypeA); j++ {
				// 寻找域名
				if domainLable == DomainTypeA[j].domain {
					// FLAG
					data[2] = 0x81
					data[3] = 0x80
					data[6] = 0x00
					data[7] = 0x01
					data[end+5] = 0xc0
					data[end+6] = 0x0c
					// type
					data[end+7] = 0x00
					data[end+8] = 0x01 // A
					// class
					data[end+9] = 0x00
					data[end+10] = 0x01

					// time
					data[end+11] = 0x00
					data[end+12] = 0x00
					data[end+13] = 0x00
					data[end+14] = 0x2f
					if DomainTypeA[j].ipType == 4 {
						// data length
						data[end+15] = 0x00
						data[end+16] = 0x04
						// ipv4
						for i := 0; i < 4; i++ {
							data[end+17+i] = DomainTypeA[j].ip[i]
						}
						_, err = listen.WriteToUDP(data[:end+21], addr) // 发送数据
						if err != nil {
							fmt.Println("err:", err)
							continue
						}

						fmt.Println(time.Now().Format("2006/1/2 15:04:05"), addr, "查询DNS成功", domainLable, getIpv4(data[end+17:end+21]))
						// 查询成功，退出循环
						isSuccess = true
						break
					}

				}

			}
			if isSuccess == true {
				continue
			}
		}
		if DnsType == 0x001c {
			// AAAA类型
			// A类型
			isSuccessAAAA := false
			domainLable := convertHex2String(Domain)
			for j := 0; j < len(DomainTypeA); j++ {
				if domainLable == DomainTypeA[j].domain {
					// 如果匹配到了
					if DomainTypeA[j].ipType == 6 {
						// FLAG
						data[2] = 0x81
						data[3] = 0x80
						//
						data[6] = 0x00
						data[7] = 0x01

						data[end+5] = 0xc0
						data[end+6] = 0x0c
						// type
						data[end+7] = 0x00
						data[end+8] = 0x1c // AAAA
						// class
						data[end+9] = 0x00
						data[end+10] = 0x01

						// time
						data[end+11] = 0x00
						data[end+12] = 0x00
						data[end+13] = 0x00
						data[end+14] = 0x2f
						// ipv6
						// data length
						data[end+15] = 0x00
						data[end+16] = 0x10
						// ipv6
						for i := 0; i < 16; i++ {
							data[end+17+i] = DomainTypeA[j].ip[i]
						}

						_, err = listen.WriteToUDP(data[:end+33], addr) // 发送数据
						if err != nil {
							fmt.Println("err:", err)
							continue
						}
						fmt.Println(time.Now().Format("2006/1/2 15:04:05"), addr, "查询DNS成功!", domainLable, getIpv6(data[end+17:end+33]))
						isSuccessAAAA = true
						break
					}
				}
			}
			if isSuccessAAAA == true {
				continue
			}

		}
		// 控制记录
		if DnsType == 0x00ee {
			isSuccessControl := false
			// 管理域名
			control := data[end+5]
			// 更新ip
			if control == 0x01 {
				domainLable := convertHex2String(Domain)
				if !checkDomainExit(DomainTypeA, domainLable) {
					_, err = listen.WriteToUDP([]byte{0x01}, addr) // 发送数据
					if err != nil {
						fmt.Println("err:", err)
						continue
					}
					continue
				}
				domainLable1 := convertHex2String(Domain)
				for j := 0; j < len(DomainTypeA); j++ {
					if domainLable1 == DomainTypeA[j].domain {
						updateTtl(&DomainTypeA, domainLable1, true)
						if data[end+7] == 0x04 {
							for i := 0; i < 4; i++ {
								DomainTypeA[j].ip[i] = data[end+8+i]
							}
							isSuccessControl = true
							fmt.Println(time.Now().Format("2006/1/2 15:04:05"), addr, "更新DNS记录成功，域名：", domainLable1, "新IP:", getIpv4(data[end+8:end+12]))
						} else if data[end+7] == 0x16 {
							for i := 0; i < 16; i++ {
								DomainTypeA[j].ip[i] = data[end+8+i]
							}
							ipv6 := ""
							for index, char1 := range hex.EncodeToString(data[end+8 : end+24]) {
								ipv6 += string(char1)
								//fmt.Println(index)
								if index%4 == 3 && index != 31 {

									ipv6 += ":"
								}
							}
							//fmt.Println(ipv6)
							isSuccessControl = true
							fmt.Println(time.Now().Format("2006/1/2 15:04:05"), addr, "更新DNS记录成功,域名:", domainLable1, ",", "IP:", net.ParseIP(ipv6).To16())
						}

						_, err = listen.WriteToUDP([]byte{0x00}, addr) // 发送数据
						if err != nil {
							fmt.Println("err:", err)
							continue
						}
						continue
					}
				}

			}
			// 添加记录
			if control == 0x02 {
				fmt.Println("add data")
				domainLable1 := convertHex2String(Domain)
				var ipTmp [16]byte
				dnsData := A{domain: domainLable1, ip: [16]byte{0x00, 0x01, 0x02, 0x03}, ipType: 4, ttl: 1}
				if data[end+7] == 0x04 {
					for i := 0; i < 4; i++ {
						dnsData.ip[i] = data[end+8+i]
						ipTmp[i] = data[end+8+i]
					}

					checkDnsAdd := checkDomainExitAndIpNotExit(DomainTypeA, domainLable1, ipTmp)
					if checkDnsAdd == 0 {
						// 域名不在表中，ip也不在表中
						dnsData.dnsType = 0
						add(&DomainTypeA, dnsData)

					} else if checkDnsAdd == 1 {
						// 域名在表中，ip不在表中
						//无法添加记录
						_, err = listen.WriteToUDP([]byte{0x01}, addr) // 添加失败
						if err != nil {
							fmt.Println("err:", err)
						}
						continue
					} else if checkDnsAdd == 2 {
						updateTtl(&DomainTypeA, domainLable1, true)
					}

					isSuccessControl = true

					fmt.Println(time.Now().Format("2006/1/2 15:04:05"), addr, "添加解析记录成功，", domainLable1, "新IP:", getIpv4(data[end+8:end+12]))
				} else if data[end+7] == 0x16 {
					for i := 0; i < 16; i++ {
						dnsData.ip[i] = data[end+8+i]
						ipTmp[i] = data[end+8+i]
					}
					//dnsData.ipType = 6
					//DomainTypeA = append(DomainTypeA, dnsData)
					checkDnsAdd := checkDomainExitAndIpNotExit(DomainTypeA, domainLable1, ipTmp)
					fmt.Println("checkDnsAdd", checkDnsAdd)
					if checkDnsAdd == 0 {
						// 域名不在表中，ip也不在表中
						dnsData.ipType = 6
						DnsType = 0
						add(&DomainTypeA, dnsData)
					} else if checkDnsAdd == 1 {
						// 域名在表中，ip不在表中
						//无法添加记录
						//fmt.Println("11111111")
						_, err = listen.WriteToUDP([]byte{0x01}, addr) // 添加失败
						if err != nil {
							fmt.Println("err:", err)
						}
						continue
					} else if checkDnsAdd == 2 {
						updateTtl(&DomainTypeA, domainLable1, true)
					}

					isSuccessControl = true
					//add(&DomainTypeA, dnsData)
					fmt.Println(time.Now().Format("2006/1/2 15:04:05"), addr, "添加DNS记录成功,域名:", domainLable1, ",", "IP:", getIpv6(data[end+8:end+24]))
				}
				// 添加成功
				_, err = listen.WriteToUDP([]byte{0x00}, addr) // 发送数据
				if err != nil {
					fmt.Println("err:", err)
					continue
				}
				continue

			}
			// 删除记录
			if control == 0x03 {
				//var tmp1 []A
				domainLable1 := convertHex2String(Domain)
				fmt.Println(domainLable1, checkDomainExit(DomainTypeA, domainLable1))
				if !checkDomainExit(DomainTypeA, domainLable1) {
					_, err = listen.WriteToUDP([]byte{0x01}, addr) // 发送数据
					if err != nil {
						fmt.Println("err:", err)
						continue

					}
					continue
				}
				// 通过指定域名删除dns记录
				del(&DomainTypeA, domainLable1)
				_, err = listen.WriteToUDP([]byte{0x00}, addr) // 删除成功
				if err != nil {
					fmt.Println("err:", err)
					continue

				}
				isSuccessControl = true
				fmt.Println(time.Now().Format("2006/1/2 15:04:05"), addr, "删除DNS解析记录成功：", domainLable1)
				continue
			}
			// 查找记录
			if control == 0x04 {
				fmt.Printf(time.Now().Format("2006/1/2 15:04:05") + " " + addr.String() + " 查询DNS解析记录成功!\n")
				for _, content := range DomainTypeA {
					if content.ipType == 4 {
						ipv4 := getIpv4(content.ip[0:4])
						fmt.Printf("\t\t" + content.domain + "  " + ipv4 + " " + strconv.Itoa(content.ttl) + "\n")
					} else if content.ipType == 6 {
						ipv6 := getIpv6(content.ip[0:])
						fmt.Printf("\t\t" + content.domain + "  " + ipv6 + " " + strconv.Itoa(content.ttl) + "\n")
					}
					//fmt.Printf("   查询DNS解析记录成功：%s,%s\n", content.domain, content.ip)
				}
				//fmt.Println(time.Now(), addr, "查询DNS解析记录成功：", DomainTypeA)
				isSuccessControl = true
				_, err = listen.WriteToUDP([]byte{0x00}, addr) // 发送数据
				if err != nil {
					fmt.Println("err:", err)
					continue

				}
				continue
			}

			if isSuccessControl == true {
				continue
			} else {
				// 其他错误
				_, err = listen.WriteToUDP([]byte{0x01}, addr) // 发送数据
				if err != nil {
					fmt.Println("err:", err)
					continue

				}
				continue
			}
		}
		data[2] = 0x81
		data[3] = 0x83
		for index, b := range authByte {
			data[n+index] = b
		}
		_, err = listen.WriteToUDP(data[:n+len(authByte)], addr) // 发送数据
		if err != nil {
			fmt.Println("err:", err)
		}
		fmt.Println(time.Now().Format("2006/1/2 15:04:05"), addr, "未找到DNS记录", convertHex2String(Domain))
	}

}
