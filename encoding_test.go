package httpio_test

import (
	"bytes"
	"fmt"
	"testing"

	httpio "github.com/advanderveer/go-httpio"
	"github.com/gorilla/schema"
)

func TestEncodingStd(t *testing.T) {
	var ef httpio.EncoderFactory
	ef = &httpio.JSON{}
	if ef.MimeType() != "application/json" {
		t.Fatal("should be application json")
	}

	var df httpio.DecoderFactory
	df = &httpio.XML{}
	if df.MimeType() != "application/xml" {
		t.Fatal("should be application xml")
	}
}

func TestEncodingForm(t *testing.T) {
	var ef httpio.EncoderFactory
	ef = httpio.NewFormEncoding(schema.NewEncoder())
	if ef.MimeType() != "application/x-www-form-urlencoded" {
		t.Fatal("mime type should be application/x-www-form-urlencoded")
	}

	var df httpio.DecoderFactory
	df = httpio.NewFormDecoding(schema.NewDecoder())
	if df.MimeType() != "application/x-www-form-urlencoded" {
		t.Fatal("mime type should be application/x-www-form-urlencoded")
	}

	t.Run("decoding", func(t *testing.T) {
		buf := bytes.NewBufferString(`foo=bar&bar=foo`)
		dec := df.Decoder(buf)
		v := struct {
			Foo string
			Bar string
		}{}
		err := dec.Decode(&v)
		if err != nil {
			t.Fatal(err)
		}

		if v.Bar != "foo" && v.Foo != "bar" {
			t.Fatal("decoding failed")
		}
	})

	t.Run("encoding", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		enc := ef.Encoder(buf)
		v := struct {
			Foo string `schema:"foo"`
			Bar string `schema:"bar"`
		}{"bar", "foo"}
		err := enc.Encode(v)
		if err != nil {
			t.Fatal(err)
		}

		if buf.String() != `bar=foo&foo=bar` {
			t.Fatal("encoding failed, got: %v", buf.String())
		}
	})
}

func TestEncodingFormNoEncoder(t *testing.T) {
	var e httpio.EncoderFactory
	e = httpio.NewFormEncoding(nil)
	if e.MimeType() != "application/x-www-form-urlencoded" {
		t.Fatal("mime type should be application/x-www-form-urlencoded")
	}

	t.Run("encoding", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		enc := e.Encoder(buf)
		v := struct {
			Foo string `schema:"foo"`
			Bar string `schema:"bar"`
		}{"bar", "foo"}
		err := enc.Encode(v)
		if fmt.Sprint(err) != "no form encoder configured" {
			t.Fatal("expected error that indicated that no form encoder was configured")
		}
	})
}

func TestEncodingFormNoDecoder(t *testing.T) {
	var e httpio.DecoderFactory
	e = httpio.NewFormDecoding(nil)
	if e.MimeType() != "application/x-www-form-urlencoded" {
		t.Fatal("mime type should be application/x-www-form-urlencoded")
	}

	t.Run("decoding", func(t *testing.T) {
		buf := bytes.NewBufferString(`foo=bar&bar=foo`)
		dec := e.Decoder(buf)
		v := struct {
			Foo string
			Bar string
		}{}
		err := dec.Decode(&v)
		if fmt.Sprint(err) != "no form decoder configured" {
			t.Fatal("expected error that indicated that no form decoder was configured")
		}
	})
}
