package actions

import (
	"testing"

	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/suite/v3"
)

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	action, err := suite.NewActionWithFixtures(App(), packr.New("Test_ActionSuite", "../fixtures"))
	if err != nil {
		t.Fatal(err)
	}

	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}

func createFixture(as *ActionSuite, f interface{}) {
	err := as.DB.Create(f)
	if err != nil {
		as.T().Errorf("error creating %T fixture, %s", f, err)
		as.T().FailNow()
	}
}
