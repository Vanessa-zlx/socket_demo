package server

import (
	"fmt"
	"io"
	"learnGo/package/socket_demo/utils"
	"net"
)

func ServeAt(serverAddr string) {
	listen, err := net.Listen("tcp", serverAddr)
	fmt.Println("listening at :", serverAddr)
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
		ReceiveFile(conn)
		fmt.Println("processing has begun\n" +
			"waiting for connection apply ")
	}
}

func ReceiveFile(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1000)
	for {
		n, err := conn.Read(buf[:])
		if err != nil {
			if err != io.EOF {
				break
			} else {
				fmt.Println("receive failed ,err: ", err)
				return
			}
		}
		fmt.Println("receive:", buf[:n])
		rawData := utils.Decipher(buf[32:n], buf[:32], 968)
		//fmt.Println("raw:", rawData)
		fmt.Println(rawData[1], rawData[2], rawData[3])
		length := int(rawData[1])*100 + int(rawData[2])*10 + int(rawData[3])*1
		piece := string(rawData[4 : 4+length])
		fmt.Println(piece)
	}
}
func storeFile(data []byte) {

}
