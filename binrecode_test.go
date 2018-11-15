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
			args: args{
				"raw", "raw", bytes.NewBufferString("test"),
			},
			wantOut: "test",
			wantErr: false,
		},
		{
			args: args{
				"raw", "base64", bytes.NewBufferString("test"),
			},
			wantOut: "dGVzdA==",
			wantErr: false,
		},
		{
			args: args{
				"raw", "base64raw", bytes.NewBufferString("test"),
			},
			wantOut: "dGVzdA",
			wantErr: false,
		},
		{
			args: args{
				"raw", "base64url", bytes.NewBufferString("test+-%!\"§!\"$%!&"),
			},
			wantOut: "dGVzdA==",
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
