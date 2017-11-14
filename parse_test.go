package httpio_test

// type testInput struct {
// 	Name  string
// 	Image string `schema:"form-image"`
// }
//
// type testErr interface {
// 	Code() httpio.ErrType
// }
//
// func TestDefaultParse(t *testing.T) {
// 	for _, c := range []struct {
// 		Name     string
// 		HIO      *httpio.IO
// 		Method   string
// 		Path     string
// 		Body     string
// 		Headers  http.Header
// 		ExpInput *testInput
// 		ExpErr   httpio.ErrType
// 	}{
// 		{
// 			Name:     "plain GET should not decode as it has no content",
// 			HIO:      &httpio.IO{},
// 			Method:   http.MethodGet,
// 			ExpInput: &testInput{Name: ""},
// 		},
// 		{
// 			Name:     "GET with query should not decode as no form decoder is configured",
// 			HIO:      &httpio.IO{},
// 			Method:   http.MethodGet,
// 			Path:     "?name=foo",
// 			ExpInput: &testInput{Name: ""},
// 		},
// 		{
// 			Name:     "GET with query should decode with form decoder and custom field name",
// 			HIO:      &httpio.IO{FormDecoder: schema.NewDecoder()},
// 			Method:   http.MethodGet,
// 			Path:     "?name=foo&form-image=bar",
// 			ExpInput: &testInput{Name: "foo", Image: "bar"},
// 		},
// 		// {
// 		// 	Name:   "GET with query should decode should fail on invalid field",
// 		// 	HIO:    &httpio.IO{FormDecoder: schema.NewDecoder()},
// 		// 	Method: http.MethodGet,
// 		// 	Path:   "?name=foo&bogus=bogus&form-image=bar2",
// 		// 	ExpErr: httpio.ErrDecodeForm,
// 		// },
// 	} {
// 		t.Run(c.Name, func(t *testing.T) {
// 			var b io.Reader
// 			if c.Body != "" {
// 				b = bytes.NewBufferString(c.Body)
// 			}
//
// 			r, err := http.NewRequest(c.Method, c.Path, b)
// 			if err != nil {
// 				t.Fatalf("failed to create request: %v", err)
// 			}
//
// 			in := &testInput{}
// 			err = c.HIO.Parse(r, in)
// 			if c.ExpErr != "" {
// 				eerr, ok := err.(testErr)
// 				if !ok {
// 					t.Fatal("parse error should be castable to test error")
// 				}
//
// 				if eerr.Code() != c.ExpErr {
// 					t.Fatalf("error code should be '%s' but got '%s'", c.ExpErr, eerr.Code())
// 				}
//
// 			}
//
// 			// if !reflect.DeepEqual(err, c.ExpErr) {
// 			//   if c.ExpErr.Code() ==
// 			//
// 			// 	t.Fatalf("Expected error: %#v, got: %#v", c.ExpErr, err)
// 			// }
//
// 			if !reflect.DeepEqual(in, c.ExpInput) {
// 				t.Fatalf("Expected input: %#v, got: %#v", c.ExpInput, in)
// 			}
// 		})
// 	}
// }
