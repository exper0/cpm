package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/exper0/user/test/cpm"
	"io/ioutil"
	"log"
)

func main() {
	configName := flag.String("config", "test.json", "Config file")
	flag.Parse()
	var c cpm.Config
	if err = json.Unmarshal(data, &c); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Hello, " + c.Processes[0].Name + ", " + c.Processes[0].Attributes.RunCmd)
}
