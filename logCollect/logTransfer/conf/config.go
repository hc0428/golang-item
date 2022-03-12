package conf


type LogTransferCfg struct {
	KafkaCfg `ini:"kafka"`
	EsCfg    `ini:"es"`
}

type KafkaCfg struct {
	Address string `ini:"address"`
	Topic string `ini:"topic"`
}

type EsCfg struct {
	Address  string `ini:"address"`
	ChanSize int    `ini:"chan_size"`
	Nums int `ini:"nums"`
}
