package main

import (
	"github.com/lyb88999/PortScan/internal/config"
	"log"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalln("failed to load the config")
	}
	_ = cfg
}
