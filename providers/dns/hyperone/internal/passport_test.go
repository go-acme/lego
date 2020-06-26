package internal

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLoadingValidPassport(t *testing.T) {
	passport, err := LoadPassportFile("fixtures/validPassport.json")
	if err != nil {
		t.Errorf("Loading passport file returned error: %+v", err)
	}

	expectedPassport := &Passport{SubjectID: "/iam/project/projectId/sa/serviceAccountId", CertificateID: "certificateID", Issuer: "https://api.hyperone.com/v2/iam/project/projectId/sa/serviceAccountId", PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\r\nlrMAsSjjkKiRxGdgR8p5kZJj0AFgdWYa3OT2snIXnN5+/p7j13PSkseUcrAFyokc\r\nV9pgeDfitAhb9lpdjxjjuxRcuQjBfmNVLPF9MFyNOvhrprGNukUh/12oSKO9dFEt\r\ns39F/2h6Ld5IQrGt3gZaBB1aGO+tw3ill1VBy2zGPIDeuSz6DS3GG/oQ2gLSSMP4\r\nOVfQ32Oajo496iHRkdIh/7Hho7BNzMYr1GxrYTcE9/Znr6xgeSdNT37CCeCH8cmP\r\naEAUgSMTeIMVSpILwkKeNvBURic1EWaqXRgPRIWK0vNyOCs/+jNoFISnV4pu1ROF\r\n92vayHDNSVw9wHcdSQ75XSE4Msawqv5U1iI7e2lD64uo1qhmJdrPcXDJQCiDbh+F\r\nhQhF+wAoLRvMNwwhg+LttL8vXqMDQl3olsWSvWPs6b/MZpB0qwd1bklzA6P+PeAU\r\nsfOvTqi9edIOfKqvXqTXEhBP8qC7ZtOKLGnryZb7W04SSVrNtuJUFRcLiqu+w/F/\r\nMSxGSGalYpzIZ1B5HLQqISgWMXdbt39uMeeooeZjkuI3VIllFjtybecjPR9ZYQPt\r\nFFEP1XqNXjLFmGh84TXtvGLWretWM1OZmN8UKKUeATqrr7zuh5AYGAIbXd8BvweL\r\nPigl9ei0hTculPqohvkoc5x1srPBvzHrirGlxOYjW3fc4kDgZpy+6ik5k5g7JWQD\r\nlbXCRz3HGazgUPeiwUr06a52vhgT7QuNIUZqdHb4IfCYs2pQTLHzQjAqvVk1mm2D\r\nkh4myIcTtf69BFcu/Wuptm3NaKd1nwk1squR6psvcTXOWII81pstnxNYkrokx4r2\r\n7YVllNruOD+cMDNZbIG2CwT6V9ukIS8tl9EJp8eyb0a1uAEc22BNOjYHPF50beWF\r\nukf3uc0SA+G3zhmXCM5sMf5OxVjKr5jgcir7kySY5KbmG71omYhczgr4H0qgxYo9\r\nZyj2wMKrTHLfFOpd4OOEun9Gi3srqlKZep7Hj7gNyUwZu1qiBvElmBVmp0HJxT0N\r\nmktuaVbaFgBsTS0/us1EqWvCA4REh1Ut/NoA9oG3JFt0lGDstTw1j+orDmIHOmSu\r\n7FKYzr0uCz14AkLMSOixdPD1F0YyED1NMVnRVXw77HiAFGmb0CDi2KEg70pEKpn3\r\nksa8oe0MQi6oEwlMsAxVTXOB1wblTBuSBeaECzTzWE+/DHF+QQfQi8kAjjSdmmMJ\r\nyN+shdBWHYRGYnxRkTatONhcDBIY7sZV7wolYHz/rf7dpYUZf37vdQnYV8FpO1um\r\nYa0GslyRJ5GqMBfDS1cQKne+FvVHxEE2YqEGBcOYhx/JI2soE8aA8W4XffN+DoEy\r\nZkinJ/+BOwJ/zUI9GZtwB4JXqbNEE+j7r7/fJO9KxfPp4MPK4YWu0H0EUWONpVwe\r\nTWtbRhQUCOe4PVSC/Vv1pstvMD/D+E/0L4GQNHxr+xyFxuvILty5lvFTxoAVYpqD\r\nu8gNhk3NWefTrlSkhY4N+tPP6o7E4t3y40nOA/d9qaqiid+lYcIDB0cJTpZvgeeQ\r\nijohxY3PHruU4vVZa37ITQnco9az6lsy18vbU0bOyK2fEZ2R9XVO8fH11jiV8oGH\r\n-----END RSA PRIVATE KEY-----\r\n", PublicKey: "-----BEGIN PUBLIC KEY-----\r\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxzYuc22QSst/dS7geYYK\r\n5l5kLxU0tayNdixkEQ17ix+CUcUbKIsnyftZxaCYT46rQtXgCaYRdJcbB3hmyrOa\r\nvkhTpX79xJZnQmfuamMbZBqitvscxW9zRR9tBUL6vdi/0rpoUwPMEh8+Bw7CgYR0\r\nFK0DhWYBNDfe9HKcyZEv3max8Cdq18htxjEsdYO0iwzhtKRXomBWTdhD5ykd/fAC\r\nVTr4+KEY+IeLvubHVmLUhbE5NgWXxrRpGasDqzKhCTmsa2Ysf712rl57SlH0Wz/M\r\nr3F7aM9YpErzeYLrl0GhQr9BVJxOvXcVd4kmY+XkiCcrkyS1cnghnllh+LCwQu1s\r\nYwIDAQAB\r\n-----END PUBLIC KEY-----\r\n"}

	passportValid := cmp.Equal(passport, expectedPassport)
	if !passportValid {
		t.Errorf("Passport struct is different than expected. Expected: %+v, got: %+v", passport, expectedPassport)
	}
}

func TestLoadingInvalidPassport(t *testing.T) {
	passport, err := LoadPassportFile("fixtures/invalidPassport.json")

	expectedError := "Error when validating passport file:Private key must be present"
	if err.Error() != expectedError {
		t.Errorf("Expected:%+v, got:%+v", expectedError, err)
	}

	if passport != nil {
		t.Errorf("Expected passport to be nil, got:%+v", passport)
	}
}

func TestProjectIDExtraction(t *testing.T) {
	expectedID := "5af0bbbcb7802508adxxxxxx"
	validPassportMock := Passport{SubjectID: fmt.Sprintf("/iam/project/%s/sa/5ef759c0ab0acab07xxxxxxx", expectedID)}
	extractedID, err := validPassportMock.ExtractProjectID()
	if err != nil {
		t.Errorf("Error when extracting project id:%+v", err)
	}
	if extractedID != expectedID {
		t.Errorf("Extracted projectID does not match expected. Got:%+s Expected:%+v", extractedID, expectedID)
	}

	invalidPassportMock := Passport{SubjectID: "ddddddd"}
	extractedID, err = invalidPassportMock.ExtractProjectID()
	if extractedID != "" {
		t.Error("Extracted projectID should be empty")
	}
	if err.Error() != "Error when extracting projectID" {
		t.Error("Extracting projectID did not throw error")
	}
}
