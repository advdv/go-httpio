package httpio

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/url"

	"github.com/advanderveer/go-httpio/encoding"
	"github.com/advanderveer/go-httpio/handling"
)

//Client is a client that uses encoding stacks to facilitate communication
type Client struct {
	client      *http.Client
	base        *url.URL
	encs        encoding.Stack
	ErrReceiver handling.ErrReceiver
}

//NewClient will setup a client that encodes and decodes using the encoding stack
func NewClient(hclient *http.Client, base string, def encoding.Encoding, other ...encoding.Encoding) (c *Client, err error) {
	c = &Client{
		client:      hclient,
		encs:        encoding.NewStack(def, other...),
		ErrReceiver: handling.HeaderErrReceiver,
	}
	c.base, err = url.Parse(base)
	if err != nil {
		return nil, err
	}

	return c, nil
}

//Request output 'out' using method 'm' on path 'p' using headers 'hdr' and input 'in' encoded as
//the default encodinbg scheme from the stack. The "Content-Type" header will be set regardless of
//what is provided as an argument
func (c *Client) Request(ctx context.Context, m, p string, hdr http.Header, in, out interface{}) (err error) {
	def, err := c.encs.Default()
	if err != nil {
		return err
	}

	body := bytes.NewBuffer(nil)
	enc := def.Encoder(body)
	err = enc.Encode(in)
	if err != nil {
		return err
	}

	ref, err := url.Parse(p)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(m, c.base.ResolveReference(ref).String(), body)
	if err != nil {
		return err
	}

	if hdr != nil {
		req.Header = hdr //@TODO test this once the library allows the server impl to retrieve certain header values
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", def.MimeType())

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	mt, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	e := c.encs.Find(mt)
	if e == nil {
		return fmt.Errorf("httpio/client: no encoder for mediate type '%s'", mt)
	}

	dec := e.Decoder(resp.Body)
	defer resp.Body.Close()
	errOut := c.ErrReceiver(ctx, resp)
	if errOut != nil {
		err = dec.Decode(errOut)
		if err != nil {
			return err
		}

		return errOut
	}

	err = dec.Decode(out)
	if err != nil {
		return err
	}

	return nil
}
