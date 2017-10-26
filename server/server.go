package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mxi4oyu/MoonSocket/protocol"
)

var msgList = make([]string, 100)
var sign = 0

type AgentMsg struct {
	Session int64
	IP      string
	Message string
	Status  string
}

//定义CheckError方法，避免写太多到 if err!=nil
func CheckError(err error) {

	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s", err.Error())

		os.Exit(1)
	}

}

//自定义log
func Log(v ...interface{}) {

	log.Println(v...)
}

var ch1 = make(chan int, 1)
var ch2 = make(chan string, 1)

var TmpList = make([]string, 10)

func main() {

	server_listener, err := net.Listen("tcp", "localhost:8848")

	CheckError(err)

	defer server_listener.Close()

	Log("Waiting for clients connect")

	go getConn()

	for {
		new_conn, err := server_listener.Accept()

		CheckError(err)

		go ServerMsgHandler(new_conn)
	}

}

func getConn() {
	http.HandleFunc("/", sayhelloName)
	err := http.ListenAndServe(":8849", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func sayhelloName(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	key := ""
	value := ""
	for k, v := range r.Form {
		if k != "" && strings.Join(v, "") != "" {
			key = k
			value = strings.Join(v, "")
		}
	}
	log.Println(key + "--" + value)
	//监听8849端口判断批量getStatus动作
	if key == "key" && value == "getstatus" {
		//向Agent发送Stauts请求
		sign = 1
		go func() {
			ch1 <- 2
		}()

		//log.Println("sayhello函数中:" + <-ch2)
		fmt.Fprintf(w, <-ch2)

	}
}

//服务端消息处理
func ServerMsgHandler(conn net.Conn) {

	//存储被截断的数据
	tmpbuf := make([]byte, 0)
	buf := make([]byte, 1024)

	defer conn.Close()

	//接收解包
	readchan := make(chan []byte, 16)
	go ReadChan(readchan)

	for {
		//读取客户端发来的消息
		n, err := conn.Read(buf)

		if err != nil {

			fmt.Println("connection close")
			return
		}

		//解包
		tmpbuf = protocol.Depack(append(tmpbuf, buf[:n]...))
		//tmpbuf = buf

		fmt.Println("client say:", string(tmpbuf))

		//判断解析json
		var agentmsg AgentMsg
		json.Unmarshal([]byte(string(tmpbuf)), &agentmsg)
		//agentmsg 客户端发送的struct
		if agentmsg.Status == "IAMRunning" {
			log.Println("有Agent正在运行")
			log.Println("Agent IP:" + agentmsg.IP)
			sign1 := 0
			for _, w := range TmpList {
				if w == agentmsg.IP {
					sign1 = 1
				}
			}
			if sign1 == 0 {
				TmpList = append(TmpList, agentmsg.IP)
			}
		}
		ll := ""
		for _, v := range TmpList {
			if v != "" && ll != "" {
				ll = ll + "," + v
			} else if v != "" {
				ll = v
			}
		}
		statusMsg := `{"Agent":"` + ll + `"}`
		//返回的信息
		log.Println(statusMsg)

		if len(ll) > 0 {
			go func(statusMsg string) {
				ch2 <- statusMsg
			}(statusMsg)
		}

		Msg := tmpbuf

		//向客户端发送消息
		go WriteMsgToClient(conn)

		go func() {
			if <-ch1 == 2 {
				WriteMsgToClient2(conn)
			}
		}()

		beatch := make(chan byte)
		//心跳计时，默认30秒
		go HeartBeat(conn, beatch, 30)
		//检测每次Client是否有数据传来
		go HeartChanHandler(Msg, beatch)

	}

}

//处理心跳,根据HeartChanHandler判断Client是否在设定时间内发来信息
func HeartBeat(conn net.Conn, heartChan chan byte, timeout int) {
	select {
	case hc := <-heartChan:
		Log("<-heartChan:", string(hc))
		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		break
	case <-time.After(time.Second * 30):
		Log("timeout")
		conn.Close()
	}
}

//服务端向客户端发送消息
func WriteMsgToClient(conn net.Conn) {
	talk := "RUN"
	smsg := protocol.Enpack([]byte(talk))
	conn.Write(smsg)
}

func WriteMsgToClient2(conn net.Conn) {
	talk := "GETSTATUS"
	smsg := protocol.Enpack([]byte(talk))
	conn.Write(smsg)
	sign = 0
}

//处理心跳channel
func HeartChanHandler(n []byte, beatch chan byte) {
	for _, v := range n {
		beatch <- v
		log.Println(v)
	}
	close(beatch)
}

//从channell中读取数据
func ReadChan(readchan chan []byte) {

	for {
		select {
		case data := <-readchan:
			Log(string(data))
		}
	}
}
