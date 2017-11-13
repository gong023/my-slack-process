package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/Zauberstuhl/go-coinbase"
	"github.com/gong023/my-slack-process/slack"
)

var c coinbase.APIClient

func main() {
	cKey := flag.String("key", "", "coinbase api key")
	cSec := flag.String("sec", "", "coinbase api secret")
	flag.Parse()
	if *cKey == "" || *cSec == "" {
		log.Fatal("missing parameter")
	}
	c = coinbase.APIClient{
		Key:    *cKey,
		Secret: *cSec,
	}

	acc, err := c.Accounts()
	if err != nil {
		log.Fatal(err)
	}

	for _, acc := range acc.Data {
		if acc.Currency == "USD" {
			continue
		}

		exchange, err := getExchange(acc.Currency)
		if err != nil {
			log.Fatal(err)
		}

		sum, err := getTransactionSum(acc.Id)
		if err != nil {
			log.Fatal(err)
		}

		nBal := acc.Native_balance.Amount
		cBal := acc.Balance.Amount
		attachment := slack.Attachment{
			Title: acc.Currency,
			Fields: []slack.Field{
				{
					Title: "Current",
					Value: fmt.Sprintf("%s X %0.2f = %0.2f USD", exchange, cBal, nBal),
					Short: true,
				},
				{
					Title: "Profit",
					Value: fmt.Sprintf("%0.2f / %0.2f = %0.2f%%", nBal, sum, nBal/sum*100),
					Short: true,
				},
			},
		}
		b, err := json.Marshal(attachment)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}

}

func getExchange(currency string) (exchange string, err error) {
	balance, err := c.GetExchangeRates(currency)
	if err != nil {
		return
	}

	return balance.Data.Rates["USD"].String(), nil
}

func getTransactionSum(accountID string) (sum float64, err error) {
	transactions := coinbase.APITransactions{}
	if err = pageTran(accountID, &transactions); err != nil {
		return
	}

	for _, data := range transactions.Data {
		sum += data.Native_amount.Amount
	}
	return
}

func pageTran(accountID string, transactions *coinbase.APITransactions) (err error) {
	if transactions.Pagination.Next_uri != "" {
		trans := coinbase.APITransactions{}
		err = c.Fetch("GET", transactions.Pagination.Next_uri, nil, &trans)
		for _, data := range trans.Data {
			transactions.Data = append(transactions.Data, data)
		}
		transactions.Pagination = trans.Pagination
		if trans.Pagination.Next_uri != "" {
			err = pageTran(accountID, transactions)
		}
		return
	}
	if len(transactions.Data) != 0 {
		return
	}
	trans, err := c.GetTransactions(accountID)
	if err != nil {
		return
	}
	for _, data := range trans.Data {
		transactions.Data = append(transactions.Data, data)
	}
	pageTran(accountID, transactions)
	return
}
