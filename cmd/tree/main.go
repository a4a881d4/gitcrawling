package main

import (
	"flag"
)

var (
	argReposDir = flag.String("r", ".", "The dir story Repos")
	argDir      = flag.String("t", "../.gitdb", "The dir story Trees")
	argDump     = flag.Bool("d", false, "dump Trees")
	argMode     = flag.String("m", "raw", "raw or flat")
	repoNum     = 0
	missNum     = 0
)

func main() {
	flag.Parse()
	switch *argMode {
	case "raw":
		Raw()
	case "db":
		ReadDB(flag.Arg(0))
	default:
		Flat()
	}
}
