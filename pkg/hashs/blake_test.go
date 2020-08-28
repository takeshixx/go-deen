package hashs

import (
	"bytes"
	"os"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
)

var blakeTestData = []byte("deenblaketest")
var blakeTestKey = "testkey123"
var blakeTestKey32 = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA__"

func TestPluginBLAKE2sProcessSteamFunc(t *testing.T) {
	p := NewPluginBLAKE2s()
	r := bytes.NewReader(blakeTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("BLAKE2sProcessStreamFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("4088080c149a6165b9a086ef4aaeb13df5fc7ffb83d5731ed9692320b5634c50")); c != 0 {
		t.Errorf("BLAKE2sProcessSteamFunc returned invalid data: %s", d)
	}
}

func TestPluginBLAKE2sProcessStreamWithCliFlagsFunc(t *testing.T) {
	p := NewPluginBLAKE2s()
	r := bytes.NewReader(blakeTestData)
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{"-key", blakeTestKey})
	d, e := p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("BLAKE2sProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("183f16fed32775091e9f63171cbdca57e6c0642dffa08c7a25ea58aa7b50f2d5")); c != 0 {
		t.Errorf("BLAKE2sProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE2s()
	r = bytes.NewReader(blakeTestData)
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	d, e = p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("BLAKE2sProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("4088080c149a6165b9a086ef4aaeb13df5fc7ffb83d5731ed9692320b5634c50")); c != 0 {
		t.Errorf("BLAKE2sProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE2s()
	r = bytes.NewReader(blakeTestData)
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{"-len", "16", "-key", blakeTestKey})
	d, e = p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("BLAKE2sProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("b789903bd37697727692abb9f0494bad")); c != 0 {
		t.Errorf("BLAKE2sProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}
}

func TestPluginBLAKE2bProcessStreamFunc(t *testing.T) {
	p := NewPluginBLAKE2b()
	r := bytes.NewReader(blakeTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("BLAKE2bProcessStreamFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("e3e8bca1c407f1ce36642d64c334bbc572f7ad06e00425d2abc567e094e9e82862b3d8f200647273ec4f1d36cc5b7371b6a4cf7ea6725529ce71ea9c68eeb66c")); c != 0 {
		t.Errorf("BLAKE2bProcessSteamFunc returned invalid data: %s", d)
	}
}

func TestPluginBLAKE2bProcessStreamWithCliFlagsFunc(t *testing.T) {
	p := NewPluginBLAKE2b()
	r := bytes.NewReader(blakeTestData)
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{"-key", blakeTestKey, "-len", "64"})
	d, e := p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("BLAKE2bProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("4e8c474aa515d314feb9cb0893e2bddaad49f007fbd1f0538776f2c11d9c9d04732b70a024642400b14707276928c94429b109424245156e438503aa312036d9")); c != 0 {
		t.Errorf("BLAKE2bProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE2b()
	r = bytes.NewReader(blakeTestData)
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{"-len", "64"})
	d, e = p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("BLAKE2bProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("e3e8bca1c407f1ce36642d64c334bbc572f7ad06e00425d2abc567e094e9e82862b3d8f200647273ec4f1d36cc5b7371b6a4cf7ea6725529ce71ea9c68eeb66c")); c != 0 {
		t.Errorf("BLAKE2bProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}
}

func TestPluginBLAKE2xProcessStreamFunc(t *testing.T) {
	p := NewPluginBLAKE2x()
	r := bytes.NewReader(blakeTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("TestPluginBLAKE2xProcessStreamFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("0924e4d71784282e91639a595475a0290a9c2caee4a03978199b4d2f7bcf8d83")); c != 0 {
		t.Errorf("BLAKE2bProcessSteamFunc returned invalid data: %s", d)
	}
}

func TestPluginBLAKE2xProcessStreamWithCliFlagsFunc(t *testing.T) {
	p := NewPluginBLAKE2x()
	r := bytes.NewReader(blakeTestData)
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{"-key", blakeTestKey, "-len", "64"})
	d, e := p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("TestPluginBLAKE2xProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("94e30c637ecf91f9873924d74e667e56099045b8e3cffaa3ad9d415b163af0ad6c0c1da67732b0d2497f152f197635d2ac76cead6c7a48fcb8c0b11ca3a726f0")); c != 0 {
		t.Errorf("TestPluginBLAKE2xProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE2x()
	r = bytes.NewReader(blakeTestData)
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{"-len", "256"})
	d, e = p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("TestPluginBLAKE2xProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("fc229cfc05261f697e49c8ae9db9a2c434c759636d77a87b8657ab40389fa1446a9fe4bd14d9b62122799ef8217b054df95a83e110c967c2bbfdeed6a41264efd1034c3279a9c3c2f9a17650d40e47a93eac18b8b59f5b28dfc0d948e178e033ff3057b90b5742a6c44eb498ac10756e4cfa2d1335615eebc5a4c6145bd18fa06ea0f6e2fd54077be86d64ec29ba30da6cf7e584fa00458ae51653da89cccf6a050b4946f8141c53164c9753d409012439e05e67d92384b6e94b2819cf5b915738e7d1e4d05c4bf6be501061614a15b10c40c70899d54b643fcc4861ed155cacefe1575c2ba1f70ba513da3916b95d1e16549db1a1f8dce722899394070d5773")); c != 0 {
		t.Errorf("TestPluginBLAKE2xProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}
}

func TestPluginBLAKE3512ProcessStreamFunc(t *testing.T) {
	p := NewPluginBLAKE3()
	r := bytes.NewReader(blakeTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("BLAKE3512ProcessStreamFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("e60cd8431ebc2c74d793a4e7256344fb5b4050311f3203a3f62eacdc608bd78b")); c != 0 {
		t.Errorf("BLAKE3512ProcessSteamFunc returned invalid data: %s", d)
	}
}

func TestPluginBLAKE3512ProcessStreamWithCliFlagsFunc(t *testing.T) {
	p := NewPluginBLAKE3()
	r := bytes.NewReader(blakeTestData)
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{"-derive-key", blakeTestKey})
	d, e := p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("7128ffee8eb9e5eca0bb")); c != 0 {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE3()
	r = bytes.NewReader(blakeTestData)
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{"-derive-key", blakeTestKey, "-context", "test context 123"})
	d, e = p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("654833e7182b2fe8c8ec")); c != 0 {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE3()
	r = bytes.NewReader(blakeTestData)
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{"-length", "64"})
	d, e = p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("e60cd8431ebc2c74d793a4e7256344fb5b4050311f3203a3f62eacdc608bd78b9d34e700bc948d24b00be997822acdad00757bd4364cbd5d994531fa492cafa3")); c != 0 {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE3()
	r = bytes.NewReader(blakeTestData)
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{"-key", blakeTestKey})
	d, e = p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e == nil {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc did not return an error with a small key")
	}

	p = NewPluginBLAKE3()
	r = bytes.NewReader(blakeTestData)
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{"-key", blakeTestKey32})
	d, e = p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("cfa55ce67ffc5b8c45bd9d5fa62947e7246783166b1c649fd8f74771919f90e7")); c != 0 {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE3()
	r = bytes.NewReader(blakeTestData)
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	d, e = p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("e60cd8431ebc2c74d793a4e7256344fb5b4050311f3203a3f62eacdc608bd78b")); c != 0 {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}
}

func TestPluginBlakeUsage(t *testing.T) {
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w

	p := NewPluginBLAKE2s()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	flags.Usage()

	p = NewPluginBLAKE2b()
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	flags.Usage()

	p = NewPluginBLAKE2x()
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	flags.Usage()

	p = NewPluginBLAKE3()
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	flags.Usage()
}
