package main

import (
	"bufio"
	"io"
	"log"
	"net/mail"
	"os"
	"os/exec"
	"strings"
)

func main() {

	config, err := ParseConfig("/etc/mailsend.json")

	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)
	body, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	bodyStr := string(body)
	endline := "\r\n"
	endHeader := strings.Index(bodyStr, endline+endline)
	if endHeader == -1 {
		endline = "\n"
		endHeader = strings.Index(bodyStr, endline+endline)
	}
	if endHeader == -1 {
		log.Fatal("cannot read header:" + bodyStr)
	}
	header := bodyStr[0:endHeader]
	lines := strings.Split(header, endline)
	headerParameters := make(map[string]string)
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			name := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])
			headerParameters[name] = value
		}
	}

	whitelisted := false

	toList, err := mail.ParseAddressList(headerParameters["to"])
	if err != nil {
		log.Fatal(err)
	}

	for _, to := range toList {
		for _, allowDestination := range config.AllowTo {
			if allowDestination == to.Address {
				whitelisted = true
				break
			}
		}
		if whitelisted {
			break
		}
	}
	if !whitelisted {
		log.Fatalf("Destination (%s) is not allowed", headerParameters["to"])
	}

	cmd := exec.Command("/usr/sbin/sendmail", os.Args[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
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
	_, err = stdin.Write(body)
	if err != nil {
		log.Fatal(err)
	}
	stdin.Close()

	go func() {
		if _, err := io.Copy(os.Stdout, stdout); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		if _, err := io.Copy(os.Stderr, stderr); err != nil {
			log.Fatal(err)
		}
	}()
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(cmd.ProcessState.ExitCode())

}
