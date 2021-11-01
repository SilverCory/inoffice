package office

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Env struct {
	SlackSigningSecret string `env:"SLACK_SIGNING_SECRET"`
	SlackBotToken      string `env:"SLACK_BOT_TOKEN"`
	SlackMainChannel   string `env:"SLACK_MAIN_CHANNEL"`

	MySQLDSN string `env:"MYSQL_DSN"`
}

func GetEnv() Env {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	cfg := Env{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	return cfg
}
