package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/kr/pty"
)

func main() {
	gmail := NewGMail()

	cmd := exec.Command("/opt/steamcmd/steamcmd.sh", os.Args[1:]...)
	application, err := pty.Start(cmd)
	if err != nil {
		panic(err)
	}

	go func() {
		scanner := bufio.NewScanner(application)
		scanner.Split(bufio.ScanBytes)
		line := ""
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
				tickerGuard := time.NewTicker(time.Second * 5)

				go func() {
					go func() {
						time.Sleep(time.Second * 60)
						tickerGuard.Stop()
					}()

					for range tickerGuard.C {
						token := gmail.getSteamGuardToken()

						if token != "" {
							tickerGuard.Stop()

							application.Write([]byte(token + "\n"))
						}
					}
				}()

				break
			}
		}
	}()

	defer cmd.Wait()

}
