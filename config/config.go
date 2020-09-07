package config

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/apex/log"
	"golang.org/x/crypto/ssh/terminal"
)

func Create(filename string) error {
	// Create directory if needed.
	basepath := path.Dir(filename)
	if err := os.MkdirAll(basepath, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filename, []byte{}, 0644); err != nil {
		return err
	}

	return nil
}

func GetCredentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		log.WithError(err).Error("could not read username")
		return "", ""
	}

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Print("\n")
	if err != nil {
		log.WithError(err).Error("could not read password")
		return "", ""
	}
	password := string(bytePassword)

	return strings.TrimSpace(username), strings.TrimSpace(password)
}

func GetDownloadPath(downloadPath string) string {
	reader := bufio.NewReader(os.Stdin)

	defaultPath := os.Getenv("HOME") + "/Movies/"

	if downloadPath == "" {
		downloadPath = defaultPath
	}

	fmt.Printf("Enter Download Path (%s):", downloadPath)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.WithError(err).Error("could not read download path")
		return ""
	}
	input = strings.TrimSpace(input)

	if input == "" {
		input = defaultPath
	}

	return input
}

func GetMediaPath(mediaPath string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter Media Path (%s):", mediaPath)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.WithError(err).Error("could not read media path")
		return ""
	}
	input = strings.TrimSpace(input)

	return input
}
