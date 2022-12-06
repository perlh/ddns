package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// 互斥变量
var mu sync.RWMutex
var databases Databases

type User struct {
	UserId   int    `json:"user_id"`
	UserName string `json:"user_name"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type Users struct {
	data []User
}

func (u *Users) add(user User) (err error) {
	// 检查是否存在这个用户
	_, err1 := u.searchByUsername(user.UserName)
	if err1 == nil {
		// 如果查到了这个用户，那么就不能添加这个用户
		return errors.New("用户已存在")
	}

	//查找最后的那个id
	user.UserId = 0
	if len(u.data) != 0 {
		user.UserId = u.data[len(u.data)-1].UserId + 1
	}
	u.data = append(u.data, user)
	return err
}

func (u *Users) del(user User) {
	var tmp []User
	if user.UserId == 0 {
		// 不能删除本地用户
		return
	}
	//var tmpTable Table
	for _, value := range u.data {
		if value.UserId != user.UserId {
			tmp = append(tmp, value)
		}
	}
	u.data = tmp
}

func (u *Users) update(user User) {
	var tmp []User
	//var tmpTable Table
	for _, value := range u.data {
		if value.UserId == user.UserId {

			tmp = append(tmp, user)
			continue
		}
		tmp = append(tmp, value)
	}
	u.data = tmp
}
func (u *Users) searchByToken(Token string) (user User, err error) {
	for _, value := range u.data {
		if value.Token == Token {
			return value, nil
			//continue
		}
	}
	return user, errors.New("no result！")
}

// searchByUsername 如果得到结果就返回nil，否则err等于errors.New("no result！")
func (u *Users) searchByUsername(username string) (user User, err error) {
	for _, value := range u.data {
		if value.UserName == username {
			return value, nil
			//continue
		}
	}
	return user, errors.New("no result！")
}

var users Users

type Table struct {
	dnsType         string
	ipv4            net.IP
	ipv6            net.IP
	cname           string
	domain          string
	ttl             int
	isLocalDNSTable bool
	ownID           int
}

type Databases struct {
	data []Table
}

func (databases *Databases) del(table Table) (err error) {
	mu.Lock()
	defer mu.Unlock()
	// 定义一个临时变量来保存数据
	var tmp []Table
	// 将数据保存保存在tmp中
	for i := 0; i < len(databases.data); i++ {
		if databases.data[i].domain == table.domain {
			continue
		}
		tmp = append(tmp, databases.data[i])
	}
	databases.data = tmp
	return nil
}

func (databases *Databases) add(table Table) (err error) {
	mu.Lock()
	defer mu.Unlock()
	// 检查域名是否已经存在
	for _, table2 := range databases.data {
		if table2.domain == table.domain {
			return errors.New("这个域名已经存在，无法再添加")
		}
	}
	databases.data = append(databases.data, table)
	return nil
}

// update 通过domain更新整个表
func (databases *Databases) update(table Table) (err error) {

	// 查一下这个域名是否存在
	resultTable, err1 := databases.searchDNS(table.domain)
	if err1 != nil {
		// 如果不存在，那么就添加
		err = databases.add(table)
		if err != nil {
			return errors.New("更新失败")
		}
		return nil
	} else {
		if resultTable.ownID != table.ownID {
			return errors.New("越权更新")
		}
		// 查看这个domain是否属于这个用户
		mu.Lock()
		defer mu.Unlock()
		//fmt.Println("rss:", data)

		// 将数据保存保存在tmp中
		for i := 0; i < len(databases.data); i++ {
			if databases.data[i].domain == table.domain {
				databases.data[i] = table
			}
		}
		return nil
	}

}

// ASearch 查找A记录，如果得到结果，那么err为空
func (databases *Databases) searchDNS(domain string) (table Table, err error) {
	mu.Lock()
	defer mu.Unlock()
	// 遍历一下数据表
	for _, table := range databases.data {
		if table.domain == domain {
			return table, nil
		}
	}
	return table, errors.New("no result")
}

// update 通过domain更新整个表
func (databases *Databases) updateTTL(table Table) (err error) {
	mu.Lock()
	defer mu.Unlock()
	//fmt.Println("rss:", data)

	// 将数据保存保存在tmp中
	for i := 0; i < len(databases.data); i++ {
		if databases.data[i].domain == table.domain {
			// 防止溢出
			if databases.data[i].ttl > 5 {
				continue
			}
			databases.data[i].ttl = databases.data[i].ttl - 1
		}
	}
	return nil
}

// ringUpdateDNS 每隔server.screenTime秒使ttl减1，当ttl减到负数的时候，就将这条记录删掉
func ringUpdateDNS() {
	d := time.NewTicker(time.Duration(server.screenTime) * time.Second)
	for {
		var tmp Databases
		tmp = databases

		//fmt.Println(recodeA)
		if server.debug {
			log.Println("ringUpdateDNS")
		}
		for _, table := range tmp.data {
			if server.debug {
				log.Println("ringUpdateDNS", table.domain, table.dnsType, table.ipv4, table.ipv6, table.cname, table.ttl, table.ownID)
			}
			// 本地dns记录不检查
			if table.isLocalDNSTable == true {
				continue
			}

			if table.ttl <= 0 {
				// 如果这条域名长期不更新，那么删除这条记录
				//fmt.Println("delete", data.domain)
				err := databases.del(table)
				if err != nil {
					log.Println(err)
					continue
				}
			} else {
				err := databases.updateTTL(table)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}
		if server.debug {
			log.Println("ringUpdateDNS")
		}
		<-d.C
	}

}

func initDnsServer() {
	// 加载本地dns文件
	if server.debug == true {
		log.Println("initDnsServer")
	}
	if server.dnsFile == "" {
		log.Println("Waring: 未找到本地DNS文件表", server.dnsFile)
		return
	}
	err := loadLocalDnsFile(server.dnsFile)
	if err != nil {
		log.Println("加载dns本地文件失败！")
	}

}

func loadLocalDnsFile(fileName string) (err error) {
	f, err := os.Open(fileName)
	if err != nil {
		//
		//log.Println("未找到dns记录文件!")
		//CreateDnsFile(fileName)
		//log.Println("自动创建dns记录文件：/%v\n", fileName)
		return errors.New("找不到这个路径下的DNS文件")
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
		// 过滤注释
		if s.Text() == "" || s.Text()[0] == '#' {
			continue
		}
		var tmpTable Table
		tmpTable.isLocalDNSTable = true
		recode := strings.Split(s.Text(), " ")
		if len(recode) != 3 {
			log.Println("dns配置文件错误！")
		}
		if server.debug {
			log.Println("loadLocalDnsFile", recode)
		}
		//fmt.Println(recode, recode[0], recode[1], recode[2])
		switch recode[0] {
		case "A":
			{

				tmpTable.domain = recode[1]
				tmpTable.ownID = 0

				//var tmpByte16 [16]byte
				ipv4 := net.ParseIP(recode[2])
				tmpTable.ipv4 = ipv4.To4()
				//tmpTable.
				tmpTable.dnsType = "A"
				//fmt.Println(tmpTable)
				//fmt.Println(DomainTypeA)
				err := databases.add(tmpTable)
				//logs.Println("err")
				//fmt.Println(err, "err")
				if err != nil {
					panic(err)
					return err
				}
				break
			}
		case "AAAA":
			{
				tmpTable.domain = recode[1]
				ipv4 := net.ParseIP(recode[2])
				tmpTable.ipv6 = ipv4.To16()
				tmpTable.dnsType = "AAAA"
				err := databases.add(tmpTable)
				if err != nil {
					return err
				}
				break
			}
		case "CNAME":
			{
				tmpTable.domain = recode[1]
				tmpTable.cname = recode[2]
				tmpTable.dnsType = "CNAME"
				err := databases.add(tmpTable)
				if err != nil {
					return err
				}
				break
			}
		default:
			break

		}
		if server.debug {
			log.Println(recode)
		}

	}
	//fmt.Println(DomainTypeA)
	err = s.Err()
	if err != nil {
		panic(err)
		return err
	}
	return nil
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

func serverDNS() {
	//
	//records = map[string]string{
	//	"google.com": "216.58.196.142",
	//	"amazon.com": "176.32.103.205",
	//}

	//Listen on UDP Port
	addr := net.UDPAddr{
		Port: server.dnsPort,
		IP:   net.ParseIP(server.listenIP),
	}
	u, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
		return
	}

	log.Println("listen udp ", addr.String())
	// Wait to get request on that port
	for {
		tmp := make([]byte, 1024)
		_, addr, _ := u.ReadFrom(tmp)
		clientAddr := addr
		//logs.Println(addr)
		packet := gopacket.NewPacket(tmp, layers.LayerTypeDNS, gopacket.Default)
		dnsPacket := packet.Layer(layers.LayerTypeDNS)
		tcp, _ := dnsPacket.(*layers.DNS)
		serveDNS(u, clientAddr, tcp)
	}
}

func serveDNS(u *net.UDPConn, clientAddr net.Addr, request *layers.DNS) {
	replyMess := request
	var dnsAnswer layers.DNSResourceRecord
	dnsAnswer.Type = layers.DNSTypeA
	//var ip string
	var err error
	log.Println(clientAddr, "query:", string(request.Questions[0].Name))
	//var table Table

	table, err := databases.searchDNS(string(request.Questions[0].Name))
	if err != nil {

		//Todo: Log no data present for the IP and handle:todo
	}

	if table.dnsType == "A" {
		dnsAnswer.Type = layers.DNSTypeA
		dnsAnswer.IP = table.ipv4
		dnsAnswer.Name = []byte(request.Questions[0].Name)
		//fmt.Println(request.Questions[0].Name)
		dnsAnswer.Class = layers.DNSClassIN
		replyMess.QR = true
		replyMess.ANCount = 1
		replyMess.OpCode = layers.DNSOpCodeNotify
		replyMess.AA = true
		replyMess.Answers = append(replyMess.Answers, dnsAnswer)
		replyMess.ResponseCode = layers.DNSResponseCodeNoErr
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{} // See SerializeOptions for more details.
		err = replyMess.SerializeTo(buf, opts)
		if err != nil {
			panic(err)
		}
		u.WriteTo(buf.Bytes(), clientAddr)
	} else if table.dnsType == "AAAA" {
		//a, _, _ := net.ParseCIDR(ip + "/24")
		dnsAnswer.Type = layers.DNSTypeAAAA
		dnsAnswer.IP = table.ipv6
		dnsAnswer.Name = []byte(request.Questions[0].Name)
		//fmt.Println(request.Questions[0].Name)
		dnsAnswer.Class = layers.DNSClassIN
		replyMess.QR = true
		replyMess.ANCount = 1
		replyMess.OpCode = layers.DNSOpCodeNotify
		replyMess.AA = true
		replyMess.Answers = append(replyMess.Answers, dnsAnswer)
		replyMess.ResponseCode = layers.DNSResponseCodeNoErr
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{} // See SerializeOptions for more details.
		err = replyMess.SerializeTo(buf, opts)
		if err != nil {
			panic(err)
		}
		u.WriteTo(buf.Bytes(), clientAddr)
	} else if table.dnsType == "CNAME" {
		//a, _, _ := net.ParseCIDR(ip + "/24")
		dnsAnswer.Type = layers.DNSTypeCNAME
		dnsAnswer.CNAME = []byte(table.cname)
		dnsAnswer.Name = []byte(request.Questions[0].Name)
		//fmt.Println(request.Questions[0].Name)
		dnsAnswer.Class = layers.DNSClassIN
		replyMess.QR = true
		replyMess.ANCount = 1
		replyMess.OpCode = layers.DNSOpCodeNotify
		replyMess.AA = true
		replyMess.Answers = append(replyMess.Answers, dnsAnswer)
		replyMess.ResponseCode = layers.DNSResponseCodeNoErr
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{} // See SerializeOptions for more details.
		err = replyMess.SerializeTo(buf, opts)
		if err != nil {
			panic(err)
		}
		u.WriteTo(buf.Bytes(), clientAddr)
	} else {
		//a, _, _ := net.ParseCIDR(ip + "/24")
		dnsAnswer.Type = layers.DNSTypeA
		dnsAnswer.IP = table.ipv4
		dnsAnswer.Name = []byte(request.Questions[0].Name)
		fmt.Println(request.Questions[0].Name)
		dnsAnswer.Class = layers.DNSClassIN
		replyMess.QR = true
		replyMess.ANCount = 1
		replyMess.OpCode = layers.DNSOpCodeNotify
		replyMess.AA = true
		replyMess.Answers = append(replyMess.Answers, dnsAnswer)
		replyMess.ResponseCode = layers.DNSResponseCodeNXDomain
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{} // See SerializeOptions for more details.
		err = replyMess.SerializeTo(buf, opts)
		if err != nil {
			panic(err)
		}
		u.WriteTo(buf.Bytes(), clientAddr)
	}

}

func initServer() {
	// 创建本地用户
	err := users.add(User{UserName: server.user, Password: server.password})
	//log.Println(users, err)
	if err != nil {
		log.Println("create account error!")
		return
	}
	//log.Println("user:", users)
	log.Println("create account:", server.user, server.password)
	logger.Println("create account:", server.user, server.password)
}

func serverStart() {
	initServer()
	initDnsServer()
	// 添加用户
	log.Println("init dns server success ")
	go ringUpdateDNS()
	go serverDNS()
	go controlServer()
	select {}
}
