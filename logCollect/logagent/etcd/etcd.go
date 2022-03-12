package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

var (
	cli *clientv3.Client
)

type LogEntry struct {
	Path string `json:"path"`
	Topic string `json:"topic"`
}

func Init(addr string, timeOut time.Duration) (err error) {
	cli, err = clientv3.New(clientv3.Config{
		Endpoints: []string{addr},
		DialTimeout: timeOut,
	})
	if err != nil {
		fmt.Println("connect etcd failed, err : ", err)
		return
	}
	fmt.Println("connect etcd success ")

	return
}

//get config for etcd depend on key

func GetConf(key string)(logEntrys []*LogEntry, err error){
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, key)
	cancel()
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
		return
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s:%s\n", ev.Key, ev.Value)
		err = json.Unmarshal(ev.Value, &logEntrys)
		if err != nil {
			fmt.Printf("unmarshal etcd value failed err:%v\n", err)
			return
		}
	}
	return
}

// etcd watch

func WatchConf(key string, newConfCh chan<- []*LogEntry) {
	ch := cli.Watch(context.Background(), key)

	for wresp := range ch {
		for _, evt := range wresp.Events{
			fmt.Printf("Type:%v, key:%v, value:%v\n", evt.Type, string(evt.Kv.Key), string(evt.Kv.Value))
			var newConf []*LogEntry

			err := json.Unmarshal(evt.Kv.Value, &newConf)
			if err != nil {
				fmt.Printf("unmarshal failed, err:%v\n", err)
				continue
			}
			fmt.Println("new conf changes")
			newConfCh <- newConf
		}
	}
}
