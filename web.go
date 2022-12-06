package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// JsonResult json返回体
type JsonResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func parseDnsType(dnsType string) (dnsTypeString string, err error) {
	if dnsType == "A" || dnsType == "a" {
		dnsTypeString = "A"
		return dnsTypeString, nil
	}
	if dnsType == "AAAA" || dnsType == "aaaa" {
		dnsTypeString = "AAAA"
		return dnsTypeString, nil
	}
	if dnsType == "CNAME" || dnsType == "cname" {
		dnsTypeString = "CNAME"
		return dnsTypeString, nil
	}
	return dnsTypeString, errors.New("not found")
}

func controlServer() {
	http.HandleFunc("/register", httpRegister)
	http.HandleFunc("/show", httpShow)
	http.HandleFunc("/update", httpUpdate)
	http.HandleFunc("/create_domain", httpCreateDomain)
	http.HandleFunc("/test", test)

	addr := server.listenIP + ":" + strconv.Itoa(server.listenPort)
	log.Println("listen udp ", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func getTime() string {
	currentTime := time.Now()
	result := currentTime.Unix()
	return strconv.Itoa(int(result >> 1))
}

func checkToken(token string) (User, bool) {
	data := strings.Split(Base58Decoding(token), ":")
	//log.Println("checkToken", Base58Decoding(token))
	//if server.debug {
	//
	//}
	//log.Println(len(data))
	if len(data) == 2 {
		user, err := users.searchByUsername(data[0])
		//log.Println("user:", user, err)
		if err == nil {

			message := user.UserName + user.Password + getTime()
			tokenMd5 := genMd5(message)
			if server.debug {
				log.Println("checkToken", getTime(), tokenMd5, message)
			}
			if tokenMd5 == data[1] {
				return user, true
			}
		}

	}
	return User{}, false
}

func encodeToken(user string, password string) (token string) {

	message := user + password + getTime()
	tokenMd5 := genMd5(message)
	parseString := user + ":" + tokenMd5
	if server.debug {
		log.Println("encodeToken", getTime(), tokenMd5, message)
	}
	token = Base58Encoding(parseString)
	//if server.debug {
	//	log.Println("token", token)
	//}
	return token
}

func Base64Encoding(str string) string { //Base64编码
	src := []byte(str)
	res := base64.StdEncoding.EncodeToString(src) //将编码变成字符串
	return res
}

func Base64Decoding(str string) string { //Base64解码
	res, _ := base64.StdEncoding.DecodeString(str)
	return string(res)
}

/*
	source :https://www.jb51.net/article/218116.htm
*/

var base58 = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func Base58Encoding(str string) string { //Base58编码
	//1. 转换成ascii码对应的值
	strByte := []byte(str)
	//fmt.Println(strByte) // 结果[70 97 110]
	//2. 转换十进制
	strTen := big.NewInt(0).SetBytes(strByte)
	//fmt.Println(strTen)  // 结果4612462
	//3. 取出余数
	var modSlice []byte
	for strTen.Cmp(big.NewInt(0)) > 0 {
		mod := big.NewInt(0) //余数
		strTen58 := big.NewInt(58)
		strTen.DivMod(strTen, strTen58, mod)             //取余运算
		modSlice = append(modSlice, base58[mod.Int64()]) //存储余数,并将对应值放入其中
	}
	//  处理0就是1的情况 0使用字节'1'代替
	for _, elem := range strByte {
		if elem != 0 {
			break
		} else if elem == 0 {
			modSlice = append(modSlice, byte('1'))
		}
	}
	//fmt.Println(modSlice)   //结果 [12 7 37 23] 但是要进行反转，因为求余的时候是相反的。
	//fmt.Println(string(modSlice))  //结果D8eQ
	ReverseModSlice := ReverseByteArr(modSlice)
	//fmt.Println(ReverseModSlice)  //反转[81 101 56 68]
	//fmt.Println(string(ReverseModSlice))  //结果Qe8D
	return string(ReverseModSlice)
}

func ReverseByteArr(bytes []byte) []byte { //将字节的数组反转
	for i := 0; i < len(bytes)/2; i++ {
		bytes[i], bytes[len(bytes)-1-i] = bytes[len(bytes)-1-i], bytes[i] //前后交换
	}
	return bytes
}

//就是编码的逆过程
func Base58Decoding(str string) string { //Base58解码
	strByte := []byte(str)
	//fmt.Println(strByte)  //[81 101 56 68]
	ret := big.NewInt(0)
	for _, byteElem := range strByte {
		index := bytes.IndexByte(base58, byteElem) //获取base58对应数组的下标
		ret.Mul(ret, big.NewInt(58))               //相乘回去
		ret.Add(ret, big.NewInt(int64(index)))     //相加
	}
	//fmt.Println(ret)  // 拿到了十进制 4612462
	//fmt.Println(ret.Bytes())  //[70 97 110]
	//fmt.Println(string(ret.Bytes()))
	return string(ret.Bytes())
}

func test(w http.ResponseWriter, r *http.Request) {
	var response JsonResult
	response.Code = 200
	currentTime := time.Now()
	result := currentTime.Unix()
	users.add(User{UserName: "hsm", Password: "905008"})
	//log.Println(users)
	enc := encodeToken("hsm", "905008")
	fmt.Println(enc)
	//time.Sleep(time.Second * 1)
	// 用户认证

	//m, _ := time.ParseDuration("-10m")
	//result := currentTime.Add(m).Unix()
	response.Msg = strconv.Itoa(int(result))
	data, _ := json.Marshal(response)
	_, _ = fmt.Fprintln(w, string(data))
}

func httpRegister(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		JsonResult
		User
	}
	var response Response

	rootToken := r.PostFormValue("root_token")
	_, ok := checkToken(rootToken)

	if ok {
		//return
		newUsername := r.PostFormValue("user")
		newPassword := r.PostFormValue("password")
		if newUsername != "" || newPassword != "" {
			if strings.Contains(newUsername, ":") || strings.Contains(newUsername, ":") {
				// 不能包含这个字符串
				response.Code = 500
				response.Msg = "参数不能包含':'字符"
				data, _ := json.Marshal(response)
				_, _ = fmt.Fprintln(w, string(data))
				return
			}
			token := genMd5(newUsername + newPassword)
			user2 := User{UserName: newUsername, Password: newPassword}
			err := users.add(user2)
			if err != nil {
				oldUser, err := users.searchByUsername(newUsername)
				if err == nil {
					if oldUser.UserName == newUsername && oldUser.Password == newPassword {
						// 针对重复添加用户
						response.Code = 200
						response.Msg = "用户注册成功"
						response.Token = token
						response.UserName = newUsername
						response.Password = newPassword
						data, _ := json.Marshal(response)
						_, _ = fmt.Fprintln(w, string(data))
						return
					}

				}
				// 添加失败
				response.Code = 500
				response.Msg = "该用户被注册，请使用别的账户"
				data, _ := json.Marshal(response)
				_, _ = fmt.Fprintln(w, string(data))
				return
			}
			response.Token = token
			response.UserName = newUsername
			response.Password = newPassword
			logger.Println("用户注册成功", response.UserName, response.Password, response.UserId)
			response.Code = 200
			response.Msg = "用户注册成功"
			data, _ := json.Marshal(response)
			_, _ = fmt.Fprintln(w, string(data))
			return
		}
		response.Code = 500
		response.Msg = "参数错误！"
		data, _ := json.Marshal(response)
		_, _ = fmt.Fprintln(w, string(data))
		return
	}
	response.Code = 500
	response.Msg = "认证失败"
	data, _ := json.Marshal(response)
	_, _ = fmt.Fprintln(w, string(data))
	return

}

// httpCreateDomain 创建一个域名
func httpCreateDomain(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		JsonResult
		Domain string `json:"domain"`
	}
	var response Response
	dnsType := r.PostFormValue("dns_type")
	value := r.PostFormValue("value")
	time := r.PostFormValue("time")
	token := r.PostFormValue("token")
	userData, ok := checkToken(token)
	if !ok {
		response.Code = 500
		response.Msg = "认证失败！"
		data, _ := json.Marshal(response)
		_, _ = fmt.Fprintln(w, string(data))
		return
	}
	if dnsType != "" || value != "" || time != "" {
		timeSecond, err := strconv.Atoi(time)
		if err != nil {
			response.Code = 500
			response.Msg = "time设置错误"
			data, _ := json.Marshal(response)
			_, _ = fmt.Fprintln(w, string(data))
			return
		}
		var table Table
		table.ttl = timeSecond / 20
		table.domain = genRandDomain(10)
		// 检查是否重复
		forTtl := 20
		for {

			_, err := databases.searchDNS(table.domain)
			if err != nil {

				break
			}
			if forTtl < 0 {
				logger.Println("域名池已满！")
				break
			}
			forTtl--

		}
		dnsTypeString, err := parseDnsType(dnsType)
		if err != nil {
			response.Code = 500
			response.Msg = "dnsType设置错误"
			data, _ := json.Marshal(response)
			_, _ = fmt.Fprintln(w, string(data))
			return
		}
		table.dnsType = dnsTypeString
		response.Domain = table.domain
		table.ownID = userData.UserId
		// 添加ip
		if dnsTypeString == "A" {
			table.ipv4 = net.ParseIP(value)
		} else if dnsTypeString == "AAAA" {
			table.ipv6 = net.ParseIP(value).To16()
		} else {
			table.cname = value
		}
		err = databases.add(table)
		if err != nil {
			response.Code = 500
			response.Msg = err.Error()
			data, _ := json.Marshal(response)
			_, _ = fmt.Fprintln(w, string(data))
			return
		}
		response.Code = 200
		data, _ := json.Marshal(response)
		_, _ = fmt.Fprintln(w, string(data))
		return
	}
	response.Code = 500
	data, _ := json.Marshal(response)
	_, _ = fmt.Fprintln(w, string(data))
	return

}

func httpShow(w http.ResponseWriter, r *http.Request) {
	// JsonResult json返回体
	type ResponseTable struct {
		Domain  string `json:"domain"`
		Value   string `json:"value"`
		DnsType string `json:"dns_type"`
		Ttl     int    `json:"ttl"`
	}

	type Response struct {
		JsonResult
		Data []ResponseTable `json:"data"`
	}
	var response Response

	var responseTableTmpArray []ResponseTable
	responseTableTmpArray = make([]ResponseTable, 0)
	for _, table := range databases.data {
		var responseTableTmp ResponseTable
		responseTableTmp.Domain = table.domain
		responseTableTmp.DnsType = table.dnsType
		responseTableTmp.Ttl = table.ttl
		if table.dnsType == "A" {
			responseTableTmp.Value = table.ipv4.To4().String()
		} else if table.dnsType == "AAAA" {
			//logs.Println(table.ipv6.To16().String())
			responseTableTmp.Value = table.ipv6.To16().String()
		} else if table.dnsType == "CNAME" {

			responseTableTmp.Value = table.cname
		}
		responseTableTmpArray = append(responseTableTmpArray, responseTableTmp)
	}

	response.Data = responseTableTmpArray
	response.Msg = ""
	response.Code = 200
	data, err := json.Marshal(response)
	if err != nil {
		_, err := fmt.Fprintln(w, "404")
		if err != nil {
			log.Fatal(err)
		}
	}
	_, err = fmt.Fprintln(w, string(data))
	if err != nil {
		log.Fatal(err)
	}
}

func httpUpdate(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		JsonResult
		Domain  string `json:"domain"`
		DnsType string `json:"dns_type"`
		Value   string `json:"value"`
	}
	var response Response
	var table Table

	// 用户认证
	token := r.PostFormValue("token")
	userData, ok := checkToken(token)
	if !ok {
		response.Code = 500
		response.Msg = "认证失败！"
		data, _ := json.Marshal(response)
		_, _ = fmt.Fprintln(w, string(data))
		return
	}
	// 检查这个域名是否被别人注册过
	//context.TODO()
	domain := r.PostFormValue("domain")
	table2, err2 := databases.searchDNS(domain)
	// 查询之前已经存在过这条记录
	if err2 == nil {
		if table2.ownID != userData.UserId {
			response.Code = 500
			response.Msg = "域名已经被别的用户注册过，请更换域名！"
			data, _ := json.Marshal(response)
			_, _ = fmt.Fprintln(w, string(data))
			return
		}
	}

	table.ownID = userData.UserId

	dnsType := r.PostFormValue("dnsType")
	response.Domain = domain
	response.DnsType = dnsType
	if domain == "" {
		response.Code = 500
		response.Msg = "域名为空"
		data, _ := json.Marshal(response)
		_, _ = fmt.Fprintln(w, string(data))
		return
	}
	//logs.Println(response)

	if dnsType == "a" || dnsType == "A" {
		ipv4 := r.PostFormValue("value")
		if ipv4 == "" {
			response.Code = 500
			data, _ := json.Marshal(response)
			_, _ = fmt.Fprintln(w, string(data))
			return
		}
		parseIpv4 := net.ParseIP(ipv4)
		table.ipv4 = parseIpv4
		table.dnsType = "A"
		response.Value = ipv4
	} else if dnsType == "aaaa" || dnsType == "AAAA" {
		ipv6 := r.PostFormValue("value")
		if ipv6 == "" {
			response.Code = 500
			data, _ := json.Marshal(response)
			_, _ = fmt.Fprintln(w, string(data))
			return
		}
		response.Value = ipv6
		parseIpv6 := net.ParseIP(ipv6)
		table.ipv6 = parseIpv6
		table.dnsType = "AAAA"
	} else if dnsType == "cname" || dnsType == "CNAME" {
		cname := r.PostFormValue("value")
		if cname == "" {
			response.Code = 500
			data, _ := json.Marshal(response)
			_, _ = fmt.Fprintln(w, string(data))
			return
		}
		response.Value = cname
		table.cname = cname
		table.dnsType = "CNAME"
	} else {
		response.Code = 500
		response.Msg = "提交的参数不对劲！"
		data, _ := json.Marshal(response)
		_, _ = fmt.Fprintln(w, string(data))
		return
	}

	if response.Code == 500 {

		data, _ := json.Marshal(response)
		_, _ = fmt.Fprintln(w, string(data))
		return
	}

	//更新
	table.ttl = 3
	table.domain = domain
	err := databases.update(table)
	if err != nil {
		response.Code = 500
		data, _ := json.Marshal(response)
		_, _ = fmt.Fprintln(w, string(data))
		return
	}
	response.Code = 200
	//log.Println("成功接收心跳包：", response.Domain, response.Value, response.DnsType)
	//logger.Println("成功接收心跳包：", response.Domain, response.Value, response.DnsType)
	data, _ := json.Marshal(response)
	_, _ = fmt.Fprintln(w, string(data))
	//log.Println("成功接收心跳包：", response.Domain, response.Value, response.DnsType)
	//logger.Println("成功接收心跳包：", response.Domain, response.Value, response.DnsType)
	return
}

func genMd5(code string) string {
	//c1 := md5.Sum([]byte(code)) //返回[16]byte数组

	Md5 := md5.New()
	_, _ = io.WriteString(Md5, code)
	return hex.EncodeToString(Md5.Sum(nil))
}

// 得到随机字符串
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandStringBytesMaskImpr  https://colobu.com/2018/09/02/generate-random-string-in-Go/
func genRandDomain(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b) + "." + server.Domain
}
