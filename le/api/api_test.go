package api_test

import (
	"testing"

	"github.com/xenolf/lego/le"
	"github.com/xenolf/lego/le/api"
)

func TestName(t *testing.T) {
	core, err := api.New(nil, "", "", "", nil)
	if err != nil {

	}

	_, err = core.Accounts.New(le.AccountMessage{})
	if err != nil {

	}
}
