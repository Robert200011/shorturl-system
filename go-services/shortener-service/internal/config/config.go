package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Mysql     MysqlConfig
	Redis     RedisConfig
	Snowflake SnowflakeConfig
	ShortUrl  ShortUrlConfig
	Log       LogConfig
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

type LogConfig struct {
	ServiceName string
	Mode        string
	Level       string
}
