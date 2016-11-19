package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
)

const clientSecretLocation = "/client_secret.json"
const credentialLocation = "/credential.json"

type GMail struct {
	ctx  context.Context
	srv  *gmail.Service
	user string
}

func NewGMail() *GMail {
	p := new(GMail)
	p.ctx = context.Background()
	p.user = "me"

	clientSecretFile, err := ioutil.ReadFile(clientSecretLocation)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(clientSecretFile, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := p.getClient(config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}

	p.srv = srv

	return p
}

func (g *GMail) getClient(config *oauth2.Config) *http.Client {
	tok, err := g.tokenFromFile(credentialLocation)
	if err != nil {
		tok = g.getTokenFromWeb(config)
		g.saveToken(credentialLocation, tok)
	}
	return config.Client(g.ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func (g *GMail) getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func (g *GMail) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func (g *GMail) saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func (g *GMail) getSteamGuardToken() string {
	steamGuardToken := ""

	messageList, err := g.srv.Users.Messages.List(g.user).Fields("messages").Q("from:noreply@steampowered.com").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages. %v", err)
	}
	if len(messageList.Messages) > 0 {
		for _, message := range messageList.Messages {
			messageFull, err := g.srv.Users.Messages.Get(g.user, message.Id).Do()
			if err != nil {
				log.Fatalf("Unable to retrieve full message. %v", err)
			}

			isAccessMessage := false
			for _, header := range messageFull.Payload.Headers {
				if header.Name == "Subject" && header.Value == "Your Steam account: Access from new computer" {
					isAccessMessage = true
					break
				}
			}

			if isAccessMessage {
				for _, part := range messageFull.Payload.Parts {
					if part.MimeType == "text/plain" {
						str, _ := base64.StdEncoding.DecodeString(part.Body.Data)

						steamGuardCodeRegExp, err := regexp.Compile(`:(?:\r?\n){2}\b(\w{5})(?:\r?\n)+`)
						if err != nil {
							log.Fatalf("RegExp Compile error. %v", err)
						}
						steamGuardCodeMatch := steamGuardCodeRegExp.FindStringSubmatch(string(str))
						if len(steamGuardCodeMatch) > 1 {
							steamGuardToken = steamGuardCodeMatch[1]
							break
						}
					}
				}

				if steamGuardToken != "" {
					err := g.srv.Users.Messages.Delete(g.user, message.Id).Do()
					if err != nil {
						log.Fatalf("Unable to delete message. %v", err)
					}
				}

				break
			}
		}
	}

	return steamGuardToken
}
