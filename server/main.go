package server

import (
	"fmt"
	"learnGo/package/redis_demo"
	"learnGo/package/socket_demo/utils"
	"net"
	"sync"
)

var (
	listenAddr, listenPort = "127.0.0.1", "20000"
)
var count = 0
var CT sync.Mutex

func send(str string, conn net.Conn, mode utils.MODE) bool {
	packet := utils.ProcessMessagePacket(str, mode)
	_, err := conn.Write(packet)
	if err != nil {
		fmt.Print("\rsend failed ,						")
		return false
	}
	return true
}
func receiveAndAnswer(conn net.Conn) {

	CT.Lock()
	count++
	CT.Unlock()

	defer closeConnection(conn)

	var action = NewTransaction(conn)
	for {
		if action.status != SUCCESS && action.status != DOING {
			return
		}
		buf := make([]byte, 1024)
		n, err := conn.Read(buf[:])
		if n < 50 {
			fmt.Println(string(buf[:]))
			conn.Write([]byte("你在赣神魔"))
			return
		}
		if err != nil {
			fmt.Println("receive from client failed ,err: ", err, "\ndestroy:"+action.cli.Id)
			return
		}
		rawData := utils.Decipher(buf[32:336], buf[:32], 304)
		length := int(rawData[33])*100 + int(rawData[34])*10 + int(rawData[35])
		msg := string(rawData[36 : 36+length])
		//32位hash，一位功能码，三位长度，300位数据
		action.CheckAndServe(rawData[32], msg, length, action)
	}
}
func closeConnection(conn net.Conn) {
	conn.Close()
	CT.Lock()
	count--
	CT.Unlock()
}
func clearUsers() bool {
	ok := redis_demo.FlushDB()
	if !ok {
		return false
	}
	ok = redis_demo.Set("users", 0)
	if !ok {
		return false
	}
	return true
}
func serve(serverIP string) {
	listen, err := net.Listen("tcp", serverIP)
	fmt.Println("listening at :", serverIP)
	if err != nil {
		fmt.Println("listen failed ,err: ", err)
		return
	}
	fmt.Println("waiting for connection apply ")
	for {
		//不断获取连接
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("accept connection failed ,err: ", err)
			fmt.Println("automatically retry accepting...")
			continue
		}
		fmt.Println("accepted from client:", conn.RemoteAddr())
		go receiveAndAnswer(conn)
		fmt.Println("processing has begun\n" +
			"waiting for connection apply ")
	}
}

func serveIPV4() {
	serve(listenAddr + ":" + listenPort)
}

func Start() {
	serveIPV4()
}

//func main() {
//	Start()
//	//clearUsers()
//}

func serveIPv6() {
	serverIPv6 := "[" + utils.GetMyIPV6() + "]:20000"
	serve(serverIPv6)
}
func Process(conn net.Conn) {
	//defer conn.Close()
	receiveAndAnswer(conn)
	//reader := bufio.NewReader(conn)
	//var buf [384]byte
	//for {
	//	//read
	//	n, err := reader.Read(buf[:])
	//	if err != nil {
	//		fmt.Println("read from client", conn.RemoteAddr(), "failed,err: ", err, "\nconnection was automatically closed")
	//		return
	//	}
	//
	//	//possess
	//	recvStr := string(buf[:n])
	//	fmt.Println("message from client:", recvStr)
	//reply
	//	_, err = conn.Write([]byte(recvStr + "ok"))
	//	if err != nil {
	//		fmt.Println("reply to client", conn.RemoteAddr(), "failed,err: ", err, "\nconnection was automatically closed")
	//		return
	//	}
	//}
}
