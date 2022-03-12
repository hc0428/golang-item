package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"time"
)

//向kafka写日志的模块

type logData struct {
	topic string
	data string
}

var (
	client sarama.SyncProducer // 声明一个全局的连接kafka的生产者client
	logDataChan chan *logData
)

func Init(addr []string, maxSize int) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	config.Producer.Return.Successes = true

	client, err = sarama.NewSyncProducer(addr, config)
	if err != nil {
		fmt.Println("producer closed err: ", err)
		return
	}

	logDataChan = make(chan *logData, maxSize)
	go SendToKafka()

	return
}

func SendToKafka() {
	// create a msg struct
	for {
		select {
		case ld := <- logDataChan:
			msg := &sarama.ProducerMessage{}
			msg.Topic = ld.topic
			msg.Value = sarama.StringEncoder(ld.data)
			// send to kafka
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				fmt.Println("seng msg err : ", err)
				return
			}

			fmt.Printf("pid : %v offset : %v \n", pid, offset)
			fmt.Println("发送成功")
		default:
			time.Sleep(time.Millisecond * 50)
		}
	}

}

func SendToChan(topic, data string)  {
	msg := &logData{
		topic: topic,
		data: data,
	}
	logDataChan <- msg
}
