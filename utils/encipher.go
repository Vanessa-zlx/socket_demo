package utils

import (
	"bytes"
	"math/rand"
	"time"
)

const ENCODE_LEVEL int = 6

func Encipher(Data []byte, code []byte, n int) []byte {
	if len(Data) < n {
		length := len(Data)
		gap := n - len(Data)
		buffer := bytes.Buffer{}
		buffer.Write(Data)
		empty := make([]byte, n-len(Data))
		buffer.Write(empty)
		Data = buffer.Bytes()
		buffer.Reset()
		rand.Seed(time.Now().UnixMilli())
		for i := 0; i < gap; i++ {
			Data[i+length] = byte(rand.Int() % 256)
		}
	}
	for i := 0; i < ENCODE_LEVEL; i++ {
		buffer := bytes.Buffer{}
		var index byte
		index = code[i] % byte(n)
		buffer.Write(Data[index:n])
		buffer.Write(Data[:index])
		Data = buffer.Bytes()
		buffer = bytes.Buffer{}
		if n < 16 {
			index = code[i] % byte(n)
		} else {
			index = code[i] % 16
		}
		buffer.Write(Data[index:n])
		buffer.Write(Data[:index])
		Data = buffer.Bytes()
	}
	for i := 0; i < len(Data); i++ {
		Data[i] = byte(int(Data[i]+code[0]) % 256)
	}
	return Data
}
func Decipher(ciperData []byte, code []byte, n int) []byte {
	for i := 0; i < len(ciperData); i++ {
		if ciperData[i] < code[0] {
			ciperData[i] = byte((int(ciperData[i]) + 256) - int(code[0]))
		} else {
			ciperData[i] -= code[0]
		}
	}
	for i := 0; i < ENCODE_LEVEL/2; i++ {
		code[i], code[ENCODE_LEVEL-1-i] = code[ENCODE_LEVEL-1-i], code[i]
	}
	for i := 0; i < ENCODE_LEVEL; i++ {
		buffer := bytes.Buffer{}
		var index byte
		index = byte(n) - code[i]%byte(n)
		buffer.Write(ciperData[index:n])
		buffer.Write(ciperData[:index])
		ciperData = buffer.Bytes()
		buffer = bytes.Buffer{}
		if n < 16 {
			index = byte(n) - code[i]%byte(n)
		} else {
			index = byte(n) - code[i]%16
		}
		buffer.Write(ciperData[index:n])
		buffer.Write(ciperData[:index])
		ciperData = buffer.Bytes()
	}
	return ciperData
}
