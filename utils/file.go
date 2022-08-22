package utils

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
)

func FileInfo(fileName, fileHash string, fileSize int) []byte {
	hash := []byte(fileHash)
	buffer := bytes.Buffer{}
	buffer.Read([]byte(fileHash))                             //64
	buffer.Read(Encipher([]byte(fileName), hash, 64))         //64
	buffer.Read(Encipher([]byte(string(fileSize)), hash, 16)) //16
	return buffer.Bytes()
}

func Sha256String(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
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
