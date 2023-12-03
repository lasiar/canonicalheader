package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	analyzer "github.com/Lasiar/canonicalHeader"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
