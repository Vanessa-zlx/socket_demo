package utils

import (
	"fmt"
	"github.com/howeyc/gopass"
	"strings"
)

func GetEmailInput() string {
	var email string
	for {
		_, err := fmt.Scanln(&email)
		if err != nil && !strings.Contains(email, "@") {
			fmt.Println("fail to read or wrong form of email")
			return ""
		}
		if strings.ToUpper(email) == "Q" {
			return ""
		}
		fmt.Println("Are you sure '", email, " 'is your email ?(Y/others)")
		var a string
		fmt.Scanln(&a)
		if strings.ToUpper(a) == "Y" {
			return email
		} else {
			continue
		}
	}
}
func GetPassHashInput() string {
	for {
		password, err := gopass.GetPasswdMasked()
		if strings.ToUpper(string(password)) == "Q" || err != nil {
			return ""
		}
		if len(password) < 8 {
			continue
		}
		passHash := string(Sha256Bytes(password))
		return passHash
	}
}
