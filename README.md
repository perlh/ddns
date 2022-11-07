# ddns-server
- dns服务器
- ddns服务器端
- DDNS（Dynamic Domain Name Server，动态域名服务）是将用户的动态IP地址映射到一个固定的域名解析服务上，用户每次连接网络的时候客户端程序就会通过信息传递把该主机的动态IP地址传送给位于服务商主机上的服务器程序
  
## 实现的功能

- 自动更新域名解析到本机IP,每隔5min向dns服务器同步IP信息
- 支持A记录（IPV4）和AAAA记录（IPV6）
- 每条记录将设置计时器，当客户端与DNS服务器长期（10min）未同步时，将删除客户端的DNS记录。
- 支持查看dns记录表
- 支持dnslog功能
- 支持dns记录解析

## 使用
1. 首先得拥有自己的域名
2. 首先在云解析DNS上添加NS记录（比如d.hsm.cool 服务器指向www.hsm.cool）
3. 在服务区上运行`ddns服务器程序` 
``` bash
hsm@orangepizero2 ~>  ./ddns -d dns.txt -ip 0.0.0.0 -port 8050 
hsm@orangepizero2 ~> # or
hsm@orangepizero2 ~>  ./ddns -c ddns.conf
```

配置文件 `ddns.conf`
```text
dns_file=dns.txt
listen_ip=0.0.0.0
listen_port=8050
```
>ddns.conf不能有注释和空格


## 为ddns创建service

`sudo vim /usr/systemd/system/ddns.service`
```text
[Unit]
Description=DDNS service
After=network.target

[Service]
Type=simple
User=nobody
Restart=on-failure
RestartSec=5s
ExecStart= /opt/ddns/bin/ddns-server-linux-amd -c /opt/ddns/etc/ddns.conf
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
```
**注意修改**
> ExecStart= /opt/ddns/bin/ddns-server-linux-amd -c /opt/ddns/etc/ddns.conf


**重载配置并且启动ddns服务**
```bash
root@VM-12-8-ubuntu:/opt/ddns/bin# systemctl daemon-reload
root@VM-12-8-ubuntu:/opt/ddns/bin# systemctl start ddns.service
root@VM-12-8-ubuntu:/opt/ddns/bin# systemctl status ddns.service
```

## 例子
``` bash
hsm@VM-12-8-ubuntu /o/d/bin> ./ddns -c ddns.conf
2022/11/07 19:56:47 ddns server:0.0.0.0:8050
2022/11/07 19:56:47 server local dns file path: dns.txt
```
