package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf // 这里已经包含了日志配置
	Mysql         MysqlConfig
	Redis         RedisConfig
	Snowflake     SnowflakeConfig
	ShortUrl      ShortUrlConfig
	// 删除 Log LogConfig 这一行
}

type MysqlConfig struct {
	DataSource string
}

type RedisConfig struct {
	Host string
	Type string
	Pass string
	DB   int
}

type SnowflakeConfig struct {
	MachineID int64
}

type ShortUrlConfig struct {
	Domain     string
	CodeLength int
	CacheTTL   int
}

// 删除整个 LogConfig 结构体
