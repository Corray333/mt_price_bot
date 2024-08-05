package gsheets

import (
	"errors"
	"fmt"
	"os"

	"github.com/Corray333/mt_price_bot/internal/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

func UpdateMessages() error {

	b, err := os.ReadFile("../secrets/credentials.json")
	if err != nil {
		return err
	}

	config, err := google.ConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return err
	}
	client := GetClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		return err
	}

	readRange := "Messages!B2:B15"
	resp, err := srv.Spreadsheets.Values.Get(os.Getenv("SHEET_ID"), readRange).Do()
	if err != nil {
		return err
	}

	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		return errors.New("no data found")
	}

	msgs := []string{}
	for _, val := range resp.Values {
		msgs = append(msgs, val[0].(string))
	}

	m := make(map[int]string)
	for i := range msgs {
		m[i] = msgs[i]
	}

	storage.Messages = m

	readRange = "Messages!D2:D20"
	resp, err = srv.Spreadsheets.Values.Get(os.Getenv("SHEET_ID"), readRange).Do()
	if err != nil {
		return err
	}

	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		return errors.New("no data found")
	}

	admins := []string{}
	for _, val := range resp.Values {
		if val[0] == nil {
			break
		}
		admins = append(admins, val[0].(string))
	}

	storage.Admins = admins

	fmt.Println(storage.Admins)

	return nil
}
