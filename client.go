package httpio

import (
	"bytes"
	"context"
	"errors"
	"mime"
	"net/http"
	"net/url"

	"github.com/advanderveer/go-httpio/encoding"
	"github.com/advanderveer/go-httpio/handling"
)

//Client is a client that uses encoding stacks to facilitate communication
type Client struct {
	client *http.Client
	base   *url.URL
	encs   encoding.Stack
}

//NewClient will setup a client that encodes and decodes using the encoding stack
func NewClient(hclient *http.Client, base string, def encoding.Encoding, other ...encoding.Encoding) (c *Client, err error) {
	c = &Client{client: hclient, encs: encoding.NewStack(def, other...)}
	c.base, err = url.Parse(base)
	if err != nil {
		return nil, err
	}

	return c, nil
}

//Request output 'out' using method 'm' on path 'p' using headers 'hdr' and input 'in' encoded as
//the default encodinbg scheme from the stack
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

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", def.MimeType())

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	mt, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	e := c.encs.Find(mt)
	if e == nil {
		return errors.New("unspported content type")
	}

	dec := e.Decoder(resp.Body)
	defer resp.Body.Close()
	if resp.Header.Get(handling.HeaderHandlingError) != "" {
		errOut := &handling.Err{} //we will decode into the general Err struct instead
		err = dec.Decode(errOut)
		if err != nil {
			return err
		}

		return errors.New(errOut.Message)
	}

	err = dec.Decode(out)
	if err != nil {
		return err
	}

	return nil
}