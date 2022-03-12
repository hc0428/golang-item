package main

import (
	"fmt"
	"gopkg.in/ini.v1"
	"logTransfer/conf"
	"logTransfer/es"
	"logTransfer/kafka"
)

//log transfer
//将日志数据从kafka取出来发往Es

func main() {
	// load ini config
	var cfg = new(conf.LogTransferCfg)
	err := ini.MapTo(cfg,"./conf/config.ini")
	if err != nil {
		fmt.Println("init config, err: ", err)
		return
	}
	fmt.Printf("cfg:%v\n", cfg)

	//init es
	err = es.Init(cfg.EsCfg.Address, cfg.EsCfg.ChanSize, cfg.EsCfg.Nums)
	if err != nil {
		fmt.Println("init Es client failed err : ", err)
	}

	//init kafka
	//connect kafka and every partitionConsumer send to Es by SendToES
	err = kafka.Init(cfg.KafkaCfg.Address, cfg.KafkaCfg.Topic)
	if err != nil {
		fmt.Println("init kafka err : ", err)
	}

	select {

	}
}
