package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	ggmail "google.golang.org/api/gmail/v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func GetOauth2Config(filename string) (*oauth2.Config, error) {
	datas, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return google.ConfigFromJSON(datas, ggmail.GmailReadonlyScope)
}

func GetClient(config *oauth2.Config, tokenFile string) *http.Client {
	token, err := GetTokenFromFile(tokenFile)
	if err != nil {
		token, err = GetTokenFromWeb(config)
		if err != nil {
			panic(err)
		}
		SaveToken(tokenFile, token)
	}
	return config.Client(context.Background(), token)
}

func GetTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("复制下列地址到浏览器执行\n%s\n输入授权码:", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, err
	}
	return config.Exchange(context.TODO(), authCode)
}

func GetTokenFromFile(tokenFile string) (*oauth2.Token, error) {
	f, err := os.Open(tokenFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}


// Saves a token to a file path.
func SaveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}


