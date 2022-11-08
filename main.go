package main

import (
	"flag"
	"fmt"
)

//startHttp

type Config struct {
	LinstenIP     string
	HTTP_Port     int
	DNS_Port      int
	DnsFile       string
	DnsUpdateTime int
}

type ClientConfig struct {
	Interface       string
	DnsParentDomain string
	Domain          string
	ServerPort      int
	Server          string
	DnsIpType       int
	HttpPort        int
}

var (
	debug bool
	//listenIp       = flag.String("ip", "", "Dns server Listen Ip")
	//listenPort     = flag.Int("port", -1, "DNS server Port")

	//dnsFilePath    = flag.String("d", "", "DNS local File recode ")
	//dnsServerPort  = flag.Int("dns--server-port", 53, "DNS server port")

	config       Config
	clientConfig ClientConfig
	//interface=eth0
	//server=hsm.cool
	//dns-server=d.hsm.cool
	//local-domain=3.d.hsm.cool
	//ip-type=ipv4
	//server = flag.String("i", "en0", "network Card")
	//networkCard  = flag.String("interface", "en0", "network Card")
	//dnsServer    = flag.String("server", "hsm.cool", "server")
	//dnsTopDoamin = flag.String("dns-server", "d.hsm.cool", "DNS server")
	//dnsIpType    = flag.Int("dns-type", 4, "4 or 6. type about ipv4 and ipv6")
	//localDoamin  = flag.String("local-domain", "1.d.hsm.cool", "local doamin")
	configFilePath = flag.String("c", "", "Config file path")
	mode           = flag.String("m", "", "mode for server or client")
)

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: ddns [-m mode] [-c config-file-path]")
		fmt.Println("  -c string\n        Config file  path\n  -m string\n        mode for server or client\n")
	}
	//flag.String("example", "", "./ddns -m server -c ddns.conf")
	flag.Parse()
	if *mode == "client" {
		client()
	} else if *mode == "server" {
		server()
	} else {
		flag.PrintDefaults()
	}

}
