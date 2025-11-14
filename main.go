package main

import (
	"flag"

	"github.com/VageLO/bsparse/parse"
)

var Path string

func main() {
    filePathFlag := flag.String("f", "", "Parse provided pdf")
    flag.Parse()

    if *filePathFlag == "" {
        panic("flag -f can't be empty")
    }

	t := bsparse.GetTransactions(*filePathFlag)
    bsparse.Csv(t)
}
