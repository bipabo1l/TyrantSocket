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
	return agentList, len(agentList), err
}
