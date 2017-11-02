package request

import (
	"io/ioutil"
	"net/http"
	"log"
	"time"
)

type ServiceReq struct {
}

func (d *ServiceReq) QueryStatus() []byte {
	client := &http.Client{
		//设置超时机制
		Timeout: 3 * time.Second,
	}
	url := "http://192.168.0.8:8849/?key=getstatus"
	//url := "http://127.0.0.1:8849/?key=getstatus"
	req, err := http.NewRequest("GET", url, nil)
	log.Println(err)
	if err != nil {
		// handle error
		log.Println("Request Error")
	}
	resp, err := client.Do(req)
	log.Println(resp)
	if resp != nil {
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Println("rrrrrrrrrrrrrrrrrrrr")
			return nil
		}
		return body
	}
	return nil

}

func (d *ServiceReq) StopStatus(ip string) error {
	client := &http.Client{
		//设置超时机制
		Timeout: 3 * time.Second,
	}
	url := "http://192.168.0.8:8849/?key=stop" + ip
	log.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	log.Println(err)
	if err != nil {
		// handle error
		log.Println("Request Error")
	}
	resp, err := client.Do(req)
	log.Println(resp)
	return err

}

func (d *ServiceReq) GetRange() []byte {
	client := &http.Client{
		//设置超时机制
		Timeout: 3 * time.Second,
	}
	url := "http://192.168.0.8:8849/?key=getiprange"
	log.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	log.Println(err)
	if err != nil {
		// handle error
		log.Println("Request Error")
	}
	resp, err := client.Do(req)
	log.Println(resp)
	if resp != nil {
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Println("rrrrrrrrrrrrrrrrrrrr")
			return nil
		}
		return body
	}
	return nil

}