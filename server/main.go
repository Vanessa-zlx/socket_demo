package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"learnGo/package/socket_demo/utils"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

var listenAddr, listenPort = "127.0.0.1", "20000"

//var listenAddr, listenPort = "[::1]", "20000

//var wg sync.WaitGroup

//var m sync.Mutex
func Process(conn net.Conn) {
	defer conn.Close()
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
func receiveAndAnswer(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf[:])
		if err != nil {
			fmt.Println("recv from client failed ,err: ", err)
			return
		}
		rawData := utils.Decipher(buf[32:336], buf[:32], 304)
		length := int(rawData[33])*100 + int(rawData[34])*10 + int(rawData[35])
		msg := string(rawData[36 : 36+length])
		switch msg {
		case "alive":
			if !send("ok", conn) {
				return
			} else {
				continue
			}
		}

		fmt.Println("client: ", string(msg))
	}
}
func accessMessagePacket(msg string) []byte {
	size := len([]byte(msg))
	raw := bytes.Buffer{}
	raw.Write([]byte{0, byte(size / 100), byte(size % 100 / 10), byte(size % 10)})
	raw.Write([]byte(msg)[:len(msg)])
	cipherData := utils.Encipher(raw.Bytes(), []byte(utils.Sha256String(msg)), 304)
	buffer := bytes.Buffer{}
	buffer.Write(utils.Sha256String(msg)) //32
	buffer.Write(cipherData)              //304
	return buffer.Bytes()
}
func send(str string, conn net.Conn) bool {
	packet := accessMessagePacket(str)
	_, err := conn.Write(packet)
	if err != nil {
		fmt.Print("\rsend failed ,						")
		return false
	}
	return true
}
func sha256File() {
	f, err := os.Open("test.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%x\n", h.Sum(nil))
}
func serve(serverIP string) {
	listen, err := net.Listen("tcp", serverIP)
	fmt.Println("listening at :", serverIP)
	if err != nil {
		fmt.Println("listen failed ,err: ", err)
		return
	}
	for {
		fmt.Println("waiting for connection apply ")
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("accept connection failed ,err: ", err)
			fmt.Println("automatically retry accepting...")
			continue
		}
		fmt.Println("accepted connection from client:", conn.RemoteAddr())
		go Process(conn)
		fmt.Println("processing has begun")
	}
}
func serveIPV4() {
	serve(listenAddr + ":" + listenPort)
}
func serveIPv6() {
	serverIPv6 := "[" + getMyIPV6() + "]:20000"
	serve(serverIPv6)
}
func getMyIPV6() string {
	s, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range s {
		i := regexp.MustCompile(`(\w+:){7}\w+`).FindString(a.String())
		if strings.Count(i, ":") == 7 {
			return i
		}
	}
	return ""
}
func main() {
	for {
		var choice int
		fmt.Println("choose 0 (ipv4) or 1 (ipv6) (suggest 0):")
		fmt.Scanln(&choice)
		switch choice {
		case 0:
			serveIPV4()
			break
		case 1:

			break
		default:
			continue
		}

	}
}
