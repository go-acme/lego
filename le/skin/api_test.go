package skin_test

import (
	"testing"

	"github.com/xenolf/lego/le"
	"github.com/xenolf/lego/le/skin"
)

func TestName(t *testing.T) {
	core, err := skin.New(nil, "", "", "", nil)
	if err != nil {

	}

	_, err = core.Accounts.New(le.AccountMessage{})
	if err != nil {

	}
}
