package utils

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// QQ 邮箱：
	// SMTP 服务器地址：smtp.qq.com（SSL协议端口：465/994 | 非SSL协议端口：25）
	// 163 邮箱：
	// SMTP 服务器地址：smtp.163.com（端口：25）
	host     = "smtp.qq.com"
	port     = 465
	userName = "2724327805@qq.com"
	password = "tsgmftwiofimdeah"
)

func GetMyIPV6() string {
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
func GetCode() string {
	var code string
	//for i := 0; i < 6; i++ {
	rand.Seed(time.Now().Unix() + 2004)
	codeNum := rand.Int() % 1000000
	if codeNum < 100000 {
		codeNum += 100000
	}
	code += strconv.Itoa(codeNum)
	//time.Sleep(time.Second)
	//}
	//message := `
	//<p> 你好 %s,</p>
	//
	//	<p style="text-indent:2em">您的验证码是: ` + code + ` ,十五分钟内有效。</p>
	//	<p style="text-indent:2em">请勿将此码告诉他人!</P>
	//	<p style="text-indent:2em">Best Wishes!</p>
	//`
	return code
}
func sendTextEmail(msg, sendTo string) {

	m := gomail.NewMessage()
	m.SetHeader("From", userName)               // 发件人
	m.SetHeader("From", "QCT"+"<"+userName+">") // 增加发件人别名
	m.SetHeader("To", sendTo)                   // 收件人，可以多个收件人，但必须使用相同的 SMTP 连接
	//m.SetHeader("Cc", "******@qq.com")                  // 抄送，可以多个
	//m.SetHeader("Bcc", "******@qq.com")                 // 暗送，可以多个
	m.SetHeader("Subject", "QCT VERIFICATION") // 邮件主题
	// text/html 的意思是将文件的 content-type 设置为 text/html 的形式，浏览器在获取到这种文件时会自动调用html的解析器对文件进行相应的处理。
	// 可以通过 text/html 处理文本格式进行特殊处理，如换行、缩进、加粗等等
	m.SetBody("text/html", fmt.Sprintf(msg, "User"))
	// text/plain的意思是将文件设置为纯文本的形式，浏览器在获取到这种文件时并不会对其进行处理
	// m.SetBody("text/plain", "纯文本")
	// m.Attach("test.sh")   // 附件文件，可以是文件，照片，视频等等
	// m.Attach("lolcatVideo.mp4") // 视频
	// m.Attach("lolcat.jpg") // 照片
	d := gomail.NewDialer(
		host,
		port,
		userName,
		password,
	)
	// 关闭SSL协议认证
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
func SendCodeMessage(sendTo string) string {
	code := GetCode()
	message := `
    <p> 你好 %s,</p>

		<p style="text-indent:2em">您的验证码是: ` + code + ` ,5分钟内有效。</p>
		<p style="text-indent:2em">请勿将此码告诉他人!</P>
		<p style="text-indent:2em">Best Wishes!</p>
	`
	sendTextEmail(message, sendTo)
	return code
}

func mac() {
	// 获取本机的MAC地址
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Poor soul, here is what you got: " + err.Error())
	}
	inter := interfaces[0]
	fmt.Println(inter.Name)
	fmt.Println("MAC", inter.HardwareAddr)
}
func getExternalIPv4() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(resp.Body)
	//s := buf.String()
	return string(content)
}
