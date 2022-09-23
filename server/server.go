package server

import (
	"fmt"
	"learnGo/package/redis_demo"
	"learnGo/package/socket_demo/utils"
	"net"
	"strconv"
)

type client struct {
	Id       string
	password string
	Name     string
	Email    string
	conn     net.Conn
}

type STATUS = int

const (
	SUCCESS STATUS = iota
	FAIL
	DOING
	SUSPEND
	ILLEGAL
	UNKNOWN
)

type Transaction struct {
	cli       *client
	status    STATUS
	msg       string
	mgsLength int
	conn      net.Conn
}

func NewTransaction(conn net.Conn) *Transaction {
	this := new(Transaction)
	this.cli = new(client)
	this.status = DOING
	this.conn = conn
	return this
}

func (*Transaction) command(this *Transaction) {
	switch this.msg {
	case "alive":
		if !send("ok", this.conn, utils.COMMAND) {
			fmt.Println("cant alive to:" + this.conn.RemoteAddr().String())
			//无法alive通信，关闭连接
			this.status = FAIL
		}
	default:
		fmt.Println("Unknown meg:" + this.msg)
	}
}
func (*Transaction) registerFirst(this *Transaction) {
	passHash := this.msg[:32]
	email := this.msg[32:this.mgsLength]
	_, b := redis_demo.Get(email)
	if b {
		//近300s发送过
		send("exist", this.conn, utils.REGISTER1)
		return
	}
	num, ok := redis_demo.Get("users")
	if !ok {
		return
	}
	id, err := strconv.Atoi(num)
	if err != nil {
		return
	}
	id++
	redis_demo.Set(email, "user"+strconv.Itoa(id))
	//email->idString
	idString, ok := redis_demo.Get(email)
	if !ok {
		redis_demo.Del(email)
		return
	}
	redis_demo.HMSet(idString, "email", email, "passHash", passHash)
	//idString->{email,passHash}

	code := utils.SendCodeMessage(email)
	redis_demo.HSet(idString, "code", code)
	//idString->{email,passHash,code}
	redis_demo.Set(code, idString)
	//code->idString
	redis_demo.Expire(code, 300)
	redis_demo.Expire(email, 300)
	redis_demo.Expire(idString, 300)
	send("code", this.conn, utils.REGISTER1)
	redis_demo.SafeSave()
}
func (*Transaction) registerSecond(this *Transaction) {
	code := this.msg[:6]
	passHash := this.msg[6:38]
	clientID, err := strconv.Atoi(this.msg[38:this.mgsLength])
	if err != nil {
		return
	}
	fmt.Println("code:" + code)

	IDString, ok := redis_demo.Get(code)
	if IDString == "" {
		send("wrong", this.conn, utils.REGISTER2)
		return
	}
	codeID, err := strconv.Atoi(IDString[4:])
	if codeID != clientID || err != nil {
		go send("wrong", this.conn, utils.REGISTER2)
		return
	}
	email, o := redis_demo.HGet(IDString, "email")
	if !ok || !o {
		return
	}
	c, b := redis_demo.HGet(IDString, "passHash")
	if !b {
		return
	}
	if c == passHash {
		redis_demo.IncrBy("users", 1)
		redis_demo.Persist(IDString)
		redis_demo.Persist(email)
		redis_demo.Del(code)
		redis_demo.HDel(IDString, "code")
		send("ok", this.conn, utils.REGISTER2)
		redis_demo.SafeSave()
		fmt.Println("login:" + IDString)
	}
}
func (*Transaction) login(this *Transaction) {
	email := this.msg[32:]
	passHash := this.msg[:32]
	idString, _ := redis_demo.Get(email)
	if idString == "" {
		send("notFound", this.conn, utils.LOGIN)
		return
	}
	redisHash, _ := redis_demo.HGet(idString, "passHash")
	if redisHash == passHash {
		send("auth", this.conn, utils.LOGIN)
		this.cli.Id, _ = redis_demo.Get(email)
		this.cli.conn = this.conn
		this.cli.password, _ = redis_demo.HGet(this.cli.Id, "passHash")
		fmt.Println("login:" +
			this.cli.Email + this.cli.Id)
	} else {
		send("wrongPass", this.conn, utils.LOGIN)
	}
}
func (*Transaction) global(this *Transaction) {
	fmt.Println("global msg from", this.cli.Id, this.msg)
}
func (*Transaction) CheckAndServe(mode utils.MODE, msg string, length int, this *Transaction) {
	this.mgsLength = length
	this.msg = msg
	if (mode != utils.COMMAND && mode != utils.REGISTER1 &&
		mode != utils.REGISTER2 && mode != utils.LOGIN) && this.cli.Id == "" {
		send("illegal", this.conn, utils.NOTICE)
		return
	}
	switch mode {
	case utils.COMMAND:
		this.command(this)
	case utils.REGISTER1:
		this.registerFirst(this)
	case utils.REGISTER2:
		this.registerSecond(this)
	case utils.LOGIN:
		this.login(this)
	case utils.GLOBAL:
		this.global(this)
	}
}
