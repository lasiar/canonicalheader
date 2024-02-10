package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/lasiar/funcliner"
)

func main() {
	singlechecker.Main(funcliner.Analyzer)
}
