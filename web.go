package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func startHttp() {
	http.HandleFunc("/dns_discover", dnsDiscover)
	http.HandleFunc("/", index)
	http.HandleFunc("/dns", handler1203)
	http.HandleFunc("/dnslog", dnslog)
	httpPort := strconv.Itoa(config.HTTP_Port)
	addr := config.LinstenIP + ":" + httpPort
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		//panic(err)
		log.Fatal(err)
	}
}
func dnsDiscover(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("接收到客户端发的请求")
	//ip := r.PostFormValue("ip")
	ipType := r.PostFormValue("type")
	dns := r.PostFormValue("dns")
	//log.Println(ip, ipType)

	domain := ""
	for {
		randomInt := genRandomInt(9999)

		domain = randomInt + "." + dns
		//fmt.Println("得到随机数：", randomInt, domain)
		if checkDomainExit(DomainTypeA, domain) {
			continue
		} else {
			break
		}
	}
	if ipType == "4" {
		_, err := fmt.Fprintln(w, domain)
		if err != nil {
			log.Fatal(err)
		}

	} else if ipType == "6" {

	} else {
		domain = ""
		_, err := fmt.Fprintln(w, domain)
		if err != nil {
			log.Fatal(err)
		}

	}
}

func handler1203(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintln(w, DomainTypeA)
	if err != nil {
		return
	}
}
func index(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintln(w, "ddns server is running")
	if err != nil {
		return
	}
}
