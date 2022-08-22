package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

//var serverNATAddr, serverNATPort = "218.89.171.148", "25118"
type STATUS = int

const (
	SUCCESS STATUS = 0
	SERVER  STATUS = 1
	LOCALE  STATUS = 2
)

var wg sync.WaitGroup
var serverIPv6Addr, serverIPv6Port = []string{"2409:8a60:1e74:15f1:6183:a810:5d4a:ea7d",
	"2409:8a60:1e74:15f1:82cb:c469:ed53:2"}, "20000"

func accessPacket(data []byte) {
	
}
func send(str string, conn net.Conn) bool {
	if strings.ToUpper(str) == "Q" {
		return false
	}
	_, err := conn.Write([]byte(str))
	if err != nil {
		fmt.Println("send failed ,message: ", str[:], "...  ,err: ", err)
		return false
	}
	return true
}
func recv(conn net.Conn) bool {
	buf := [512]byte{}
	n, err := conn.Read(buf[:])
	if err != nil {
		fmt.Println("recv from server failed ,err: ", err)
		return false
	}
	fmt.Println("recv from server:", string(buf[:n]))
	return true
}
func SendFile(filePath string, conn net.Conn) STATUS {
	var str string
	fmt.Print("请输入文件的完整路径：")
	fmt.Scanln(&str)
	//获取文件信息
	fileInfo, err := os.Stat(str)
	if err != nil {
		fmt.Println("路径有误或无法打开文件")
		return LOCALE
	}
	fileName := fileInfo.Name()
	fileSize := fileInfo.Size()
	//fileHash := Sha256File(fileName)
	//发送文件名称到服务端
	fmt.Println("正在连接服务器...")
	_, err = conn.Write([]byte(fileName))
	if err != nil {
		return SERVER
	}
	//读取服务端内容
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		return SERVER
	}
	revData := string(buf[:n])
	if revData != "ok" {
		return SERVER
	}
	//读取发送文件
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return LOCALE
	}
	defer f.Close()
	var count int64
	for {
		buf := make([]byte, 1024)
		//读取文件内容
		n, err := f.Read(buf)

		if err != nil && io.EOF == err {
			fmt.Println("文件传输完成")
			//告诉服务端结束文件接收-
			conn.Write([]byte("finish"))
			return 0
		}
		//发送给服务端
		conn.Write(buf[:n])

		count += int64(n)
		sendPercent := float64(count) / float64(fileSize) * 100
		value := fmt.Sprintf("%.2f", sendPercent)
		//打印上传进度
		fmt.Println("文件上传：" + value + "%")
	}
}
func main() {
	for {
		var choice int
		for {
			fmt.Println("choose 0 or 1:")
			n, err := fmt.Scanln(&choice)
			if err == nil && n == 1 {
				break
			}
		}
		if 0 <= choice && choice < 2 {
			var DialServer string = "[" + serverIPv6Addr[choice] + "]:" + serverIPv6Port
			wg.Add(1)
			go chat(DialServer)
			wg.Wait()
		}
	}
}
func chat(DialServer string) {
	defer wg.Done()
	for {

		conn, err := net.Dial("tcp", DialServer)
		fmt.Println("dialing at ", DialServer)
		if err != nil {
			fmt.Println("dial failed, err:", err)
			break
		}
		defer conn.Close()
		inputReader := bufio.NewReader(os.Stdin)
		for {
			var inputInfo string
			fmt.Println("connected! you can send now:")
			for {
				input, err := inputReader.ReadString('\n')

				if err != nil {
					fmt.Println("read err", err)
				}
				inputInfo = strings.Trim(input, "\r\n")
				if inputInfo != "" {
					break
				}
			}
			if strings.ToUpper(inputInfo) == "Q" {
				break
			}
			if send(inputInfo, conn) {
				//t1 = time.Now().UnixMilli() / 1000
			} else {
				break
			}
			if !recv(conn) {
				break
			}

		}

		fmt.Println("connection was closed")
		fmt.Println("connect again ?[Y/any key]")
		var answer string
		fmt.Scanln(&answer)
		if strings.ToUpper(answer) != "Y" {
			fmt.Println("retry conn... ")
			return
		}
	}
}
