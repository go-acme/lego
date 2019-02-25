package envprovider

import (
	"fmt"
	"os"
)

//GetCompartmentID 環境変数からCompartmentIDを取得する
func GetCompartmentID() (value string, err error) {
	envKey := "OCI_COMPARTMENT_ID"
	var ok bool
	if value, ok = os.LookupEnv(envKey); !ok {
		err = fmt.Errorf("can not read CompartmentID from environment variable %s", envKey)
		return "", err
	}
	return value, nil
}
