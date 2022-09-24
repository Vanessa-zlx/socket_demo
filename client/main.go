package client

import (
	"bufio"
	"fmt"
	"io"
	"learnGo/package/socket_demo/utils"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

//var serverNATAddr, serverNATPort = "218.89.171.148", "25118"
var serverNATAddr, serverNATPort = "127.0.0.1", "20000"

var wg sync.WaitGroup
var STAT sync.Mutex
var LIVE sync.Mutex

//var serverIPv6Addr, serverIPv6Port = []string{"2409:8a60:1e74:15f1:6183:a810:5d4a:ea7d",
//	"2409:8a60:1e74:15f1:82cb:c469:ed53:2"}, "20000"

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
func register(this *Client) bool {

	fmt.Println("Please input you email:(q to quit)")
	email := utils.GetEmailInput()
	if email == "" {
		return false
	}

	fmt.Println("Please set your password(no less than 8 letters,q to quit)")
	passHash := utils.GetPassHashInput()
	if passHash == "" {
		return false
	}
	fmt.Println("Please input you password again: (q to quit)")
	for {
		passHashAgain := utils.GetPassHashInput()
		if passHashAgain == "" {
			return false
		}
		if passHashAgain == passHash {
			break
		}
	}

	this.check = make(chan string, 1)
	wg.Add(1)
	this.send(passHash+email, utils.REGISTER1, this)

	this.waitTime = time.Now().Unix()
	go this.timeOut(time.Duration(300)*time.Second, this)
	switch <-this.check {
	case "exist":
		fmt.Println("email already exist !")
		return false
	case "code":
	default:
		fmt.Println("sorry this email cant be registered")
		return false
	}
	fmt.Println("Please check your mailbox for new or last email(300 seconds) ")
	for {
		fmt.Println(" Then input the right auth code:(q to quit)")
		var code string
		fmt.Scanln(&code)
		if strings.ToUpper(code) == "Q" {
			return false
		}
		if len(code) != 6 {
			continue
		}
		wg.Add(1)
		this.send(code+passHash+this.Id, utils.REGISTER2, this)
		if <-this.check == "ok" {
			fmt.Println("Register success!")
			time.Sleep(time.Second)
			this.waitTime = 0
			break
		}
		fmt.Println("wrong code ! try again?")
		time.Sleep(time.Second * time.Duration(2))
		continue
	}

	return true
}
func login(this *Client) bool {
	fmt.Println("Please input you email:(q to quit)")
	email := utils.GetEmailInput()
	if email == "" {
		return false
	}
	fmt.Println("Please input you password:(q to quit)")
	passHash := utils.GetPassHashInput()
	if passHash == "" {
		return false
	}

	wg.Add(1)
	go this.send(passHash+email, utils.LOGIN, this)

	switch <-this.auth {
	case "notFound":
		fmt.Println("this email have not been registered!")
		return false
	case "wrongPass":
		fmt.Println("wrong password !")
		return false
	case "auth":
		fmt.Println("login success !")
		this.auth <- "true"
	}

	return true
}
func (*Client) connect(DialServer string, this *Client) {
	this.status = LOCALE
	this.alive = UNKNOWN
	this.auth = make(chan string, 1)
	/*连接远程穿透端口*/
	//os.Stdout.Write([]byte{0x1B, 0x5B, 0x33, 0x3B, 0x4A, 0x1B, 0x5B, 0x48, 0x1B, 0x5B, 0x32, 0x4A})
	conn, err := net.Dial("tcp", DialServer)
	if err != nil {
		fmt.Println("cant start connect, err:				", err, "\nconn:", conn)
		this.changeStatus(UNKNOWN, this)
		return
	}
	this.conn = conn
	/*向服务器发起第一次连接*/
	this.init(this)
	if !this.isConnValid(this) {
		return
	}

	defer conn.Close()
	/*登录注册*/
	var choice int
	for {
		os.Stdout.Write([]byte{0x1B, 0x5B, 0x33, 0x3B, 0x4A, 0x1B, 0x5B, 0x48, 0x1B, 0x5B, 0x32, 0x4A})
		fmt.Println("[----->]\nhello,come and join us ! :\n" + ">>register\t0\n>>login    \t1\t(others to quit)\n[<-----]")
		_, err := fmt.Scanln(&choice)
		if err != nil {
			continue
		}
		if this.registerWindowCheck(choice, this) {
			break
		}
	}
	inputReader := bufio.NewReader(os.Stdin)

	/*循环发送消息*/
	for {
		if this.inputAndSend(this, inputReader) {
			break
		}
	}
	/*等待各个发送任务结束*/
	fmt.Println("quiting...")
	wg.Wait()
	return

}
func Start() {
	cli := new(Client)
	cli.connect(serverNATAddr+":"+serverNATPort, cli)
}
func main() {
	Start()
}
