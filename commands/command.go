package commands

import (
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/schollz/progressbar/v3"
	"go.uber.org/zap"
)

type Options struct {
	// Flag for activating logging while the program running
	Verbose bool

	// Define action to be executed in command
	Action string

	// Define root directory of the action
	Dir string

	// Exclude Group name, separated by commas
	Exgroup string

	// Exclude Project name, separated by commas
	Exproject string

	// a flag that used to make decision wether it's need to be hard reset
	// before the pull being executed
	Hardreset bool

	// Define authentication used for interacting with git repository
	Auth *Auth

	// Define base url (usefull for update-gitlab or clone project)
	Baseurl string

	// Set the zap logger
	Logs *zap.Logger
}

type Command struct {
	verbose   bool
	action    string
	dir       string
	exGroup   string
	exProject string
	auth      *Auth
	hardReset bool
	baseurl   string
	bar       *progressbar.ProgressBar

	// default logger for the package command (zap logger)
	log *zap.Logger
}

type Auth struct {
	// Username being used for authentication with git.
	// If this token was set, then this field will have default value "token" (acording to go-git docs to use like this)
	Username string

	// Password for authentication with git
	// If token was set, then it's gonna be put in here
	Password string
}

var (
	ErrCommandNotFound    = errors.New("Command not found/unrecognize")
	ErrCredentialNotFound = errors.New("Credential has not been set completely")
	ErrActionNotFound     = errors.New("Action not been initialize")
	ErrLogsNotDefined     = errors.New("Zap logger has not been defined")
	ErrDirNotExist        = errors.New("Directory not valid/exist")
)

// Generate new command struct for executing update
func New(opt *Options) (*Command, error) {

	if err := validate(opt); err != nil {
		return nil, err
	}

	c := Command{
		verbose:   opt.Verbose,
		action:    opt.Action,
		auth:      opt.Auth,
		dir:       opt.Dir,
		exGroup:   opt.Exgroup,
		exProject: opt.Exproject,
		hardReset: opt.Hardreset,
		baseurl:   opt.Baseurl,
		log:       opt.Logs,
	}

	if !opt.Verbose {
		c.bar = progressbar.Default(-1)
		c.bar.Describe("Start executing action ...")
	}

	return &c, nil
}

// Validate given options is enough to do the task
// Credential, action performed, directory and the logs
func validate(opt *Options) error {

	if opt.Action == "" {
		return ErrActionNotFound
	}

	if match, _ := regexp.MatchString(`[/\\]{2,}$`, opt.Dir); match {
		return ErrDirNotExist
	}

	if strings.HasSuffix(opt.Dir, "\\") || strings.HasSuffix(opt.Dir, "/") {
		opt.Dir = strings.TrimSuffix(opt.Dir, "/")
		opt.Dir = strings.TrimSuffix(opt.Dir, "\\")
	}

	if _, err := os.Stat(opt.Dir); err != nil {
		return ErrDirNotExist
	}

	if opt.Auth == nil ||
		(opt.Auth != nil && (opt.Auth.Username == "" || opt.Auth.Password == "")) {
		return ErrCredentialNotFound
	}

	if opt.Logs == nil {
		return ErrLogsNotDefined
	}

	return nil
}

func (c *Command) getCommandDispatcher() map[string]func() error {
	return map[string]func() error{
		"update": func() error {
			return c.updateGit()
		},
		"update-gitlab": func() error {
			return c.updateGitlab()
		},
		"clone-gitlab": func() error {
			return c.CloneGitlab()
		},
	}
}

// Execute action based on action key provided
func (c *Command) Execute() error {

	dispatcher := c.getCommandDispatcher()
	if dispatcher[c.action] == nil {
		return ErrCommandNotFound
	}

	err := dispatcher[c.action]()
	if err != nil {
		return err
	}

	return nil
}
