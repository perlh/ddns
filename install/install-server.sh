#! /usr/bin/bash
echo "install ddns service"
echo "mkdir /opt/ddns/"
cp -r install /opt/ddns
cp "/opt/ddns/services/ddns-client-"`uanme -m`".service" /etc/systemd/system/ddns-client.service
systemctl systemctl daemon-reload
systemctl start ddns-client.service