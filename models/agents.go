package models

import (
	"TyrantSocket/request"
	"encoding/json"
	"log"
	"strings"
)

type AgentServer struct {
	Agent  string `json:"agent"`
	Status string `json:"status"`
}


func (this *AgentServer) GetStatus() ([]AgentServer, int, error) {
	var mServiceReq = new(request.ServiceReq)
	r := mServiceReq.QueryStatus()
	var agentServer AgentServer
	var strList = []string{"192.168.0.9","192.168.0.10","192.168.0.11","192.168.0.12","192.168.0.13","192.168.0.14","192.168.0.15","192.168.0.16","192.168.0.17","192.168.0.18","192.168.0.19","192.168.0.20","192.168.0.21","192.168.0.22","192.168.0.23","192.168.0.24","192.168.0.25","192.168.0.26","192.168.0.27","192.168.0.28","192.168.0.29","192.168.0.30","192.168.0.31","192.168.0.32","192.168.0.33","192.168.0.34","192.168.0.35","192.168.0.36","192.168.0.37","192.168.0.38","192.168.0.39","192.168.0.40","192.168.0.41","192.168.0.42","192.168.0.43","192.168.0.44","192.168.0.45","192.168.0.46"}
	err := json.Unmarshal(r, &agentServer)
	if err != nil {
		log.Println("Json解析错误")
		return nil, 0, err
	}
	sss := strings.Split(agentServer.Agent, ",")
	agentList := make([]AgentServer, 0)
	for _, v := range sss {
		var agent1 AgentServer
		agent1.Agent = v
		agent1.Status = "Running"
		agentList = append(agentList, agent1)
	}
	for _,w := range strList{
		sign := 0
		for _ ,y := range sss{
			if y == w {
				sign = 1
			}
		}
		if sign == 0{
			var agent1 AgentServer
			agent1.Agent = w
			agent1.Status = "Stopped"
			agentList = append(agentList, agent1)
		}
	}
	return agentList, len(agentList), err
}
