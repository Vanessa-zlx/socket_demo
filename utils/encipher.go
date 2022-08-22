package utils

import "bytes"

func Encipher(Data []byte, code []byte, n int) []byte {
	buffer := bytes.Buffer{}
	for i := 0; i < 6; i++ {
		var index byte
		if n < 16 {
			index = code[i] % byte(n)
		} else {
			index = code[i]
		}
		buffer.Write(Data[index:])
		buffer.Write(Data[:index])
		Data = buffer.Bytes()
	}
	return Data
}
