package cli

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Create folder for validate file testing
	folder := "test"
	err := os.Mkdir(folder, os.ModeDir)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	code := m.Run()

	// Remove folder after test done
	os.Remove(folder)
	// End test
	os.Exit(code)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		Name     string
		Param    Cli
		Expected error
	}{
		{
			Name:     "No Param Given",
			Param:    Cli{},
			Expected: ErrActionNotFound,
		},
		{
			Name: "Credential Test 1",
			Param: Cli{
				Action:   "update",
				Username: "asdf",
				Password: "",
			},
			Expected: ErrCredentialNotFound,
		},
		{
			Name: "Credential Test 2",
			Param: Cli{
				Action:   "update",
				Username: "",
				Password: "asdf",
			},
			Expected: ErrCredentialNotFound,
		},
		{
			Name: "Credential Test 3",
			Param: Cli{
				Action: "update",
				Token:  "token",
			},
			Expected: nil,
		},
		{
			Name: "Credential Test 4",
			Param: Cli{
				Action:   "update",
				Username: "",
				Password: "asdf",
				Token:    "token",
			},
			Expected: nil,
		},
		{
			Name: "Credential Test 5",
			Param: Cli{
				Action:   "update",
				Username: "user",
				Password: "pass",
				Token:    "token",
			},
			Expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := test.Param.Validate()
			require.Equal(t, test.Expected, err)
		})
	}
}

func TestValidateDirExist(t *testing.T) {
	cli := Cli{
		Action:  "update",
		Token:   "token",
		Rootdir: "test",
	}
	err := cli.Validate()
	require.Nil(t, err)
}

func TestValidateDirNotExist(t *testing.T) {
	cli := Cli{
		Action:  "update",
		Token:   "token",
		Rootdir: "test2",
	}
	err := cli.Validate()
	require.NotNil(t, err)
}
