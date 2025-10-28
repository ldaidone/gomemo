package main

import (
	"flag"
	"fmt"
	"strings"
)

func usage() {
	var help string
	help = `
Usage: from module root

make demo NAME="metrics|fibonacci|db_query|api"

or with go:
go cmd/examples/*.go -name <metrics|fibonacci|db_query|api>
`
	fmt.Printf("\n%s\n\n", help)
}

func main() {
	var name string

	flag.StringVar(&name, "name", "metrics", "Select the example to be executed: metrics, fibonacci, db_query, api. Default is metrics")
	flag.Parse()

	if name == "" {
		usage()
		return
	}

	switch strings.ToLower(name) {
	case "api":
		RunAPIMemo()
		break
	case "fibonacci":
		RunFibonacci()
		break
	case "db_query":
		RunDBQuery()
		break
	case "metrics":
		RunMetrics()
	default:
		usage()
	}
}
