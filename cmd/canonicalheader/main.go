package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	analyzer "github.com/lasiar/canonicalHeader"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
