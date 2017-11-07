package main

import (
	"GoRadar/core"
	"GoRadar/lib"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/benmanns/goworker"
	"github.com/levigross/grequests"
	"github.com/bipabo1l/TyrantSocket/protocol"
)

//ch1 为-1时，与server端停止通信
var chSign = make(chan int, 1)
//var IPrange = make([]string, 0, 100)
//var IPPortrange = make([]string, 0, 100)
var IPRange = ""
var IPPortrange = ""

// 扫描
func ScanActivityTask(queue string, args ...interface{}) error {
	fmt.Println("调用队列:" + queue)
	ip_range := args[0].(string)
	//IPrange = append(IPrange, ip_range)
	IPRange = ip_range
	is_limit_scan_rate := args[1].(bool)
	sa := core.NewScanActivity()
	sa.Scanner(ip_range, is_limit_scan_rate)
	return nil
}
func ScanPortTask(queue string, args ...interface{}) error {
	fmt.Println("调用队列:" + queue)
	ip_range := args[0].(string)
	//IPPortrange = append(IPPortrange, ip_range)
	IPPortrange = ip_range
	scan_mode := args[1].(string)
	is_limit_scan_rate := args[2].(bool)
	sa := core.NewScanPort()
	sa.Scan(ip_range, scan_mode, is_limit_scan_rate)
	return nil
}

var (
	version      = "1.0.6"
	download_url = ""
)

func Version_validate(c chan string) bool {
	resp, err := grequests.Get("http://43.226.164.114/version.txt", nil)
	// You can modify the request by passing an optional RequestOptions struct
	if err != nil {
		fmt.Println("Validate version error: Unable to make request ")
		return false
	} else {
		new_version := resp.String()[0:5]
		fmt.Println("new_version:" + new_version)
		fmt.Println("version:" + version)
		if version < new_version {
			os_name := runtime.GOOS
			if os_name == "linux" {
				download_url = "http://43.226.164.114/linux/" + new_version
			} else if os_name == "windows" {
				download_url = "http://43.226.164.114/windows/" + new_version
			}
			download, _ := Download_new_agent(download_url, os_name)
			if download == true {
				c <- "new"
				fmt.Println("New agent version found !")
				return true
			} else {
				c <- "old"
				return false
			}
		} else {
			return false
		}
	}
}

func Download_new_agent(url string, os_name string) (bool, error) {
	res, err := http.Get(url)
	if err != nil {
		return false, err
	}
	var (
		file_name string
	)
	if os_name == "windows" {
		file_name = "Agent.exe"
	} else if os_name == "linux" {
		file_name = "Agent"
	} else {
		file_name = "Agent"
	}
	cmd := exec.Command("rm", "-rf", file_name)
	cmd.Run()
	f, err := os.Create(file_name)
	if err != nil {
		return false, err
	}
	_, er := io.Copy(f, res.Body)
	if er != nil {
		return false, er
	}
	if os_name == "linux" {
		cmdd := exec.Command("chmod", "+x", file_name)
		cmdd.Run()
	}
	res.Body.Close()
	f.Close()
	return true, er

}

func Restart_process() {
	filePath, _ := filepath.Abs(os.Args[0])
	cmd := exec.Command(filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalf("GracefulRestart: Failed to launch, error: %v", err)
	}
}

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

	settings := goworker.WorkerSettings{
		URI:            dsn_addr,
		Connections:    100,
		Queues:         []string{"ScanActivityQueue", "ScanPortQuene"},
		UseNumber:      true,
		ExitOnComplete: false,
		Concurrency:    50,
		Namespace:      "goradar:",
		Interval:       5.0,
	}

	goworker.SetSettings(settings)
	//read scan option
	activeswitch, _ := cfg.GetString("agent_default", "scanactivity")
	portswitch, _ := cfg.GetString("agent_default", "scanport")
	if activeswitch == "yes" {
		goworker.Register("ScanActivityTask", ScanActivityTask)
		fmt.Println("Start active scan !")
	} else if activeswitch == "no" {
		fmt.Println("Doesn't start active scan !")
	} else {
		fmt.Println("Error: config anget->scanactivity param error, only 'yes' or 'no' allowed")
	}
	if portswitch == "yes" {
		goworker.Register("ScanPortTask", ScanPortTask)
		fmt.Println("Start ports scan !")
	} else if portswitch == "no" {
		fmt.Println("Doesn't start ports scan !")
	} else {
		fmt.Println("Error: Config anget->scanport param error,only 'yes' or 'no' allowed")
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

	signals := make(chan string)
	go func() {
		for {
			Version_validate(signals)
			time.Sleep(1 * time.Minute)
		}

	}()

	go func() {
		for {
			log.Println("GOWoker Running")
			err := goworker.Work()
			if err != nil {
				fmt.Println("Error:", err)
			}

		}
	}()

	go func() {
		sendMsgToServer()
	}()

	for {
		select {
		case signal := <-signals:
			if signal == "new" {
				Restart_process()
				return
			}
		case <-time.After(time.Second * 10):
			fmt.Println("timeout, check again...")
			continue
		}
	}
}

func sendMsgToServer() {
	//动态传入服务端IP和端口号
	service := "192.168.0.8:8848"
	//service := "127.0.0.1:8848"

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)

	CheckError(err)

	for {

		conn, err := net.DialTCP("tcp", nil, tcpAddr)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal error:%s", err.Error())
		} else {
			defer conn.Close()

			//连接服务器端成功
			doWork(conn)
		}

		time.Sleep(3 * time.Second)

	}
}

//定义CheckError方法，避免写太多到 if err!=nil
func CheckError(err error) {

	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s", err.Error())

		os.Exit(1)
	}

}

//解决断线重连问题
func doWork(conn net.Conn) error {
	ch := make(chan int, 100)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case stat := <-ch:
			if stat == 2 {
				return errors.New("None Msg")
			}
		case <-ticker.C:
			ch <- 1
			go ClientMsgHandler(conn, ch)
			go ReadMsg(conn, ch)
		case <-time.After(time.Second * 2):
			defer conn.Close()
			fmt.Println("timeout")
		}

	}

	return nil
}

//客户端消息处理
func ClientMsgHandler(conn net.Conn, ch chan int) {

	<-ch
	//获取当前时间
	//msg := "+++++++++++++++++++++++++++++++++++++"
	SendMsg(conn, "rrrrr")
	go ReadMsg(conn, ch)

}

func GetSession() string {
	gs1 := time.Now().Unix()
	gs2 := strconv.FormatInt(gs1, 10)
	return gs2
}

//接收服务端发来的消息
func ReadMsg(conn net.Conn, ch chan int) {
	<-ch

	//存储被截断的数据
	tmpbuf := make([]byte, 0)
	buf := make([]byte, 1024)

	//将信息解包
	n, _ := conn.Read(buf)
	tmpbuf = protocol.Depack(append(tmpbuf, buf[:n]...))
	msg := string(tmpbuf)
	fmt.Println("server say:", msg)
	if len(msg) == 0 {
		//服务端无返回信息
		ch <- 2
	} else {
		//接收到了服务器端发来的非空包，证明已经和服务器连接成功
		if msg == "RUN" {
			log.Println("收到服务器命令：发送Running包")
			go func() {
				SendMsg(conn, msg)
				time.Sleep(time.Second * 2)
			}()
		} else if msg == "GETSTATUS" {
			//服务器端想知道当前在没在工作
			log.Println("Server端请求了解本Agent状态")
			go SendRunMsg(conn, msg)
		} else if msg == "STOP" {
			//服务器让当前agent停止
			log.Println("服务器让当前agent停止")
			go func() {
				StopRunMsg(conn, msg)
				chSign <- -1
			}()
			//go StopRunMsg(conn, msg)
			//conn.Close()
		} else if msg == "BEGIN" {
			//服务器让当前agent停止
			log.Println("服务器让当前agent重新连接")
			go BeginRunMsg(conn, msg)
			//conn.Close()
		} else if msg == "IPRANGE" {
			//服务器让当前agent停止
			log.Println("服务器需要当前IPrange")
			log.Println(IPRange)
			go IPRangeMsg(conn, IPRange)
		}
	}
}

//向服务端发送消息
func SendMsg(conn net.Conn, msg string) {

	session := GetSession()

	words := "{\"Session\":" + session + ",\"IP\":\"" + GetMyIP() + "\",\"Message\":\"" + msg + "\",\"Status\":\"" + "running" + "\"}"
	//将信息封包
	smsg := protocol.Enpack([]byte(words))
	conn.Write(smsg)

}

//向服务端发送消息
func SendRunMsg(conn net.Conn, msg string) {

	session := GetSession()
	words := "{\"Session\":" + session + ",\"IP\":\"" + GetMyIP() + "\",\"Message\":\"" + msg + "\",\"Status\":\"" + "IAMRunning" + "\"}"
	//将信息封包
	smsg := protocol.Enpack([]byte(words))
	conn.Write(smsg)

}

//向服务端发送消息
func StopRunMsg(conn net.Conn, msg string) {

	session := GetSession()
	words := "{\"Session\":" + session + ",\"IP\":\"" + GetMyIP() + "\",\"Message\":\"" + msg + "\",\"Status\":\"" + "STOPPED" + "\"}"
	//将信息封包
	smsg := protocol.Enpack([]byte(words))
	conn.Write(smsg)
	log.Println("----------------------------------")
	conn.Close()
}

//向服务端发送消息
func BeginRunMsg(conn net.Conn, msg string) {
	session := GetSession()
	words := "{\"Session\":" + session + ",\"IP\":\"" + GetMyIP() + "\",\"Message\":\"" + msg + "\",\"Status\":\"" + "Begin" + "\"}"
	//将信息封包
	smsg := protocol.Enpack([]byte(words))
	conn.Write(smsg)
	log.Println("----------------------------------")
}

func IPRangeMsg(conn net.Conn, msg string) {
	session := GetSession()
	words := "{\"Session\":" + session + ",\"IP\":\"" + GetMyIP() + "\",\"Message\":\"" + msg + "\",\"Status\":\"" + "IPRange" + "\"}"
	//将信息封包
	smsg := protocol.Enpack([]byte(words))
	conn.Write(smsg)
	log.Println("----------------------------------")
}

func GetMyIP() string {
	addrs, err := net.InterfaceAddrs()
	ip := ""
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
			}

		}
	}
	return ip
}

func get_external() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	log.Println("++++++++++++++++++++++++++++++++++++++++++++")
	log.Println(string(content))
	return string(content)

}
