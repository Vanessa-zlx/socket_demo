package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"learnGo/package/socket_demo/utils"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var serverNATAddr, serverNATPort = "218.89.171.148", "25118"

type STATUS = int

type MODE = int

const (
	SUCCESS STATUS = iota
	SERVER
	LOCALE
	UNKNOWN
)
const (
	COMMAND MODE = iota
	GLOBAL
	FILE
)

var ALIVE = UNKNOWN

var wg sync.WaitGroup
var mux sync.Mutex
var live sync.Mutex

//var serverIPv6Addr, serverIPv6Port = []string{"2409:8a60:1e74:15f1:6183:a810:5d4a:ea7d",
//	"2409:8a60:1e74:15f1:82cb:c469:ed53:2"}, "20000"

func accessMessagePacket(msg string, mode MODE) []byte {
	size := len([]byte(msg))
	raw := bytes.Buffer{}
	raw.Write([]byte{byte(mode), byte(size / 100), byte(size % 100 / 10), byte(size % 10)})
	raw.Write([]byte(msg)[:len(msg)])
	//fmt.Println(raw.Bytes()) //
	cipherData := utils.Encipher(raw.Bytes(), []byte(utils.Sha256String(msg)), 304)
	buffer := bytes.Buffer{}
	buffer.Write(utils.Sha256String(msg)) //32
	buffer.Write(cipherData)              //304
	return buffer.Bytes()
}
func send(str string, conn net.Conn, status *STATUS, mode MODE) {
	defer wg.Done()

	/*
		第一次检查连接状态——处理信息前
	*/
	mux.Lock()
	if *status != SUCCESS {
		mux.Unlock()
		fmt.Println("not allowed to send ")
		return
	}
	mux.Unlock()

	packet := accessMessagePacket(str, mode)
	/*
		第二次检查连接状态——处理信息后
	*/
	mux.Lock()
	if *status != SUCCESS {
		mux.Unlock()
		fmt.Println("not allowed to send ")
		return
	}
	mux.Unlock()

	_, err := conn.Write(packet)
	if err != nil {

		/*
			发送失败，修改连接状态
		*/
		mux.Lock()
		if *status == UNKNOWN {
			mux.Unlock()
			return
		} else {
			*status = UNKNOWN
			mux.Unlock()
			fmt.Println("send failed ")
			return
		}
	}

}
func receive(conn net.Conn, status *STATUS) {
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf[:])
		if err != nil {
			mux.Lock()
			*status = UNKNOWN
			mux.Unlock()
			//fmt.Println("receive failed,err: ", err)
			return
		}
		rawData := utils.Decipher(buf[32:336], buf[:32], 304)
		length := int(rawData[33])*100 + int(rawData[34])*10 + int(rawData[35])
		msg := string(rawData[36 : 36+length])
		switch rawData[32] {
		case 0:
			switch msg {
			case "ok":
				live.Lock()
				ALIVE = SUCCESS
				live.Unlock()
				//fmt.Println("alive")
			default:
				fmt.Println(msg)
			}
		case 1:
			fmt.Println(msg)
		}
	}
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
		os.Stdout.Write([]byte{0x1B, 0x5B, 0x33, 0x3B, 0x4A, 0x1B, 0x5B, 0x48, 0x1B, 0x5B, 0x32, 0x4A})
		var choice int
		fmt.Println("Hello, what do you want to do ?")
		fmt.Println("	0:	register")
		fmt.Println("	1:	global chat	(less than 100 Chinese character/300 English letter) ")
		fmt.Println("	2:	send file")
		fmt.Println("choose 0 or 1:")
		_, err := fmt.Scanln(&choice)
		if err != nil {
			fmt.Println("input failed			")
			continue
		}
		switch choice {
		case 0:
			chat(serverNATAddr+":"+serverNATPort, COMMAND)
		case 1:
			chat(serverNATAddr+":"+serverNATPort, GLOBAL)
		case 2:
			os.Stdout.Write([]byte{0x1B, 0x5B, 0x33, 0x3B, 0x4A, 0x1B, 0x5B, 0x48, 0x1B, 0x5B, 0x32, 0x4A})
			fmt.Println("不好意思还没做,嘿嘿嘿")
			time.Sleep(time.Second * time.Duration(2))
		}
	}
}
func keepLive(conn net.Conn, status *STATUS) {
	go func(conn net.Conn, status *STATUS) {
		for {
			live.Lock()
			if ALIVE != SUCCESS {
				live.Unlock()
				mux.Lock()
				*status = SERVER
				mux.Unlock()
				//fmt.Println("not alive")
				return
			}
			live.Unlock()
			live.Lock()
			ALIVE = UNKNOWN
			live.Unlock()
			if *status == LOCALE {
				*status = SUCCESS
			}
			time.Sleep(time.Second * time.Duration(10))
		}
	}(conn, status)

	time.Sleep(time.Second)
	go func(conn net.Conn, status *STATUS) {
		for {
			if *status != SUCCESS {
				return
			}
			//fmt.Println("running auto keep alive")
			wg.Add(1)
			go send("alive", conn, status, COMMAND)
			time.Sleep(time.Second * time.Duration(5))
		}
	}(conn, status)

}
func chat(DialServer string, mode MODE) {
	var status STATUS = LOCALE
	for {
		//os.Stdout.Write([]byte{0x1B, 0x5B, 0x33, 0x3B, 0x4A, 0x1B, 0x5B, 0x48, 0x1B, 0x5B, 0x32, 0x4A})
		/*
			连接远程穿透端口
		*/
		conn, err := net.Dial("tcp", DialServer)
		if err != nil {
			fmt.Println("can not start chat, err:				", err)
			mux.Lock()
			status = UNKNOWN
			mux.Unlock()
			break
		}

		/*
			向服务器发起第一次连接
		*/
		fmt.Println("connecting to server...")
		wg.Add(1)
		go send("alive", conn, new(STATUS), mode)
		go receive(conn, &status)
		time.Sleep(time.Second)
		keepLive(conn, &status)
		time.Sleep(time.Second)
		mux.Lock()
		if status != SUCCESS {
			mux.Unlock()
			/*
				第一次连接失败，返回主页面
			*/
			fmt.Println("connect failed")
			break
		}
		mux.Unlock()
		defer conn.Close()

		/*
			循环发送消息
		*/
		inputReader := bufio.NewReader(os.Stdin)
		for {
			var inputInfo string
			fmt.Println("You can send now:               ")
			for {
				input, err := inputReader.ReadString('\n')
				if err != nil {
					fmt.Println("input err							")
				}
				inputInfo = strings.Trim(input, "\r\n")
				if inputInfo != "" {
					break
				}
			}
			if strings.ToUpper(inputInfo) == "Q" {
				mux.Lock()
				status = LOCALE
				mux.Unlock()
				break
			} else {

				/*
					发消息前检测连接状态
				*/
				mux.Lock()
				if status != SUCCESS {
					mux.Unlock()
					fmt.Println("connect failed ")
					break
				}
				mux.Unlock()
				wg.Add(1)
				go send(inputInfo, conn, &status, mode)
			}
		}

		/*
			等待各个发送任务结束
		*/
		fmt.Println("quiting")
		wg.Wait()

		return
	}
}
