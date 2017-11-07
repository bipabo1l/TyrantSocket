package main

import (
	"GoRadar/core"
	"GoRadar/lib"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/benmanns/goworker"
	"gopkg.in/mgo.v2/bson"
	"github.com/bipabo1l/TyrantSocket/protocol"
	"net/http"
)

var waitgroup sync.WaitGroup
var mIPRangePool lib.MongoDriver
var is_private int
var scan_mode_param string
var is_limit_scan_rate bool
var IPRange1 string

type IPRangePool struct {
	IPRange string `bson:"ip_range"`
}

//记录所有Agent
var clientArr = make(map[string]net.Conn)
var ipRangeArr = make(map[string]string)

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

func init() {

	cfg := lib.NewConfigUtil("")
	redis_host, _ := cfg.GetString("redis_default", "host")
	redis_port, _ := cfg.GetString("redis_default", "port")
	redis_pass, _ := cfg.GetString("redis_default", "pass")
	redis_db, _ := cfg.GetString("redis_default", "db")

	var dsn_addr string
	if redis_pass != "" {
		dsn_addr = fmt.Sprintf("redis://:%s@%s:%s/%s", redis_pass, redis_host, redis_port, redis_db)
	} else {
		dsn_addr = fmt.Sprintf("redis://%s:%s/%s", redis_host, redis_port, redis_db)
	}

	// 初始化
	settings := goworker.WorkerSettings{
		URI:            dsn_addr,
		Connections:    100,
		Queues:         []string{"scannerQueue", "ScanPortQuene"},
		UseNumber:      true,
		ExitOnComplete: false,
		Concurrency:    2,
		Namespace:      "goradar:",
		Interval:       5.0,
	}

	goworker.SetSettings(settings)

	// 初始化数据库连接
	mIPRangePool = lib.MongoDriver{TableName: "ip_range_pool"}
	err := mIPRangePool.Init()
	if err != nil {
		fmt.Println("INIT MONGODB ERRPR:" + err.Error())
	}

	//initialize deploy mode
	deploy, _ := cfg.GetString("server_default", "deploy")
	if deploy == "inner" {

		is_private = 1
		scan_mode_param = "q"
		is_limit_scan_rate = true
	} else if deploy == "outer" {
		is_private = 2
		scan_mode_param = "d"
		is_limit_scan_rate = false
	} else {
		fmt.Println("config server_defaul->deploy error: only 'inner' or 'outer' allowed")
		os.Exit(1)
	}
}

func main() {

	// 加入守护进程机制

	if os.Getppid() != 1 {
		//判断当其是否是子进程，当父进程return之后，子进程会被 系统1 号进程接管
		filePath, _ := filepath.Abs(os.Args[0])
		//将命令行参数中执行文件路径转换成可用路径
		cmd := exec.Command(filePath)
		//将其他命令传入生成出的进程
		cmd.Stdin = os.Stdin
		//给新进程设置文件描述符，可以重定向到文件中
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		//开始执行新进程，不等待新进程退出
		cmd.Start()
		return
	}

	//mode_param := flag.String("mode", "q", "for quick scan: -mode q , for detail scan -mode d ;default -mode q ")
	//if *mode_param != "q" && *mode_param != "d" {
	//	fmt.Println("for quick scan: -mode q , for detail scan -mode d ; default -mode q")
	//	os.Exit(1)
	//}

	waitgroup.Add(3)

	go ListenMsg()

	// 清理昨天数据(5分钟)
	go func() {
		for {
			timer1 := time.NewTimer(time.Minute * 5)
			<-timer1.C
			cna := core.NewClearNotActivity()
			cna.Clear()
			fmt.Println("清理完成")
		}
	}()

	// 添加任务(10分钟)
	go func() {

		// 首次运行
		one_run := true

		for {

			if one_run == false {
				timer1 := time.NewTimer(time.Minute * 10)
				<-timer1.C
			}

			// 添加扫描任务
			IPRangePool := new([]IPRangePool)
			ip_range_pool, err := mIPRangePool.NewTable()
			if err == nil {

				// TODO 由于扫内网存活会造成大量Arp请求，目前暂时只扫描外网IP段
				ip_range_pool.Find(bson.M{"is_private": is_private}).All(IPRangePool)
				for _, ip_range := range *IPRangePool {

					fmt.Println("添加一条存活探测任务:" + ip_range.IPRange)
					// 不阻塞,增加扫描任务
					goworker.Enqueue(&goworker.Job{
						Queue: "ScanActivityQueue",
						Payload: goworker.Payload{
							Class: "ScanActivityTask",
							Args:  []interface{}{string(ip_range.IPRange), is_limit_scan_rate},
						},
					})

				}

			} else {
				fmt.Println("ERROR:" + err.Error())
			}

			one_run = false
		}
	}()

	// 添加端口扫描任务(5)
	go func() {

		// 首次运行
		one_run := true
		for {

			if one_run == false {
				timer1 := time.NewTimer((time.Hour * 24) * 2)
				<-timer1.C
			}

			// 添加扫描任务
			IPRangePool := new([]IPRangePool)
			ip_range_pool, err := mIPRangePool.NewTable()
			if err == nil {

				// TODO 由于扫外网占用大量session表，目前暂时只扫描内网
				ip_range_pool.Find(bson.M{"is_private": is_private}).All(IPRangePool)
				for _, ip_range := range *IPRangePool {
					fmt.Println("添加一条端口扫描任务:" + ip_range.IPRange)
					// 不阻塞,增加扫描任务
					goworker.Enqueue(&goworker.Job{
						Queue: "ScanPortQuene",
						Payload: goworker.Payload{
							Class: "ScanPortTask",
							Args:  []interface{}{string(ip_range.IPRange), scan_mode_param, is_limit_scan_rate},
						},
					})
				}

			} else {
				fmt.Println("ERROR:" + err.Error())
			}

			one_run = false
		}
	}()

	waitgroup.Wait()

	fmt.Println("添加任务完成")
}

////自定义log
//func Log(v ...interface{}) {
//
//	log.Println(v...)
//}

func ListenMsg() {
	server_listener, err := net.Listen("tcp", "192.168.0.8:8848")
	//server_listener, err := net.Listen("tcp", "127.0.0.1:8848")

	CheckError(err)

	defer server_listener.Close()

	Log("Waiting for clients connect")

	go getConn()

	for {
		new_conn, err := server_listener.Accept()

		clientIPList := strings.Split(new_conn.RemoteAddr().String(), ":")

		clientIP := new_conn.RemoteAddr().String()
		if len(clientIPList) > 0 {
			clientIP = clientIPList[0]
		}
		log.Println(clientIP)
		clientArr[clientIP] = new_conn

		CheckError(err)

		log.Println(new_conn.RemoteAddr().String() + " 上线了")

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

type IPList struct {
	IPPort   string   `json:"ipport"`
	ConnList net.Conn `json:"netconn"`
}

type IPListMap struct {
	IPListMap []IPList
}

func sayhelloName(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	key := ""
	value := ""
	sign := 0
	for k, v := range r.Form {
		if k != "" && strings.Join(v, "") != "" {
			key = k
			value = strings.Join(v, "")
		}
	}

	//监听8849端口判断批量getStatus动作
	if key == "key" && value == "getstatus" {
		data, err := json.Marshal(clientArr)
		if err != nil {
			log.Println(err)
			return
		}
		s := strings.Split(string(data[:]), ",")
		ll := ""
		if len(s) > 0 {
			for _, w := range s {
				k := strings.Split(w, ":")
				if len(k) > 0 && ll != "" {
					ll = ll + "," + k[0]
					ll = strings.Replace(ll, "{", "", -1)
					ll = strings.Replace(ll, `"`, "", -1)
					ll = strings.Replace(ll, `}`, "", -1)
				} else if len(k) > 0 {
					ll = k[0]
					ll = strings.Replace(ll, "{", "", -1)
					ll = strings.Replace(ll, `"`, "", -1)
					ll = strings.Replace(ll, `}`, "", -1)
				}
			}
		}

		fmt.Fprintf(w, `{"Agent":"`+ll+`"}`)
	} else if key == "key" && strings.Count(value, "")-1 >= 4 {
		if protocol.Substr2(value, 0, 4) == "stop" {
			sign = 1
			ip := protocol.Substr2(value, 4, len(value))

			log.Println("-----------------------------")
			log.Println(clientArr)

			findConn := clientArr[ip]

			go StopClient(findConn)

			fmt.Fprintf(w, "1")
		}

		if strings.Count(value, "")-1 >= 5 && sign == 0 {
			if protocol.Substr2(value, 0, 5) == "start" {
				sign = 1
				ip := protocol.Substr2(value, 5, len(value))

				log.Println("-----------------------------")
				log.Println(clientArr)

				findConn := clientArr[ip]

				go BeginClient(findConn)

				fmt.Fprintf(w, "1")
			}
		}

		if strings.Count(value, "")-1 >= 10 && sign == 0 {
			if protocol.Substr2(value, 0, 10) == "getiprange" {
				log.Println("服务端请求当前扫描的IPRange")
				ip := protocol.Substr2(value, 10, len(value))
				log.Println("服务端接收到ip:")
				log.Println(ip)
				fmt.Println(clientArr)
				fmt.Println(ipRangeArr)
				if _, ok := ipRangeArr[ip]; ok {
					go IPRangeClient(clientArr[ip])
				}
				if len(ipRangeArr) > 0 {
					//by, err := json.Marshal(ipRangeArr)
					//if err != nil {
					//	log.Println(err)
					//}
					//fmt.Fprintf(w, string(by))
					if _, ok := ipRangeArr[ip]; ok {
						fmt.Fprintf(w, `{"IPrange":"`+ipRangeArr[ip]+`"}`)
					}

				} else {
					fmt.Fprintf(w, `{"IPrange":"`+`"}`)
				}

			}
		}
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
			removeClient(conn)
			conn.Close()
			return
		}

		//解包
		tmpbuf = protocol.Depack(append(tmpbuf, buf[:n]...))
		//tmpbuf = buf

		fmt.Println("client say:", string(tmpbuf))

		//判断解析json
		var agentmsg AgentMsg
		json.Unmarshal([]byte(string(tmpbuf)), &agentmsg)
		Msg := tmpbuf
		log.Println(agentmsg.Status)
		if agentmsg.Status == "STOPPED" {
			removeClient(conn)
		}

		if agentmsg.Status == "IPRange" {
			log.Println("Agent.Sattus")
			IPRange1 = agentmsg.Message
			ipRangeArr[agentmsg.IP] = agentmsg.Message
		}

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
	log.Println(conn.RemoteAddr())
	talk := "RUN"
	smsg := protocol.Enpack([]byte(talk))
	conn.Write(smsg)
}

func WriteMsgToClient2(conn net.Conn) {
	talk := "GETSTATUS"
	smsg := protocol.Enpack([]byte(talk))
	conn.Write(smsg)
}

//Server表示不想跟您通信咯
func StopClient(conn net.Conn) {
	talk := "STOP"
	smsg := protocol.Enpack([]byte(talk))
	conn.Write(smsg)
}

//Server表示不想跟您通信咯
func BeginClient(conn net.Conn) {
	talk := "BEGIN"
	smsg := protocol.Enpack([]byte(talk))
	conn.Write(smsg)
}

func IPRangeClient(conn net.Conn) {
	talk := "IPRANGE"
	smsg := protocol.Enpack([]byte(talk))
	conn.Write(smsg)
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

func removeClient(new_conn net.Conn) {
	log.Println(clientArr)
	log.Println(new_conn.RemoteAddr().String() + " 已经阵亡")

	clientIPList := strings.Split(new_conn.RemoteAddr().String(), ":")

	clientIP := new_conn.RemoteAddr().String()
	if len(clientIPList) > 0 {
		clientIP = clientIPList[0]
	}
	log.Println(clientIP)

	delete(clientArr, clientIP)
	log.Println("delete close conn")
	log.Println(clientArr)
	return
}
