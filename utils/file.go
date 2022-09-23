package utils

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
)

type MODE = byte

const (
	COMMAND MODE = iota
	NOTICE
	GLOBAL
	FILE
	REGISTER1 = 101
	REGISTER2 = 102
	LOGIN     = 202
	NOTFOUND  = 505 + iota
	AUTH
	WRONGPASS
)

func FileInfo(fileName, fileHash string, fileSize int) []byte {
	hash := []byte(fileHash)
	buffer := bytes.Buffer{}
	buffer.Read([]byte(fileHash))                             //64
	buffer.Read(Encipher([]byte(fileName), hash, 64))         //64
	buffer.Read(Encipher([]byte(string(fileSize)), hash, 16)) //16
	return buffer.Bytes()
}
func Sha256File(filepath string) string {
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
func ProcessMessagePacket(msg string, mode MODE) []byte {
	size := len([]byte(msg))
	raw := bytes.Buffer{}
	raw.Write([]byte{byte(mode), byte(size / 100), byte(size % 100 / 10), byte(size % 10)})
	raw.Write([]byte(msg)[:len(msg)])
	//功能、长度和msg一起加密
	cipherData := Encipher(raw.Bytes(), []byte(Sha256String(msg)), 304)
	buffer := bytes.Buffer{}
	buffer.Write(Sha256String(msg)) //32
	buffer.Write(cipherData)        //304
	return buffer.Bytes()
}
func Sha256String(str string) []byte {
	h := sha256.New()
	h.Write([]byte(str))
	sum := h.Sum(nil)
	return sum
}
func Sha256Bytes(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	sum := h.Sum(nil)
	return sum
}
