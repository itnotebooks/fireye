package tools

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

func HttpGet(url string) string {

	// 发起http get请求
	resp, errClient := http.Get(url)
	if errClient != nil {
		fmt.Println("调用外网失败，请检查网络")
		return ""
	}
	// 程序在使用完 response 后必须关闭 response 的主体。
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return string(body)

}

// 通过向DNS发起请求，判断出口网卡，获取对应的网卡地址
func GetOutBoundIP() string {
	conn, err := net.Dial("udp", "223.5.5.5:53")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := strings.Split(localAddr.String(), ":")[0]
	return ip
}
