package hashs

import (
	"bytes"
	"testing"
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
	testFlags := p.AddCliOptionsFunc(&p, []string{"-key", blakeTestKey})
	d, e := p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE2sProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("183f16fed32775091e9f63171cbdca57e6c0642dffa08c7a25ea58aa7b50f2d5")); c != 0 {
		t.Errorf("BLAKE2sProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE2s()
	r = bytes.NewReader(blakeTestData)
	testFlags = p.AddCliOptionsFunc(&p, []string{})
	d, e = p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE2sProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("4088080c149a6165b9a086ef4aaeb13df5fc7ffb83d5731ed9692320b5634c50")); c != 0 {
		t.Errorf("BLAKE2sProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}
}

func TestPluginBLAKE2s128ProcessStreamFunc(t *testing.T) {
	p := NewPluginBLAKE2s128()
	r := bytes.NewReader(blakeTestData)
	testFlags := p.AddCliOptionsFunc(&p, []string{"-key", blakeTestKey})
	d, e := p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE2s128ProcessStreamFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("b789903bd37697727692abb9f0494bad")); c != 0 {
		t.Errorf("BLAKE2s128ProcessStreamFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE2s128()
	r = bytes.NewReader(blakeTestData)
	testFlags = p.AddCliOptionsFunc(&p, []string{})
	d, e = p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e == nil {
		t.Error("BLAKE2s128ProcessStreamFunc without a key did not trigger an error")
	}
}

func TestPluginBLAKE2s128ProcessStreamWithCliFlagsFunc(t *testing.T) {
	p := NewPluginBLAKE2s128()
	r := bytes.NewReader(blakeTestData)
	testFlags := p.AddCliOptionsFunc(&p, []string{"-key", blakeTestKey})
	d, e := p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE2s128ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("b789903bd37697727692abb9f0494bad")); c != 0 {
		t.Errorf("BLAKE2s128ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
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
	testFlags := p.AddCliOptionsFunc(&p, []string{"-key", blakeTestKey})
	d, e := p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE2bProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("4e8c474aa515d314feb9cb0893e2bddaad49f007fbd1f0538776f2c11d9c9d04732b70a024642400b14707276928c94429b109424245156e438503aa312036d9")); c != 0 {
		t.Errorf("BLAKE2bProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE2b()
	r = bytes.NewReader(blakeTestData)
	testFlags = p.AddCliOptionsFunc(&p, []string{})
	d, e = p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE2bProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("e3e8bca1c407f1ce36642d64c334bbc572f7ad06e00425d2abc567e094e9e82862b3d8f200647273ec4f1d36cc5b7371b6a4cf7ea6725529ce71ea9c68eeb66c")); c != 0 {
		t.Errorf("BLAKE2bProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}
}

func TestPluginBLAKE2b384ProcessStreamFunc(t *testing.T) {
	p := NewPluginBLAKE2b384()
	r := bytes.NewReader(blakeTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("BLAKE2b384ProcessStreamFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("4b9857fa13e3c90de236b5004bd9ed91c4b5da1a83e9b75f6450c07e4577d971a478e5865e12a6d262498b04b9560847")); c != 0 {
		t.Errorf("BLAKE2b384ProcessSteamFunc returned invalid data: %s", d)
	}
}

func TestPluginBLAKE2b384ProcessStreamWithCliFlagsFunc(t *testing.T) {
	p := NewPluginBLAKE2b384()
	r := bytes.NewReader(blakeTestData)
	testFlags := p.AddCliOptionsFunc(&p, []string{"-key", blakeTestKey})
	d, e := p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE2b384ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("19297fafa024f3408591cac9343686a39cb4f306148f40b38fe51a4a0937a87f4367c66744e577c6360748f78bdad648")); c != 0 {
		t.Errorf("BLAKE2b384ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE2b384()
	r = bytes.NewReader(blakeTestData)
	testFlags = p.AddCliOptionsFunc(&p, []string{})
	d, e = p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE2b384ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("4b9857fa13e3c90de236b5004bd9ed91c4b5da1a83e9b75f6450c07e4577d971a478e5865e12a6d262498b04b9560847")); c != 0 {
		t.Errorf("BLAKE2b384ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}
}

func TestPluginBLAKE2b256ProcessStreamFunc(t *testing.T) {
	p := NewPluginBLAKE2b256()
	r := bytes.NewReader(blakeTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("BLAKE2b256ProcessStreamFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("9c5a85bed8e8f3fb41e1d21f7dd0ae4033d506aa125da2c03f2dd527fb6b1868")); c != 0 {
		t.Errorf("BLAKE2b256ProcessSteamFunc returned invalid data: %s", d)
	}
}

func TestPluginBLAKE2b256ProcessStreamWithCliFlagsFunc(t *testing.T) {
	p := NewPluginBLAKE2b256()
	r := bytes.NewReader(blakeTestData)
	testFlags := p.AddCliOptionsFunc(&p, []string{"-key", blakeTestKey})
	d, e := p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE2b256ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("6979810b7ec2cc132a2bcaaef73ab661d3e38d6248694c4ddc3b6261772b3128")); c != 0 {
		t.Errorf("BLAKE2b256ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE2b256()
	r = bytes.NewReader(blakeTestData)
	testFlags = p.AddCliOptionsFunc(&p, []string{})
	d, e = p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE2b256ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("9c5a85bed8e8f3fb41e1d21f7dd0ae4033d506aa125da2c03f2dd527fb6b1868")); c != 0 {
		t.Errorf("BLAKE2b256ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
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
	testFlags := p.AddCliOptionsFunc(&p, []string{"-derive-key", blakeTestKey})
	d, e := p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("7128ffee8eb9e5eca0bb")); c != 0 {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE3()
	r = bytes.NewReader(blakeTestData)
	testFlags = p.AddCliOptionsFunc(&p, []string{"-derive-key", blakeTestKey, "-context", "test context 123"})
	d, e = p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("654833e7182b2fe8c8ec")); c != 0 {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE3()
	r = bytes.NewReader(blakeTestData)
	testFlags = p.AddCliOptionsFunc(&p, []string{"-length", "64"})
	d, e = p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("e60cd8431ebc2c74d793a4e7256344fb5b4050311f3203a3f62eacdc608bd78b9d34e700bc948d24b00be997822acdad00757bd4364cbd5d994531fa492cafa3")); c != 0 {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE3()
	r = bytes.NewReader(blakeTestData)
	testFlags = p.AddCliOptionsFunc(&p, []string{"-key", blakeTestKey})
	d, e = p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e == nil {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc did not return an error with a small key")
	}

	p = NewPluginBLAKE3()
	r = bytes.NewReader(blakeTestData)
	testFlags = p.AddCliOptionsFunc(&p, []string{"-key", blakeTestKey32})
	d, e = p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("cfa55ce67ffc5b8c45bd9d5fa62947e7246783166b1c649fd8f74771919f90e7")); c != 0 {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}

	p = NewPluginBLAKE3()
	r = bytes.NewReader(blakeTestData)
	testFlags = p.AddCliOptionsFunc(&p, []string{})
	d, e = p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("e60cd8431ebc2c74d793a4e7256344fb5b4050311f3203a3f62eacdc608bd78b")); c != 0 {
		t.Errorf("BLAKE3512ProcessStreamWithCliFlagsFunc returned invalid data: %s", d)
	}
}
