package main

import (
	"bytes"
	"io"
	"testing"
)

func Test_doEncode(t *testing.T) {
	type args struct {
		from string
		to   string
		in   io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
		wantErr bool
	}{
		{
			name: "raw-raw",
			args: args{
				"raw", "raw", bytes.NewBufferString("test"),
			},
			wantOut: "test",
			wantErr: false,
		},
		{
			name: "raw-base64",
			args: args{
				"raw", "base64", bytes.NewBufferString("test"),
			},
			wantOut: "dGVzdA==",
			wantErr: false,
		},
		{
			name: "raw-base64raw",
			args: args{
				"raw", "base64raw", bytes.NewBufferString("test"),
			},
			wantOut: "dGVzdA",
			wantErr: false,
		},
		{
			name: "raw-base64-long",
			args: args{
				"raw", "base64url", bytes.NewBufferString("test+-%!\"§!\"$%!&"),
			},
			wantOut: "dGVzdCstJSEiwqchIiQlISY=",
			wantErr: false,
		},
		{
			name: "raw-go",
			args: args{
				"raw", "go", bytes.NewBufferString("test+-%!\"§!\"$%!&12345678901234567890"),
			},
			wantOut: `[]byte{
0x74, 0x65, 0x73, 0x74, 0x2b, 0x2d, 0x25, 0x21, 0x22, 0xc2, 0xa7, 0x21, 0x22, 0x24, 0x25, 0x21
, 0x26, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35
, 0x36, 0x37, 0x38, 0x39, 0x30}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			if err := doEncode(tt.args.from, tt.args.to, tt.args.in, out); (err != nil) != tt.wantErr {
				t.Errorf("doEncode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("doEncode() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
