package online_user

import (
	"fmt"
	"net"
)

type User struct {
	Title   string
	Addr    string
	Channel chan string
	Conn    net.Conn
}

func NewUser(conn net.Conn) *User {
	addr := conn.RemoteAddr().String()
	u := &User{
		Title:   addr,
		Addr:    addr,
		Channel: make(chan string),
		Conn:    conn,
	}
	go u.listen()

	return u
}

func (u *User) listen() {
	fmt.Println("user监听中")
	for {
		msg := <-u.Channel
		_, err := u.Conn.Write([]byte(msg + "\n"))
		if err != nil {
			continue
		}
	}
}

// SendMsgClient 向用户端发送消息
func (u *User) SendMsgClient(msg string) {
	_, err := u.Conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("发送消息给用户端失败………………", err)
		return
	}
}
