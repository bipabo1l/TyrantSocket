# TyrantSocket
一台Server管理多台Agent，建立双向通信，自定义发送命令，全双工模式，支持断线重连机制，基于github.com/mxi4oyu/MoonSocket,增加Server端主动发送命令、http接口化调用、基于beego的web页面展示。

Installation

Use the go command:

$ go get github.com/bipabo1l/TyrantSocket

usage

语法

go run server/server.go
go run client/agent.go
说明

分别执行在server和agent目录下的server.go agent.go,然后访问http://localhost:8849 web页面。
目前支持批量获取Agent状态，json形式返回。接口url:http://localhost:8849?key=getstatus

示例

![image](http://ovnsp3bhk.bkt.clouddn.com/Snipaste_2017-11-07_16-15-17.png)

2017-10-27
v1.1 支持web页面查看当前开放的agent

v1.2 可以查看当前开放的agent正在扫描的ip段信息

