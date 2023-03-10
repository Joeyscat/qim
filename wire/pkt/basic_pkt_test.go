package pkt

import (
	"bytes"
	"io"
	"testing"
)

func TestBasicPkt_Decode(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantP   BasicPkt
		wantErr bool
	}{
		{
			name:    "OK",
			args:    args{r: bytes.NewReader([]byte{7, 8, 0, 12, 104, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100})},
			wantP:   BasicPkt{Code: 0x0708, Length: 12, Body: []byte("hello, world")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &BasicPkt{}
			if err := p.Decode(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("BasicPkt.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !p.Eq(tt.wantP) {
				t.Errorf("Want: %+v, Got: %+v", tt.wantP, *p)
			}
		})
	}
}

func (p *BasicPkt) Eq(other BasicPkt) bool {
	return p.Code == other.Code && p.Length == other.Length && bytes.Equal(p.Body, other.Body)
}

func TestBasicPkt_Encode(t *testing.T) {
	type fields struct {
		Code   uint16
		Length uint16
		Body   []byte
	}
	tests := []struct {
		name    string
		fields  fields
		wantW   []byte
		wantErr bool
	}{
		{
			name:    "OK",
			fields:  fields{Code: 0x0102, Length: 0},
			wantW:   []byte{1, 2, 0, 0},
			wantErr: false,
		},
		{
			name:    "OK",
			fields:  fields{Code: 0x0708, Length: 12, Body: []byte("hello, world")},
			wantW:   []byte{7, 8, 0, 12, 104, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &BasicPkt{
				Code:   tt.fields.Code,
				Length: tt.fields.Length,
				Body:   tt.fields.Body,
			}
			w := &bytes.Buffer{}
			if err := p.Encode(w); (err != nil) != tt.wantErr {
				t.Errorf("BasicPkt.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.Bytes(); !bytes.Equal(gotW, tt.wantW) {
				t.Errorf("BasicPkt.Encode() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
