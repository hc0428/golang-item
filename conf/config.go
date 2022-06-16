package conf

type AppConf struct {
	MysqlConf `ini:"mysql"`
}

type MysqlConf struct {
	Host     string `ini:"host"`
	Port     int    `ini:"port"`
	Username string `ini:"username"`
	Password string `ini:"password"`
	Database string `ini:"database"`
}
