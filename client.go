package httpio

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/url"
)

//ErrReceiver is used by the client to determine if the response holds an error value and specify the
//struct to decode it into. If the reponse holds an error this function should return a non-nil value
type ErrReceiver func(ctx context.Context, resp *http.Response) error

//Client is a client that uses encoding stacks to facilitate communication
type Client struct {
	client *http.Client
	base   *url.URL
	encs   EncoderList
	decs   DecoderList

	ErrReceiver ErrReceiver
}

type stdClientErr struct {
	Message string `json:"message"`
}

func (e *stdClientErr) Error() string { return e.Message }

//NewClient will setup a client that encodes and decodes using the encoding stack
func NewClient(hclient *http.Client, base string, def EncoderFactory, defd DecoderFactory, others ...DecoderFactory) (c *Client, err error) {
	encs := EncoderList{def}
	decs := DecoderList{defd}
	decs = append(decs, others...)

	c = &Client{
		client: hclient,
		encs:   encs,
		decs:   decs,

		//@TODO move this to some Std implementation
		ErrReceiver: func(ctx context.Context, resp *http.Response) error {
			if resp.Header.Get("X-Has-Handling-Error") != "" {
				errOut := &stdClientErr{} //we will decode into the general Err struct instead
				return errOut
			}

			return nil
		},
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
	def := c.encs.Default()

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
	decf := c.decs.Find(mt)
	if decf == nil {
		return fmt.Errorf("httpio/client: no encoder for media type '%s'", mt)
	}

	dec := decf.Decoder(resp.Body)
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
