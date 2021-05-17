package formatters

import "testing"

func TestNewPluginJwt(t *testing.T) {

}

func TestDecodeToken(t *testing.T) {
	// Input is a default JWT Token
}

func TestEncodeToken(t *testing.T) {
	// Input is a string with a JSON object
}

func TestNoneSignatureToken(t *testing.T) {
	// Input should be a default token, output should set alg to "none"
}

func TestDecodeOutputFormats(t *testing.T) {
	// Take a default token, output in all 3 supported formats
}
