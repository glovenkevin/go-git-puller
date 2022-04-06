package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/glovenkevin/go-git-puller/commands"
	"go.uber.org/zap"
)

type Cli struct {
	// Action Needs parameter
	Verbose   bool
	Rootdir   string
	Hardreset bool
	Action    string
	Baseurl   string

	// Exclude Feature
	ExGroups  string
	ExProject string

	// Credential
	Username string
	Password string
	Token    string
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
		fmt.Fprintln(os.Stderr, "Usage Command go-git-pull [-c/-action action] [-option option ...]")
		fmt.Fprintln(os.Stderr, "\tAction: update, update-gitlab, clone-gitlab")

		fmt.Fprintln(os.Stderr, "\nList option available: ")
		flag.CommandLine.PrintDefaults()
	}

	flag.StringVar(&c.Action, "c", "", "Set action")
	flag.StringVar(&c.Action, "action", "", "Set action")

	flag.StringVar(&c.Baseurl, "u", "https://gitlab.com/", "Url gitlab repository")
	flag.StringVar(&c.Baseurl, "url", "https://gitlab.com/", "Url gitlab repository")

	flag.StringVar(&c.Username, "U", "", "Put username for git")
	flag.StringVar(&c.Username, "username", "", "Put username for git")

	flag.StringVar(&c.Password, "P", "", "Put password for git")
	flag.StringVar(&c.Password, "password", "", "Put password for git")

	flag.StringVar(&c.Token, "t", "", "Token for authentication")
	flag.StringVar(&c.Token, "token", "", "Token for authentication")

	flag.StringVar(&c.ExGroups, "eg", "", "Exclude group specified by group name. Put multiple groups separately with commas")
	flag.StringVar(&c.ExProject, "ep", "", "Exclude project specified by project name. Put multiple projects separately with commas")

	flag.StringVar(&c.Rootdir, "path", ".", "Set Working directory root path")
	flag.BoolVar(&c.Verbose, "verbose", false, "Activate verbose/debug print")
	flag.BoolVar(&c.Hardreset, "hard-reset", false, "Set false to use softreset or otherwise")

	flag.Parse()
	return c.Validate()
}

var (
	ErrCredentialNotFound  = errors.New("Username/Password/Token has not ben set")
	ErrActionNotFound      = errors.New("Action was not specified")
	ErrDirectoryNotValid   = errors.New("Directory is not valid")
	ErrGroupNameNotValid   = errors.New("Group name not valid")
	ErrProjectNameNotValid = errors.New("Project name not valid")
)

// Validate mandatory input that has been set
// Action is a must: update, update-gitlab
// Credential is a must: using username & password or git token
// Root directory must valid or will using current dir
func (c *Cli) Validate() error {

	if c.Action == "" {
		return ErrActionNotFound
	}

	if c.Token == "" && (c.Username == "" || c.Password == "") {
		return ErrCredentialNotFound
	}

	if c.Token != "" {
		c.Username = "token"
		c.Password = c.Token
	}

	if c.Rootdir == "" {
		c.Rootdir = "."
	}

	if match, _ := regexp.MatchString(`[/\\]{2,}$`, c.Rootdir); match {
		return ErrDirectoryNotValid
	}

	if strings.HasSuffix(c.Rootdir, "\\") || strings.HasSuffix(c.Rootdir, "/") {
		c.Rootdir = strings.TrimSuffix(c.Rootdir, "/")
		c.Rootdir = strings.TrimSuffix(c.Rootdir, "\\")
	}

	if _, err := os.Stat(c.Rootdir); err != nil {
		return err
	}

	regex := regexp.MustCompile("[^a-zA-Z0-9,_-]+")
	if c.ExGroups != "" && regex.MatchString(c.ExGroups) {
		return ErrGroupNameNotValid
	}

	if c.ExProject != "" && regex.MatchString(c.ExProject) {
		return ErrProjectNameNotValid
	}

	return nil
}

func (c *Cli) NewCommand(zLog *zap.Logger) (*commands.Command, error) {
	command, err := commands.New(&commands.Options{
		Verbose: c.Verbose,
		Action:  c.Action,
		Dir:     c.Rootdir,
		Baseurl: c.Baseurl,
		Auth: &commands.Auth{
			Username: c.Username,
			Password: c.Password,
		},
		Logs: zLog,
	})

	return command, err
}
