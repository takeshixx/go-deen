package hashs

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

var bcryptTestData = []byte("deenpassword")

func TestPluginBcrypt(t *testing.T) {
	out := runHash(t, NewPluginBcrypt(), bcryptTestData)
	if err := bcrypt.CompareHashAndPassword(out, bcryptTestData); err != nil {
		t.Error("bcrypt returned a non-matching hash")
	}

	out = runHash(t, NewPluginBcrypt(), bcryptTestData, "-cost", "7")
	if err := bcrypt.CompareHashAndPassword(out, bcryptTestData); err != nil {
		t.Error("bcrypt with custom cost returned a non-matching hash")
	}
}
