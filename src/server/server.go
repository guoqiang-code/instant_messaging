package server

import (
	"fmt"
	curUser "instant_messaging/src/online_user"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Ip        string `json:"ip"`
	Post      string `json:"post"`
	OnlineMap map[string]*curUser.User
	MapLock   sync.Mutex
	Channel   chan string
}

func NewServer(ip, post string) *Server {
	s := &Server{
		Ip:        ip,
		Post:      post,
		OnlineMap: make(map[string]*curUser.User),
		Channel:   make(chan string),
	}
	return s
}

// Start 开启服务
func (s *Server) Start() {
	fmt.Println("服务器监听中……………………")
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.Ip, s.Post))
	if err != nil {
		fmt.Println("监听异常：", err)
	}

	defer func(listen net.Listener) {
		err := listen.Close()
		if err != nil {
			fmt.Println("监听器关闭异常………………")
		}
	}(listen)

	// 开启协程监听端口号
	for {
		accept, err := listen.Accept()
		if err != nil {
			fmt.Println("listen.Accept()异常：", err)
			continue
		}
		// 建立连接
		go func(conn net.Conn) {
			fmt.Println("连接建立成功：LocalAddr", conn.LocalAddr(), " RemoteAddr", conn.RemoteAddr())
			// 存储在线用户信息
			newUser := curUser.NewUser(conn)
			// 用户上线通知
			s.Online(newUser)

			// 监听用户发送的消息
			go s.SendMsg(newUser)
		}(accept)

		// 监听server
		go s.ListenServerMsg()
	}
}

// BoardCast 广播消息，当前登录用户
func (s *Server) BoardCast(u *curUser.User, msg string) {
	sendMsg := "title: " + u.Title + "addr:" + u.Addr + "" + " msg:" + msg
	s.Channel <- sendMsg
}

// ListenServerMsg 监听server消息 有消息就广播给全体在线用户
func (s *Server) ListenServerMsg() {
	for {
		msg := <-s.Channel
		fmt.Println("监听到的信息：", msg)
		s.MapLock.Lock()
		for _, user := range s.OnlineMap {
			user.Channel <- msg
		}
		s.MapLock.Unlock()
	}
}

// Online 用户上线
func (s *Server) Online(u *curUser.User) {
	s.MapLock.Lock()
	s.OnlineMap[u.Title] = u
	s.MapLock.Unlock()

	// 发送广播消息
	go s.BoardCast(u, "已经上线………………")
}

// Offline 用户下线
func (s *Server) Offline(u *curUser.User) {
	s.MapLock.Lock()
	delete(s.OnlineMap, u.Title)
	s.MapLock.Unlock()

	// 发送广播消息
	go s.BoardCast(u, "已下线………………")
}

// SendMsg 用户发送消息
func (s *Server) SendMsg(u *curUser.User) {
	isLive := make(chan bool)
	for {
		go s.timeoutExit(isLive, u)

		bytes := make([]byte, 4096)
		read, err := u.Conn.Read(bytes)
		if err != nil {
			fmt.Println("消息接收异常：err: ", err)
		}
		if read == 0 {
			s.Offline(u)
			return
		}
		msg := string(bytes[:read])
		isLive <- true
		if msg == "search" {
			fmt.Println("用户查询在线用户列表……………………")
			s.MapLock.Lock()
			sendClientMsg := ""
			for _, user := range s.OnlineMap {
				sendClientMsg += "\n用户为：title【" + user.Title + "】,addr为：【" + user.Addr + "】的用户在线"
			}
			u.SendMsgClient(sendClientMsg)
			s.MapLock.Unlock()
		} else if len(msg) > 7 && msg[:6] == "rename" {
			s.MapLock.Lock()
			delete(s.OnlineMap, u.Title)
			u.Title = msg[7:]
			s.OnlineMap[u.Title] = u
			s.MapLock.Unlock()
			u.SendMsgClient("用户名称修改成功……")
		} else if len(msg) > 3 && msg[:2] == "to" {
			split := strings.Split(msg, " ")
			if len(split) != 3 {
				fmt.Println("消息发送格式不匹配…………")
			}
			user, ok := s.OnlineMap[split[1]]
			if !ok {
				fmt.Println("指定用户：" + split[1] + " 不存在")
				u.SendMsgClient("指定用户：" + split[1] + "不存在")
			} else {
				user.SendMsgClient("用户：" + u.Title + "给您发送消息-->" + split[2] + "\n")
			}
		} else {
			s.BoardCast(u, msg)
		}

	}
}

func (s *Server) timeoutExit(isLive chan bool, user *curUser.User) bool {
	select {
	case <-isLive:
	//不做处理，表示有消息
	case <-time.After(time.Second * 120):
		//表示超过十秒未发送消息，超时踢出
		fmt.Println("用户 " + user.Title + " 超时下线")
		user.SendMsgClient("超时下线")
		err := user.Conn.Close()
		if err != nil {
			return true
		}
		close(user.Channel)
		return true
	}
	return false
}
