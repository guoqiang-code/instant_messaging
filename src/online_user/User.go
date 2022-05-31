package online_user

import (
	"fmt"
	srcServer "instant_messaging/src/server"
	"net"
)

type User struct {
	Title   string
	Addr    string
	Channel chan string
	Conn    net.Conn
	Server  *srcServer.Server
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

// Online 用户上线
func (u *User) Online() {
	u.Server.MapLock.Lock()
	u.Server.OnlineMap[u.Title] = u
	u.Server.MapLock.Unlock()

	// 发送广播消息
	go u.Server.BoardCast(u, "已经上线………………")
}

// Offline 用户下线
func (u *User) Offline() {
	u.Server.MapLock.Lock()
	delete(u.Server.OnlineMap, u.Title)
	u.Server.MapLock.Unlock()

	// 发送广播消息
	go u.Server.BoardCast(u, "已下线………………")
}

// SendMsg 用户发送全局消息
func (u *User) SendMsg(conn net.Conn) {
	bytes := make([]byte, 4096)
	read, err := conn.Read(bytes)
	if err != nil {
		fmt.Println("消息接收异常：err: ", err)
	}
	if read == 0 {
		u.Offline()
		return
	}
	u.Server.BoardCast(u, string(bytes[:read-1]))
}
