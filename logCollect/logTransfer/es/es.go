package es

import (
	"context"
	"fmt"
	"github.com/olivere/elastic"
	"strings"
)

type LogData struct {
	Topic string `json:"topic"`
	Data string `json:"data"`
}

//init es to receive data from kafka

var (
	client  *elastic.Client
	ch chan *LogData
)

func Init(addr string, chanSize, nums int) (err error){
	if !strings.HasPrefix(addr, "http://"){
		addr = "http://" + addr
	}
	client, err = elastic.NewClient(elastic.SetURL(addr))
	if err != nil {	
		return 
	}
	
	fmt.Println("connect to es success")

	ch = make(chan *LogData, chanSize)

	for i := 0; i < nums; i++ {
		go sendToEs()
	}

	go sendToEs()

	return 
}

func SendToESChan(msg *LogData)  {
	ch <- msg
}

func sendToEs() {
	for {
		select {
		case msg := <- ch:
		put1, err := client.Index().Index(msg.Topic).Type("_doc").BodyJson(msg).Do(context.Background())
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Printf("index %v -- type %v -- id %v", put1.Index, put1.Type, put1.Id)
		}
	}
}
