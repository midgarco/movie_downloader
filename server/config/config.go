package config

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Configuration ...
type Configuration struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func saveConfig(c Configuration, filename string) error {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, bytes, 0644)
}

// LoadConfig will attempt to load the yaml configuration file. If
// the file doesn't exist it will prompt the user to create one.
func LoadConfig(filename string) (*Configuration, error) {
	log.Printf("Loading config file from '%s'", filename)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return createConfig(filename)
	}

	var c Configuration
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func createConfig(filename string) (*Configuration, error) {
	// Create directory if needed.
	basepath := path.Dir(filename)
	if os.MkdirAll(basepath, 0755) != nil {
		log.Panic("Unable to create directory for configuration file")
	}

	username, password := getCredentials()

	var c Configuration
	c.Username = username
	c.Password = password

	err := saveConfig(c, filename)
	if err != nil {
		log.Panic(err)
	}

	return &c, nil
}

func getCredentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		log.Errorf("could not read username: %v\n", err)
		return "", ""
	}

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Errorf("could not read password: %v\n", err)
		return "", ""
	}
	password := string(bytePassword)

	fmt.Println("\n")

	return strings.TrimSpace(username), strings.TrimSpace(password)
}
