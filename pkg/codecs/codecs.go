package codecs

import (
	"bytes"
	"io"
)

// encodeStream copies r through the encoder built by newEnc (which writes to w)
// and closes it. Used by codecs whose encoder is an io.WriteCloser.
func encodeStream(r io.Reader, w io.Writer, newEnc func(io.Writer) io.WriteCloser) error {
	enc := newEnc(w)
	if _, err := io.Copy(enc, r); err != nil {
		return err
	}
	return enc.Close()
}

// decodeTrimmed reads all of r, trims surrounding whitespace (so a trailing
// newline from the shell does not break decoding) and streams the input through
// the decoder built by newDec into w.
func decodeTrimmed(r io.Reader, w io.Writer, newDec func(io.Reader) io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, newDec(bytes.NewReader(bytes.TrimSpace(data))))
	return err
}
