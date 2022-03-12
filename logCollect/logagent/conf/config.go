package conf

type AppConf struct {
	KafkaConf `ini:"kafka"`
	EtcdConf  `ini:"etcd"`
}

type KafkaConf struct {
	Address string `ini:"address"`
	ChanMaxSize int `ini:"chan_max_size"`
}

type EtcdConf struct {
	Address string `ini:"address"`
	Timeout int `int:"timeout"`
	Key string `ini:"key"`
}

type TailLogConf struct {
	Path string `ini:"path"`
}
