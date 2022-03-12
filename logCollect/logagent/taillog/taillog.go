package taillog

import (
	"context"
	"fmt"
	"github.com/hpcloud/tail"
	"logagent/kafka"
)

//读取日志的模块

type TailTask struct {
	path string
	topic string
	instance *tail.Tail
	ctx context.Context
	cancelFunc context.CancelFunc
}

func NewTailTask(path, topic string) (tailObj *TailTask){
	ctx, cancel := context.WithCancel(context.Background())
	tailObj = &TailTask{
		path: path,
		topic: topic,
		ctx: ctx,
		cancelFunc: cancel,
	}
	tailObj.init()

	return
}

func (t *TailTask) init() () {
	config := tail.Config{
		ReOpen: true, //重新打开
		Follow: true, //是否跟随
		Location: &tail.SeekInfo{Offset: 0, Whence: 2}, //从文件那里开始读
		MustExist: false,//文件不存在不报错
		Poll: true,//
	}

	var err error
	t.instance, err = tail.TailFile(t.path, config)
	if err != nil {
		fmt.Println("tailfile error ", err)
		return
	}

	go t.run()
}

func (t *TailTask) run() {
	for {
		select {
		case line := <-t.instance.Lines:
			kafka.SendToChan(t.topic, line.Text)
		case <- t.ctx.Done():
			fmt.Println("cancel success ")
			return
		}
	}
}
