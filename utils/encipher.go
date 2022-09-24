package utils

import (
	"bytes"
	"math"
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
		empty := make([]byte, gap)
		buffer.Write(empty)
		Data = buffer.Bytes()
		buffer.Reset()
		rand.Seed(time.Now().UnixMilli())
		for i := 0; i < gap; i++ {
			Data[i+length] = byte(rand.Int() % 256)
		}
	} else {
		return nil
	}
	factor := int(math.Pow(float64(ENCODE_LEVEL), 5))
	for i := 0; i < ENCODE_LEVEL; i++ {
		buffer := bytes.Buffer{}
		var index int
		//index = int(code[i]) * factor % n
		////原byte（n）
		//buffer.Write(Data[index:n])
		//buffer.Write(Data[:index])
		//Data = buffer.Bytes()
		//buffer = bytes.Buffer{}
		if n < 16 {
			index = int(code[i]) * factor % 16
		} else {
			index = int(code[i]) * factor % n
		}
		buffer.Write(Reverse(Data[index:n]))
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
	factor := int(math.Pow(float64(ENCODE_LEVEL), 5))

	for i := 0; i < ENCODE_LEVEL; i++ {
		buffer := bytes.Buffer{}
		var index int
		//index = n - int(code[i])*factor%n
		//
		//buffer.Write(ciperData[index:n])
		//buffer.Write(ciperData[:index])
		//ciperData = buffer.Bytes()
		//buffer = bytes.Buffer{}
		if n < 16 {
			index = n - int(code[i])*factor%16
		} else {
			index = n - int(code[i])*factor%n
		}
		buffer.Write(ciperData[index:n])
		buffer.Write(Reverse(ciperData[:index]))

		ciperData = buffer.Bytes()
	}
	return ciperData
}
func Reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
