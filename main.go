package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	log.SetPrefix("csm: ")
	configFile := flag.String("config", "test.json", "Config file")
	flag.Parse()
	if config, err := NewConfig(*configFile); err == nil {
		ctr := NewServiceManager(config)
		if err := ctr.StartAll(); err != nil {
			log.Fatal(err)
		}
		os.Exit(ctr.MainLoop())
	} else {
		log.Fatal(err)
	}
}
