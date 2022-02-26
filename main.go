package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fa1se/dlc-parser/geosite"
)

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	sites := geosite.ParseCollection(data)
	if sites == nil {
		panic("malformed geodata")
	}
	for _, expr := range os.Args[1:] {
		selected := sites.Select(expr)
		for _, record := range selected {
			if record.Type == geosite.RECORD_KEYWORD || record.Type == geosite.RECORD_REGEXP {
				continue
			}
			fmt.Println(record.Value)
		}
	}
}
