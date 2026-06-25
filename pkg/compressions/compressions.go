package compressions

import "io"

// compressStream builds a compressing WriteCloser around w, copies r into it
// and closes it. It returns on the first error (never continuing with a nil
// writer).
func compressStream(r io.Reader, w io.Writer, newC func(io.Writer) (io.WriteCloser, error)) error {
	c, err := newC(w)
	if err != nil {
		return err
	}
	if _, err := io.Copy(c, r); err != nil {
		return err
	}
	return c.Close()
}

// decompressStream builds a decompressing reader around r and copies it to w.
func decompressStream(r io.Reader, w io.Writer, newD func(io.Reader) (io.Reader, error)) error {
	d, err := newD(r)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, d)
	return err
}
