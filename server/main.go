package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

var listenAddr, listenPort = "127.0.0.1", "5040"

//var listenAddr, listenPort = "[::1]", "20000

//var wg sync.WaitGroup

//var m sync.Mutex
func Process(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	var buf [256]byte
	for {
		//read
		n, err := reader.Read(buf[:])
		if err != nil {
			fmt.Println("read from client", conn.RemoteAddr(), "failed,err: ", err, "\nconnection was automatically closed")
			return
		}

		//possess
		recvStr := string(buf[:n])
		fmt.Println("message from client:", recvStr)

		//reply
		_, err = conn.Write([]byte(recvStr + "ok"))
		if err != nil {
			fmt.Println("reply to client", conn.RemoteAddr(), "failed,err: ", err, "\nconnection was automatically closed")
			return
		}
	}
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
			serveIPv6()
			break
		case 1:

			break
		default:
			continue
		}

	}
}
