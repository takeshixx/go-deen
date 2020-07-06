package codecs

import (
	"bytes"
	"flag"
	"fmt"
	"io"

	"github.com/dgrijalva/jwt-go"
	"github.com/takeshixx/deen/pkg/types"
)

func doJWT(reader *io.Reader, secret string, key []byte, alg string) ([]byte, error) {
	var outBuf bytes.Buffer
	var err error

	// TODO: check alg

	return outBuf.Bytes(), err
}

func undoJWT(reader *io.Reader, verify bool) ([]byte, error) {
	var outBuf bytes.Buffer
	var err error

	inBuf := new(bytes.Buffer)
	inBuf.ReadFrom(*reader)
	tokenString := inBuf.String()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		secret := []byte("my_secret")
		return secret, nil
	})

	if verify {
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			fmt.Println(claims["foo"], claims["nbf"])
		} else {
			fmt.Println(err)
		}
	}

	return outBuf.Bytes(), err
}

// NewPluginJwt creates a new PluginJwt object
func NewPluginJwt() (p types.DeenPlugin) {
	p.Name = "jwt"
	p.Aliases = []string{".jwt"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error

		return outBuf.Bytes(), err
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error

		return outBuf.Bytes(), err
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return undoJWT(&reader, false)
	}
	p.UnprocessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error

		return outBuf.Bytes(), err
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		jwtCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		jwtCmd.Usage = func() {
			fmt.Printf("Usage of %s:\n\n", p.Name)
			fmt.Printf("JSON Web Tokens (JWT) (RFC7519).\n\n")
			jwtCmd.PrintDefaults()
		}
		jwtCmd.String("alg", "", "algorithm")
		jwtCmd.String("secret", "", "secret")
		jwtCmd.Parse(args)
		return jwtCmd
	}
	return
}
