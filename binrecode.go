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

type reverseWriter struct {
	out io.WriteCloser
	buf bytes.Buffer
}

type gowriter struct {
	out      io.Writer
	notfirst bool
	written  int64
}

func (gw *gowriter) Write(dat []byte) (int, error) {
	cnt := 0
	for _, b := range dat {
		var err error
		if gw.notfirst {
			if gw.written%16 == 0 {
				_, err = fmt.Fprintf(gw.out, ",\n\t%#02x", b)
			} else {
				_, err = fmt.Fprintf(gw.out, ", %#02x", b)
			}
		} else {
			_, err = fmt.Fprintf(gw.out, "%#02x", b)
			gw.notfirst = true
		}
		gw.written++

		if err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}

func (gw *gowriter) Close() error {
	_, err := fmt.Fprint(gw.out, "}")
	return err
}

func (rw *reverseWriter) Write(data []byte) (int, error) {
	return rw.buf.Write(data)
}

func (rw *reverseWriter) Close() error {
	data := rw.buf.Bytes()
	for i := len(data)/2 - 1; i >= 0; i-- {
		opp := len(data) - 1 - i
		data[i], data[opp] = data[opp], data[i]
	}

	if _, err := rw.out.Write(data); err != nil {
		return err
	}
	return rw.out.Close()
}

func encodeHexReverse(out io.Writer) io.WriteCloser {
	return &reverseWriter{
		out: encodeHex(out),
	}
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

func encodeGo(out io.Writer) io.WriteCloser {
	fmt.Fprint(out, "[]byte{\n\t")
	return &gowriter{
		out: out,
	}
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
	"rhex":         encodeHexReverse,
	"go":           encodeGo,
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
	fmt.Printf(usageString, filepath.Base(os.Args[0]), decs, encs)
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
