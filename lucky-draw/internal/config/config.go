package config

type Config struct {
	RedisAddr  string
	ServerAddr string
}

func New() *Config {
	return &Config{
		RedisAddr:  "localhost:6379",
		ServerAddr: ":9004",
	}
}