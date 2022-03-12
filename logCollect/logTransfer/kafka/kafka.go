package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"logTransfer/es"
	"sync"
)



func Init(addr string, topic string) error {
	consumer, err := sarama.NewConsumer([]string{addr}, nil)
	if err != nil {
		fmt.Println("fail to start consumer err: ", err)
		return err
	}
	fmt.Println(topic)
	partitionList, err := consumer.Partitions(topic)
	if err != nil {
		fmt.Println("fail to get list of partition err : ", err)
		return err
	}
	fmt.Println("partition list ", partitionList)

	for partition := range partitionList {
		//for every partition to create a consumer for a partition
		pc, err := consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			fmt.Printf("failed to start consumer partition %d err %v ", partition, err)
			return err
		}

		defer pc.AsyncClose()

		var wg sync.WaitGroup
		wg.Add(1)

		go func(pc sarama.PartitionConsumer) {
			fmt.Println("enter goroutine")
			defer wg.Done()
			for message := range pc.Messages() {
				fmt.Printf("partition:%v offset:%v key:%v value:%v\n", message.Partition, message.Offset, message.Key, string(message.Value))

				ld := es.LogData{
					Topic: topic,
					Data: string(message.Value),
				}

				es.SendToESChan(&ld)
			}
			fmt.Println("exit goroutine")
		}(pc)

		wg.Wait()
	}

	return err
}
