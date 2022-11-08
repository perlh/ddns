# 2022/11/7 这是一次全新的更新
## 更新内容
1. 将客户端和服务器端整合到一个项目中来
2. 支持服务器分配子域名给客户端，避免了域名冲突（当然客户端依然可以在本地指派指定域名）
3. 丰富了使用文档
# ddns
- DDNS（Dynamic Domain Name Server，动态域名服务）是将用户的动态IP地址映射到一个固定的域名解析服务上，用户每次连接网络的时候客户端程序就会通过信息传递把该主机的动态IP地址传送给位于服务商主机上的服务器程序

## 实现的功能

- 自动更新域名解析到本机IP,每隔一个时间间隔（配置文件可调）向dns服务器同步IP信息
- 目前支持`A记录`（ipv4）和`AAAA记录`（ipv6）
- 每条记录将设置计时器，当客户端与dns
服务器长期未同步时，dns服务器将删除客户端的dns记录。
- 支持查看dns记录表
    > http://dns-ip:http-port/dns
- 支持dnslog功能
    > http://dns-ip:http-port/dnslog
- 支持dns记录解析





# 编译ddns
```bash
cd ddns
go build
```

# server
1. 购买域名
2. 在云解析DNS上添加NS记录
3. 编辑配置文件`ddns.conf`
4. 在服务区上运行`ddns`
  ``` bash
./ddns -m server -c ddns.conf
  ```

### 为ddns创建service

`sudo vim /usr/systemd/system/ddns.service`

根据自己ddns的路径修改`ExecStart= /opt/ddns/bin/ddns -m server -c /opt/ddns/etc/ddns.conf
`
```text
[Unit]
Description=DDNS service
After=network.target

[Service]
Type=simple
User=nobody
Restart=on-failure
RestartSec=5s
ExecStart= /opt/ddns/bin/ddns -m server -c /opt/ddns/etc/ddns.conf
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
```
**注意修改**


**重载配置并且启动ddns服务**
```bash
systemctl daemon-reload
systemctl start ddns.service
systemctl status ddns.service
```

# client
编辑client配置文件
```bash
vim ddns-client.conf
```
运行ddns client
```bash
./ddns -m client -c ddns-client.conf
```
