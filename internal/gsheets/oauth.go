package gsheets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

func GetClient(config *oauth2.Config) *http.Client {
	tokFile := "../secrets/token.json"
	tok, err := TokenFromFile(tokFile)

	if err != nil {
		tok = GetTokenFromWeb(config)
		SaveToken(tokFile, tok)
	}

	client := config.Client(context.Background(), tok)
	client.Transport = &tokenSourceTransport{
		base:    config.TokenSource(context.Background(), tok),
		config:  config,
		tokFile: tokFile,
	}

	return client
}

type tokenSourceTransport struct {
	base    oauth2.TokenSource
	config  *oauth2.Config
	tokFile string
}

func (t *tokenSourceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.base.Token()
	if err != nil {
		if token == nil || err.Error() == "oauth2: cannot fetch token: 400 Bad Request: {\"error\":\"invalid_grant\"}" {
			token = GetTokenFromWeb(t.config)
			SaveToken(t.tokFile, token)
			t.base = t.config.TokenSource(context.Background(), token)
		} else {
			log.Fatalf("Unable to retrieve refreshed token: %v", err)
		}
	}
	return oauth2.NewClient(context.Background(), t.base).Transport.RoundTrip(req)
}

func GetTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func TokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func SaveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
