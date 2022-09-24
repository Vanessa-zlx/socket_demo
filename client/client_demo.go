package client

import (
	"fmt"
	"io"
	"learnGo/package/files"
	"learnGo/package/socket_demo/utils"
	"net"
	"strconv"
)

func CallAt(DialServer string) net.Conn {
	conn, err := net.Dial("tcp", DialServer)
	if err != nil {
		fmt.Println("cant start connect, err:				", err, "\nconn:", conn)
		return nil
	}
	return conn
}

func SendFileDemo(path string, conn net.Conn) {

	file := files.OpenFile(path)
	if file == nil {
		return
	}
	//info := strconv.FormatInt(files.GetFileSize(path), 10) + files.GetName(path)
	//infoBytes := []byte(info)
	//temp := bytes.Buffer{}
	//temp.Write(utils.Sha256Bytes(infoBytes))
	//temp.Write([]byte{0, 0, 0, 0})
	//temp.Write(infoBytes)
	//_, err := conn.Write(temp.Bytes())
	//if err != nil {
	//	fmt.Println("init failed", err)
	//	return
	//}

	buf := make([]byte, 900)
	count := 0
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF && n == 0 {
				fmt.Println("done<!>")
				break
			} else {
				fmt.Println("read failed<!>")
				return
			}
		}
		packet := utils.ProcessFilePacket(buf[:n], utils.FILE)
		m, err := conn.Write(packet)
		if err != nil {
			fmt.Println("write err:", err)
			return
		} else if m != 1000 {
			fmt.Println("write less err<!>" + ":" + strconv.Itoa(m))
			return
		}
		count++
		fmt.Println(count)
	}
}
