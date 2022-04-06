package commands

import (
	"testing"

	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
)

func TestCloneGitlab(t *testing.T) {
	zLog, _ := zap.NewDevelopment()
	n := &nodeGitlab{
		auth: &Auth{
			Username: "kevin.chandra",
			Password: "wings123",
		},
		log: zLog,
	}
	p := &gitlab.Project{
		Name:          "External",
		HTTPURLToRepo: "http://172.20.5.240/wingsdev/dependency/external.git",
	}

	t.Log("Start pulling project ...")
	n.cloneRepo("D:/temp/test/WingsDev/Dependency/External", p)
	t.Log("Done Pull Project")
}
