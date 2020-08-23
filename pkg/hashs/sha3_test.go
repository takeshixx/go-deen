package hashs

import (
	"bytes"
	"os"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
)

func TestNewPluginSHA3224(t *testing.T) {
	p := NewPluginSHA3224()
	r := bytes.NewReader(shaTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("TestNewPluginSHA3224 failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("4c6f98b0df46532ed5f0b30de780362d3bdadb08ae6ff9f0a0dc468b")); c != 0 {
		t.Errorf("TestNewPluginSHA3224 returned invalid data: %s", d)
	}
}

func TestNewPluginSHA3256(t *testing.T) {
	p := NewPluginSHA3256()
	r := bytes.NewReader(shaTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("TestNewPluginSHA3256 failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("281ead2ece78635fa0ec9cfe26b0df3342ccceffd630890597cd3d34bec9ad58")); c != 0 {
		t.Errorf("TestNewPluginSHA3256 returned invalid data: %s", d)
	}
}

func TestNewPluginSHA3384(t *testing.T) {
	p := NewPluginSHA3384()
	r := bytes.NewReader(shaTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("TestNewPluginSHA3384 failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("0f7aa1bd9ccaf36dbdbe965f0a98394ca3aad8a9ea6a6d13fb6783f45b73f5c73c333f2bf063e070ecc8fd47602b70a5")); c != 0 {
		t.Errorf("TestNewPluginSHA3384 returned invalid data: %s", d)
	}
}

func TestNewPluginSHA3512(t *testing.T) {
	p := NewPluginSHA3512()
	r := bytes.NewReader(shaTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("TestNewPluginSHA3512 failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("454c165b1f13db38ce7fa7dbad1d4c9f26b7cc085832b32c6fb1c965eb64894ac9133f9ad7691cc1da7ff95bbbd2259df898ac6686b692f066a62ec996b9b4ec")); c != 0 {
		t.Errorf("TestNewPluginSHA3512 returned invalid data: %s", d)
	}
}

func TestSHA3Usage(t *testing.T) {
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w

	p := NewPluginSHA3224()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(&p, flags, []string{})
	flags.Usage()

	p = NewPluginSHA3256()
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(&p, flags, []string{})
	flags.Usage()

	p = NewPluginSHA3384()
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(&p, flags, []string{})
	flags.Usage()

	p = NewPluginSHA3512()
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(&p, flags, []string{})
	flags.Usage()
}
