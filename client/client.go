package client

import (
	"bufio"
	"fmt"
	"learnGo/package/socket_demo/utils"
	"net"
	"strings"
	"time"
)

type STATUS = int

const (
	SUCCESS STATUS = iota
	SERVER
	LOCALE
	UNKNOWN
)

type Client struct {
	auth     chan string
	Id       string
	password string
	Name     string
	Email    string
	status   STATUS
	alive    STATUS
	conn     net.Conn
	check    chan string
	waitTime int64
}

func (*Client) send(str string, mode utils.MODE, this *Client) {
	defer wg.Done()
	/*
		第一次检查连接状态——处理信息前
	*/
	STAT.Lock()
	if this.status != SUCCESS && this.status != LOCALE {
		STAT.Unlock()
		fmt.Println("not allowed to send ")
		return
	}
	//允许local发送
	STAT.Unlock()

	packet := utils.ProcessMessagePacket(str, mode)
	/*
		第二次检查连接状态——处理信息后
	*/
	STAT.Lock()
	if this.status != SUCCESS && this.status != LOCALE {
		STAT.Unlock()
		fmt.Println("not allowed to send ")
		return
	}
	STAT.Unlock()

	_, err := this.conn.Write(packet)
	if err != nil {
		/*
			发送失败，修改连接状态
		*/
		STAT.Lock()
		if this.status == UNKNOWN {
			STAT.Unlock()
			return
		} else {
			this.status = UNKNOWN
			STAT.Unlock()
			fmt.Println("send failed ")
			return
		}
	}

}
func (*Client) receive(this *Client) {
	for {
		if this.status != LOCALE && this.status != SUCCESS {
			return
		}
		buf := make([]byte, 1024)
		_, err := this.conn.Read(buf[:])
		if err != nil {
			this.changeStatus(UNKNOWN, this)
			return
		}
		rawData := utils.Decipher(buf[32:336], buf[:32], 304)
		length := int(rawData[33])*100 + int(rawData[34])*10 + int(rawData[35])
		msg := string(rawData[36 : 36+length])
		switch rawData[32] {
		case utils.NOTICE:
			fmt.Println("notice: ", msg)
		case utils.REGISTER1:
			this.check <- msg
		case utils.REGISTER2:
			this.check <- msg
		case utils.LOGIN:
			this.auth <- msg
		case utils.COMMAND:
			switch msg {
			case "ok":
				LIVE.Lock()
				this.alive = SUCCESS
				LIVE.Unlock()
			default:
				fmt.Println(msg)
			}
		default:
			fmt.Println("group message: ", msg)
		}
	}
}
func (*Client) keepLive(this *Client) {
	go func(this *Client) {
		for {

			LIVE.Lock()
			if this.alive != SUCCESS {
				LIVE.Unlock()
				STAT.Lock()
				this.status = SERVER
				STAT.Unlock()
				return
			}
			this.alive = UNKNOWN
			this.status = SUCCESS
			LIVE.Unlock()

			//检查成功，置反
			time.Sleep(time.Second * time.Duration(30))
			//三十秒检查一次
		}
	}(this)
	time.Sleep(time.Second)
	go func(this *Client) {
		for {

			STAT.Lock()
			if this.status != SUCCESS {
				STAT.Unlock()
				return
			}
			STAT.Unlock()

			wg.Add(1)
			go this.send("alive", utils.COMMAND, this)
			time.Sleep(time.Second * time.Duration(15))
			//十五秒发送一次
		}
	}(this)
}
func (*Client) timeOut(duration time.Duration, this *Client) {
	for {
		time.Sleep(duration / 10)
		t := time.Now().Unix()
		if this.waitTime == 0 {
			return
		} else if time.Duration(t-this.waitTime) > duration {
			this.check <- "timeout"
			this.waitTime = 0
			return
		}
	}

}
func (*Client) isConnValid(this *Client) bool {
	STAT.Lock()
	if this.status != SUCCESS {
		STAT.Unlock()
		fmt.Println("connect failed ")
		return false
	}
	STAT.Unlock()
	return true
}
func (*Client) changeStatus(status STATUS, this *Client) {
	STAT.Lock()
	this.status = status
	STAT.Unlock()
}
func (*Client) init(this *Client) {
	fmt.Println("connecting to server...")
	wg.Add(1)
	//每一次send就add一次waitGroup
	go this.send("alive", utils.COMMAND, this)
	go this.receive(this)
	time.Sleep(time.Second)
	this.keepLive(this)
}
func (*Client) registerWindowCheck(choice int, this *Client) bool {
	switch choice {
	case 0:
		if !register(this) {
			fmt.Println("register canceled")
		}
		//注册成功不自动登录
	case 1:
		if !login(this) {
			break
		}
		return true
	//登录成功可以切换窗口
	case -1:
		return true
	default:
	}
	return false

}
func (*Client) inputAndSend(this *Client, reader *bufio.Reader) (canceled bool) {
	var inputInfo string
	fmt.Println("You can send now:               ")
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("input err							")
		}
		inputInfo = strings.Trim(input, "\r\n")
		if inputInfo != "" {
			break
		}
	}
	if strings.ToUpper(inputInfo) == "Q" {
		return true
	} else {
		/*发消息前检测连接状态*/
		if !this.isConnValid(this) {
			return true
			//因为可能还有消息没发完，外层循环可等待处理完
		}
		wg.Add(1)
		go this.send(inputInfo, utils.GLOBAL, this)
		return false
	}
}
