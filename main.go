package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/dslipak/pdf"
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

    Path = *filePathFlag

	r, err := pdf.Open(Path)
	check(err)

	var transactions TransactionSlice

	for pageNum := 1; pageNum <= r.NumPage(); pageNum++ {

		page := r.Page(pageNum)

		var sentence pdf.Text
		var temp pdf.Text

		var transaction Transaction

		var tranID int

		var temp_str string

		texts := page.Content().Text
		for i, text := range texts {
			if text.Y == sentence.Y {
				sentence.S = sentence.S + text.S
			} else if sentence.X == text.X {
				temp = text
				temp.S = sentence.S + " " + temp.S
				sentence = temp
			} else {
				sentence = text
			}

			if i+1 == len(texts) {
				temp_str += parse(texts[tranID:len(texts)])
				extractRegex(temp_str, &transaction)
				transactions = append(transactions, transaction)
			}
			if !TransactionNum.MatchString(sentence.S) {
				continue
			}

			if tranID != 0 {
				temp_str += parse(texts[tranID : i-(len(sentence.S)-1)])
				tranID = i + 1
				extractRegex(temp_str, &transaction)
				transactions = append(transactions, transaction)
				temp_str = ""
				transaction = Transaction{}
			}
			transaction.id = TransactionNum.FindString(sentence.S)
			tranID = i + 1
			sentence.S = ""
		}
	}
    transactions.sort()
	csv(transactions)
}

func parse(array []pdf.Text) string {
	var str string
	var sentence pdf.Text
	var temp pdf.Text

	for _, text := range array {
		if text.Y == sentence.Y {
			sentence.S = text.S
		} else if sentence.X == text.X {
			temp = text
			temp.S = " " + temp.S
			sentence = temp
		} else {
			sentence = text
			sentence.S = " " + sentence.S
		}
		str += strings.ReplaceAll(sentence.S, `"`, "")
	}
	return str
}

func extractRegex(str string, transaction *Transaction) {
	var priceIndex, dateIndex, timeIndex, typeIndex []int

	if TransactionType.MatchString(str) {
		typeIndex = TransactionType.FindStringIndex(str)
		find := str[typeIndex[0]:(typeIndex[1] - 2)]
		find = strings.Trim(find, " ")
		transaction.typeof = find
	}
	if PriceWithAcronym.MatchString(str) {
		priceIndex = PriceWithAcronym.FindStringIndex(str)
		price_acronym := str[priceIndex[0]:priceIndex[1]]

		transaction.price = Price.FindString(price_acronym)
		transaction.acronym = Acronym.FindString(price_acronym)
		transaction.description = str[priceIndex[1]:len(str)]
	}
	if len(typeIndex) != 0 && len(priceIndex) != 0 {
		find := str[typeIndex[1]-2 : priceIndex[0]]
		find = strings.Trim(find, " ")
		transaction.status = find
	}
	str = strings.ReplaceAll(str, " ", "")

	if Date.MatchString(str) {
		dateIndex = Date.FindStringIndex(str)
		transaction.date = str[dateIndex[0]:dateIndex[1]]
	}
	if Time.MatchString(str) {
		timeIndex = Time.FindStringIndex(str)
		transaction.time = str[timeIndex[0]:timeIndex[1]]
	}

}

func changeExtToCSV(path string) string {
    return filepath.Join(filepath.Dir(path), strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) + ".csv")
}

func csv(t []Transaction) {
    Path = changeExtToCSV(Path)
	file, err := os.Create(Path)
	check(err)

	debit := []string{"Безналичная операция", "Отправление средств", "Банкомат"}
	str := "ID;DATE;TIME;TYPE_OF_TRANSACTION;STATUS;PRICE;ACRONYM;DESCRIPTION\n"
	for _, transaction := range t {
		if contains(debit, transaction.typeof) {
			transaction.price = "-" + transaction.price
		}
		temp := []string{transaction.id, transaction.date, transaction.time,
			transaction.typeof, transaction.status, transaction.price, transaction.acronym, transaction.description}
		str += strings.Join(temp, ";")
		str += "\n"
	}
	file.WriteString(str)
	file.Close()
    fmt.Println(Path)
}

func (t TransactionSlice) sort() {
    slices.SortFunc(t, func(a, b Transaction) int {
        layout := "2006-01-02 15:04:05"

        dateA, err := time.Parse(layout, a.date + " " + a.time)
        check(err)

        dateB, err := time.Parse(layout, b.date + " " + b.time)
        check(err)

        if dateA.Before(dateB) {
            return 1
        } else if dateA.After(dateB) {
            return -1
        } else if dateA.Equal(dateB) {
            return 0
        } else {
            return 0
        }
    })
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
