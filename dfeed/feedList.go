package dfeed

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
  Hook string
  Frequency int8
  Feeds []string
}

func SetConfig() Config{
	c, err := os.ReadFile("config.yml")
	if err != nil {
		log.Fatal(err)
	}

  t := Config{}
  err = yaml.Unmarshal([]byte(c), &t)
  if err != nil {
    log.Fatalf("error: %v", err)
  }

  return t
}
