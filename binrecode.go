package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func makeEncodeBase64(encoding *base64.Encoding) func(io.Writer) io.WriteCloser {
	return func(out io.Writer) io.WriteCloser { return base64.NewEncoder(encoding, out) }
}

func encodeHex(out io.Writer) io.WriteCloser {
	oe := hex.NewEncoder(out)
	return nopWriteCloser{oe}
}

func encode0xHex(out io.Writer) io.WriteCloser {
	out.Write([]byte("0x"))
	return encodeHex(out)
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}

func encodeRaw(out io.Writer) io.WriteCloser {
	wc, ok := out.(io.WriteCloser)
	if ok {
		return wc
	}
	return nopWriteCloser{out}
}

func decodeRaw(in io.Reader) io.Reader {
	return in
}

func decodeHex(in io.Reader) io.Reader {
	return hex.NewDecoder(in)
}

func makeDecodeBase64(encoding *base64.Encoding) func(io.Reader) io.Reader {
	return func(in io.Reader) io.Reader { return base64.NewDecoder(encoding, in) }
}

var usageString = `Usage %v <sourceencoding> <targetencoding> [data]
if data is not set, data is read from stdin
inputs: %v
outputs: %v
`
var encoders = map[string]func(io.Writer) io.WriteCloser{
	"base64":       makeEncodeBase64(base64.StdEncoding),
	"base64raw":    makeEncodeBase64(base64.RawStdEncoding),
	"base64url":    makeEncodeBase64(base64.URLEncoding),
	"base64urlraw": makeEncodeBase64(base64.RawURLEncoding),
	"raw":          encodeRaw,
	"hex":          encodeHex,
	"0xhex":        encode0xHex,
}

var decoders = map[string]func(io.Reader) io.Reader{
	"raw":          decodeRaw,
	"hex":          hex.NewDecoder,
	"base64":       makeDecodeBase64(base64.StdEncoding),
	"base64raw":    makeDecodeBase64(base64.RawStdEncoding),
	"base64url":    makeDecodeBase64(base64.URLEncoding),
	"base64urlraw": makeDecodeBase64(base64.RawURLEncoding),
}

func usage() {
	var encs, decs []string
	for e := range encoders {
		encs = append(encs, e)
	}
	for d := range decoders {
		decs = append(decs, d)
	}
	sort.Strings(encs)
	sort.Strings(decs)
	fmt.Printf(usageString, filepath.Base(os.Args[0]), encs, decs)
	os.Exit(1)

}

func doEncode(from, to string, in io.Reader, out io.Writer) error {
	enc := encoders[to]
	dec := decoders[from]
	if enc == nil {
		return fmt.Errorf("unknown target format: %v", to)
	}
	if dec == nil {
		return fmt.Errorf("unknown source format: %v", from)
	}

	outc := enc(out)
	defer outc.Close()
	in = dec(in)
	_, err := io.Copy(outc, in)
	return err
}

func main() {
	if len(os.Args) < 3 {
		usage()
	}

	var in io.Reader
	in = os.Stdin
	if len(os.Args) >= 4 {
		in = bytes.NewBufferString(os.Args[3])
	}

	err := doEncode(os.Args[1], os.Args[2], in, os.Stdout)
	if err != nil {
		fmt.Println(err)
		usage()
	}
}
