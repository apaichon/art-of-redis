package config

type Config struct {
    RedisAddr     string
    ServerAddress string
    LeaderboardKey string
}

func New() *Config {
    return &Config{
        RedisAddr:      "localhost:6379",
        ServerAddress:  ":9002",
        LeaderboardKey: "leaderboard",
    }
}