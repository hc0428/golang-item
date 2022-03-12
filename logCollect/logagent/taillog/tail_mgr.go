package taillog

import (
	"fmt"
	"logagent/etcd"
	"time"
)

var tskMgr *tailMgr

type tailMgr struct {
	logEntry []*etcd.LogEntry
	tskMap map[string]*TailTask
	newConfChan chan []*etcd.LogEntry
}

func Init(logEntryConf []*etcd.LogEntry) {
	tskMgr = &tailMgr{
		logEntry: logEntryConf,
		tskMap: make(map[string]*TailTask, 16),
		newConfChan: make(chan []*etcd.LogEntry),
	}

	for _, LogConfig := range logEntryConf{
		tailTask := NewTailTask(LogConfig.Path, LogConfig.Topic)
		mk := fmt.Sprintf("%s_%s", LogConfig.Topic, LogConfig.Path)
		tskMgr.tskMap[mk] = tailTask
	}

	go tskMgr.run()
}

//listen own newConfChan, when new config is coming, deal with this changes

func (t *tailMgr) run() {
	for {
		select {
		case newConf := <-t.newConfChan	:
			//config add
			for _, conf := range newConf{
				mk := fmt.Sprintf("%s_%s", conf.Topic, conf.Path)
				_, ok := t.tskMap[mk]
				if ok {
					 continue
				}else {
					fmt.Println("config add : ", conf.Path)
					tailTask := NewTailTask(conf.Path, conf.Topic)
					t.tskMap[mk] = tailTask
				}
			}
			//config del
			for _, c1 := range t.logEntry {
				isDelete := true
				for _, c2 := range newConf{
					if c1.Path == c2.Path && c1.Topic == c2.Topic{
						isDelete = false
					}
				}
				if isDelete{
					mk := fmt.Sprintf("%s_%s", c1.Topic, c1.Path)
					tailTask := tskMgr.tskMap[mk]
					fmt.Println("cancel : ",c1.Path)
					tailTask.cancelFunc()
					delete(tskMgr.tskMap, mk)
				}
			}
			//config modify equal first add and next del
			fmt.Println("config has changed")
		default:
			time.Sleep(time.Second)
		}

	}

}

//expose newConfChanx

func NewConfChan() (newConf chan []*etcd.LogEntry) {
	return tskMgr.newConfChan
}

