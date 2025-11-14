package main

import (
	"flag"
	"regexp"

	"github.com/VageLO/bsparse"
)

var (
	PriceWithAcronym = regexp.MustCompile(`(\d+\.\d{2}\s*[A-Z]{3})`)
	Price            = regexp.MustCompile(`\d+\.\d{2}`)
	Acronym          = regexp.MustCompile(`[A-Z]{3}`)
	Date             = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	Time             = regexp.MustCompile(`\d{2}:\d{2}:\d{2}`)
	TransactionNum   = regexp.MustCompile(`\d{16}`)
	TransactionType  = regexp.MustCompile(`[А-Я][^А-Я]*[А-Я]`)
)

type Transaction struct {
	id          string
	date        string
	time        string
	typeof      string
	status      string
	price       string
	acronym     string
	description string
}

type TransactionSlice []Transaction

var Path string

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
    filePathFlag := flag.String("f", "", "Parse provided pdf")
    flag.Parse()

    if *filePathFlag == "" {
        panic("flag -f can't be empty")
    }

	t := bsparse.GetTransactions(*filePathFlag)
    bsparse.Csv(t)
}
