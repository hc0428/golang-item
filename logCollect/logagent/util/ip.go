package util

import (
	"fmt"
	"net"
	"strings"
)

//获取对外ip

func GetOutboundIp() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.IP.String(), ":")[0]
	fmt.Println("ip : ", ip)
	return
}
