package netutil

import (
	"net"
	"regexp"
	"strconv"
	"strings"
)

// 检测 地址是否为 IP:端口 格式
func HostAddrCheck(addr string) bool {

	if "" == addr {
		return false
	}

	items := strings.Split(addr, ":")
	if items == nil || len(items) != 2 {
		return false
	}

	a := net.ParseIP(items[0])
	if a == nil {
		return false
	}

	match, err := regexp.MatchString("^[0-9]*$", items[1])
	if err != nil {
		return false
	}

	i, err := strconv.Atoi(items[1])
	if err != nil {
		return false
	}
	if i < 0 || i > 65535 {
		return false
	}

	if match == false {
		return false
	}
	return true
}

// 获取一个空闲的TCP端口
func GetFreePort() int {
	var port int
	for i := 17070; i < 65536; i++ {
		addr, _ := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(i))
		listener, err := net.ListenTCP("tcp", addr)
		if err == nil {
			listener.Close()
			port = i
			break
		}
	}
	return port
}
