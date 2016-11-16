package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
)

func main() {
	cmd := exec.Command("/opt/steamcmd/steamcmd.sh", strings.Join(os.Args[1:], " "))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	go io.Copy(os.Stderr, stderr)

	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanBytes)
		var line string
		line = ""
		for scanner.Scan() {
			text := scanner.Text()
			fmt.Print(text)

			line += text
			if text == "\n" {
				line = ""
				continue
			}

			switch line {
			case "Steam Guard code:":
				// DO SOMETHING
				break
			}
		}
	}()

	defer cmd.Wait()

}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
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

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gmail-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
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
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getSteamGuardToken() {
	var steamGuardCode string
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/gmail-go-quickstart.json
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}

	user := "me"
	messageList, err := srv.Users.Messages.List(user).Fields("messages").Q("from:noreply@steampowered.com is:unread").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages. %v", err)
	}
	if len(messageList.Messages) > 0 {
		for _, l := range messageList.Messages {
			message, err := srv.Users.Messages.Get(user, l.Id).Do()
			if err != nil {
				log.Fatalf("Unable to retrieve messages. %v", err)
			}
			// fmt.Printf("- %v\n", message.)
			var isAccessMessage bool
			isAccessMessage = false
			for _, header := range message.Payload.Headers {
				if header.Name == "Subject" && header.Value == "Your Steam account: Access from new computer" {
					isAccessMessage = true
					break
				}
			}

			if isAccessMessage {
				for _, part := range message.Payload.Parts {
					if part.MimeType == "text/plain" {
						str, _ := base64.StdEncoding.DecodeString(part.Body.Data)

						steamGuardCodeRegExp, err := regexp.Compile(`:(?:\r?\n){2}\b(\w{5})(?:\r?\n)+`)
						if err != nil {
							log.Fatalf("RegExp Compile error. %v", err)
						}
						steamGuardCodeMatch := steamGuardCodeRegExp.FindStringSubmatch(string(str))
						if len(steamGuardCodeMatch) > 1 {
							steamGuardCode = steamGuardCodeMatch[1]
							break
						}
					}
				}

				fmt.Printf("%s", steamGuardCode)

				break
			}
		}
	} else {
		fmt.Print("No messages found.")
	}
}
