package cli

import (
	"errors"
	"flag"
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

	// Exclude Groups/SubGroups
	ExGroups sliceName

	// Exclude Project by Name
	ExProject sliceName

	// Credential
	Username string
	Password string
	Token    string
}

type sliceName []string

func (rr *sliceName) String() string {
	return strings.Join(*rr, ",")
}

func (rr *sliceName) Set(param string) error {
	*rr = append(*rr, param)
	return nil
}

var (
	ErrCredentialNotFound  = errors.New("Auth was not specified")
	ErrActionNotFound      = errors.New("Action is not recognized")
	ErrActionNotProvided   = errors.New("Action is not provided")
	ErrDirectoryNotValid   = errors.New("Directory is not valid")
	ErrGroupNameNotValid   = errors.New("Group name not valid")
	ErrProjectNameNotValid = errors.New("Project name not valid")
)

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

	action := os.Args[1]
	if action == "" {
		return ErrActionNotProvided
	}

	actions := map[string]struct{}{
		"clone-gitlab":  {},
		"update-gitlab": {},
		"update":        {},
		"version":       {},
		"usage":         {},
	}

	if _, ok := actions[action]; !ok {
		return ErrActionNotFound
	}
	c.Action = action

	subCommand := flag.NewFlagSet("action", flag.ExitOnError)

	subCommand.StringVar(&c.Baseurl, "u", "https://gitlab.com/", "Url gitlab repository")
	subCommand.StringVar(&c.Baseurl, "url", "https://gitlab.com/", "Url gitlab repository")

	subCommand.StringVar(&c.Username, "U", "", "Put username for git")
	subCommand.StringVar(&c.Username, "username", "", "Put username for git")

	subCommand.StringVar(&c.Password, "P", "", "Put password for git")
	subCommand.StringVar(&c.Password, "password", "", "Put password for git")

	subCommand.StringVar(&c.Token, "t", "", "Token for authentication")
	subCommand.StringVar(&c.Token, "token", "", "Token for authentication")

	subCommand.Var(&c.ExGroups, "eg", "Exclude group specified by group name")
	subCommand.Var(&c.ExProject, "ep", "Exclude project specified by project name")

	subCommand.StringVar(&c.Rootdir, "path", ".", "Set Working directory root path")
	subCommand.BoolVar(&c.Verbose, "verbose", false, "Activate verbose/debug print")
	subCommand.BoolVar(&c.Hardreset, "hard-reset", false, "Set false to use softreset or otherwise")

	_ = subCommand.Parse(os.Args[2:])
	return c.Validate()
}

// Validate mandatory input that has been set
// Action is a must: update, update-gitlab
// Credential is a must: using username & password or git token
// Root directory must valid or will using current dir
func (c *Cli) Validate() error {

	if c.Action == "version" || c.Action == "usage" {
		return nil
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
		Logs:       zLog,
		Exgroups:   ([]string)(c.ExGroups),
		Exprojects: ([]string)(c.ExProject),
	})

	return command, err
}
