package main

//入口函数
import (
	"fmt"
	"gopkg.in/ini.v1"
	"logagent/conf"
	"logagent/etcd"
	"logagent/kafka"
	"logagent/taillog"
	"logagent/util"
	"sync"
	"time"
)

var (
	cfg = new(conf.AppConf)
)


func main() {
	//load ini file
	err := ini.MapTo(cfg, "./conf/config.ini")
	if err != nil {
		fmt.Println("load ini failed, err ", err)
		return
	}

	//初始化kafka连接
	err = kafka.Init([]string{cfg.KafkaConf.Address}, cfg.KafkaConf.ChanMaxSize)
	if err != nil {
		fmt.Println("init kafka failed err : ", err)
		return
	}
	fmt.Println("init kafka success")

	//init etcd
	err = etcd.Init(cfg.EtcdConf.Address, time.Duration(cfg.EtcdConf.Timeout) * time.Second)
	if err != nil {
		fmt.Println("init etcd failed err : ", err)
		return
	}
	fmt.Println("init etcd success")

	// to achieve every logAgent have their unique config depend on ip
	ip, err := util.GetOutboundIp()
	etcdConfKey := fmt.Sprintf(cfg.EtcdConf.Key, ip)

	//pull the config of log collection for etcd
	logEntryConf, err := etcd.GetConf(etcdConfKey)
	if err != nil {
		fmt.Printf("pull config failed err : %v \n", err)
		return
	}
	for _, val := range logEntryConf{
		fmt.Println(val)
	}

	//create a tailObj for a log
	taillog.Init(logEntryConf)

	newConfCh := taillog.NewConfChan()

	var wg sync.WaitGroup
	wg.Add(1)
	go etcd.WatchConf(cfg.EtcdConf.Key, newConfCh)
	wg.Wait()


}
