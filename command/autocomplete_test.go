package command

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/posener/complete"
)

func TestMetaCompletePredictWorkspaceName(t *testing.T) {
	// Create a temporary working directory that is empty
	td := tempDir(t)
	os.MkdirAll(td, 0755)
	defer os.RemoveAll(td)
	defer testChdir(t, td)()

	// make sure a vars file doesn't interfere
	err := ioutil.WriteFile(DefaultVarsFilename, nil, 0644)
	if err != nil {
		t.Fatal(err)
	}

	ui := new(cli.MockUi)
	meta := &Meta{Ui: ui}

	predictor := meta.completePredictWorkspaceName()

	got := predictor.Predict(complete.Args{
		Last: "",
	})
	want := []string{"default"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, want)
	}
}
