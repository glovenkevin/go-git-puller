package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type Cli struct {
	Verbose   bool
	Rootdir   string
	Hardreset bool
	Action    string
	Baseurl   string
	Username  string
	Password  string
}

// Return new Cli struct for consuming
// parameter that needed to do actions
func New() *Cli {
	c := Cli{
		Verbose:   false,
		Hardreset: false,
		Rootdir:   ".",
	}

	return &c
}

// Define all args needed for being set
// And then parse it
func (c *Cli) Parse() error {

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage Command go-git-pull [-c/-action=action] [-option=option ...]")
		fmt.Fprintln(os.Stderr, "\tAction: update, *update-gitlab, *pull-gitlab")

		fmt.Fprintln(os.Stderr, "\nList option available: ")
		flag.CommandLine.PrintDefaults()
	}

	flag.StringVar(&c.Action, "c", "", "Set action")
	flag.StringVar(&c.Action, "action", "", "Set action")

	flag.StringVar(&c.Baseurl, "u", "https://gitlab.com/", "Url gitlab repository")
	flag.StringVar(&c.Baseurl, "url", "https://gitlab.com/", "Url gitlab repository")

	flag.StringVar(&c.Username, "U", "token", "Put username for git")
	flag.StringVar(&c.Username, "username", "token", "Put username for git")

	flag.StringVar(&c.Password, "P", "", "Put password for git")
	flag.StringVar(&c.Password, "password", "", "Put password for git")

	flag.StringVar(&c.Password, "t", "", "Token for authentication")
	flag.StringVar(&c.Password, "token", "", "Token for authentication")

	flag.StringVar(&c.Rootdir, "path", ".", "Set Working directory root path")
	flag.BoolVar(&c.Verbose, "verbose", false, "Activate verbose/debug print")
	flag.BoolVar(&c.Hardreset, "hard-reset", false, "Set false to use softreset or otherwise")

	flag.Parse()
	return c.validateInput()
}

var (
	ErrCredentialNotFound = errors.New("Username/Password/Token has not ben set")
	ErrActionNotFound     = errors.New("Action not defined")
)

// Validate mandatory input that has been set
func (c *Cli) validateInput() error {
	if c.Username == "" || c.Password == "" {
		return ErrCredentialNotFound
	}

	if _, err := os.Stat(c.Rootdir); err != nil {
		return err
	}

	if c.Action == "" {
		return ErrActionNotFound
	}

	return nil
}
