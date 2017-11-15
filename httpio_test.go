package httpio

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/schema"
)

type capitalizeInput struct {
	Name string `json:"name"`
}

type capitalizeOutput struct {
	Err  `json:"error,omitempty"`
	Name string `json:"name,omitempty"`
}

type capitalizeOutputT struct {
	*template.Template
	Name string
}

//Render a custom template for this output
func (o *capitalizeOutputT) Render(w http.ResponseWriter) error {
	if o.Name == "error" {
		return errors.New(o.Name)
	}

	return o.Execute(w, o)
}

func capitalize(i *capitalizeInput) (o *capitalizeOutput, err error) {
	return &capitalizeOutput{Name: strings.ToUpper(i.Name)}, nil
}

func capitalizeWithErr(i *capitalizeInput) (o *capitalizeOutput, err error) {
	return nil, errors.New("failed to capitalize")
}

func TestErrorEmbedding(t *testing.T) {
	out := &capitalizeOutput{}
	if out.UnboxErr() != nil {
		t.Fatal("embedded error should now be nil")
	}

	if out.StatusCode() != -1 {
		t.Fatal("status for embedded error should be -1")
	}

	out.BoxErr(errors.New("some error"))
	if out.UnboxErr().Error() != "some error" {
		t.Fatal("embedded error should now not be nil")
	}

	if out.StatusCode() != http.StatusInternalServerError {
		t.Fatal("status for embedded error should be 500")
	}
}

func TestBoxing(t *testing.T) {
	hio := IO{
		Encoders: map[string]func(w io.Writer) Encoder{
			"application/json": func(w io.Writer) Encoder { return json.NewEncoder(w) },
		},
	}

	v1 := &capitalizeOutput{}
	r1, _ := http.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	hio.boxErrOrRender(w1, r1, v1, errors.New("foo error"))
	if w1.Body.String() != `{"error":"foo error"}`+"\n" {
		t.Fatalf("expected boxed error to show, but got: %v", w1.Body.String())
	}

	if w1.Result().StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status code to be correct but got: %d", w1.Result().StatusCode)
	}

	v2 := &capitalizeInput{}
	r2, _ := http.NewRequest("GET", "/", nil)
	w2 := httptest.NewRecorder()
	hio.boxErrOrRender(w2, r2, v2, errors.New("bar error"))
	if w2.Body.String() != `{"message":"bar error"}`+"\n" {
		t.Fatalf("expected boxed error to show, but got: %v", w2.Body.String())
	}

	if w2.Result().StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status code to be correct but got: %d", w2.Result().StatusCode)
	}
}

func TestCustomRendering(t *testing.T) {
	hio := IO{
		Encoders: map[string]func(w io.Writer) Encoder{
			"application/json": func(w io.Writer) Encoder { return json.NewEncoder(w) },
		},
	}

	t.Run("with nil pointer template", func(t *testing.T) {
		v1 := &capitalizeOutputT{}
		r1, _ := http.NewRequest("GET", "/", nil)
		w1 := httptest.NewRecorder()
		hio.Render(w1, r1, v1)
		if !strings.Contains(w1.Body.String(), "ERR_RENDER_FATAL") {
			t.Fatalf("expected fatal error to show, but got: %v", w1.Body.String())
		}

		if w1.Result().StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status code to be correct but got: %d", w1.Result().StatusCode)
		}
	})

	t.Run("with valid rendering", func(t *testing.T) {
		v2 := &capitalizeOutputT{
			Template: template.Must(template.New("default").Parse(`hello, {{.Name}}`)),
			Name:     "world",
		}
		r2, _ := http.NewRequest("GET", "/", nil)
		w2 := httptest.NewRecorder()
		hio.Render(w2, r2, v2)

		if w2.Body.String() != "hello, world" {
			t.Fatalf("expected body to be equal here, got: %s", w2.Body)
		}

		if w2.Result().StatusCode != http.StatusOK {
			t.Fatalf("expected status code to be correct but got: %d", w2.Result().StatusCode)
		}
	})

	t.Run("with error while rendering", func(t *testing.T) {
		v2 := &capitalizeOutputT{
			Template: template.Must(template.New("default").Parse(`hello, {{.Name}}`)),
			Name:     "error",
		}
		r2, _ := http.NewRequest("GET", "/", nil)
		w2 := httptest.NewRecorder()
		hio.Render(w2, r2, v2)

		if !strings.Contains(w2.Body.String(), "ERR_RENDER_OUTPUT") {
			t.Fatalf("expected fatal error to show, but got: %v", w2.Body.String())
		}

		if w2.Result().StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status code to be correct but got: %d", w2.Result().StatusCode)
		}
	})
}

type failingEncoder struct{}

//Encode will error always
func (ee *failingEncoder) Encode(v interface{}) error {
	return errors.New("test error")
}

func TestEncodingError(t *testing.T) {
	hio := IO{
		Encoders: map[string]func(w io.Writer) Encoder{
			"application/json": func(w io.Writer) Encoder { return &failingEncoder{} },
		},
	}

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	out := capitalizeOutput{}
	hio.Render(w, r, out)

	if w.Body.String() != "failed to encode response body\n" {
		t.Fatalf("expected encoder errors to be shown exactly as is, got: %s", w.Body.String())
	}

	//@TODO we would like to assert the responsecode in this instance but the header cannot be
	//changed after the body has been written
}

func TestIsErrType(t *testing.T) {
	if IsErrType(nil, ErrParseForm) {
		t.Fatal("should not be considered this error")
	}

	if IsErrType(errors.New("foo"), ErrParseForm) {
		t.Fatal("should not be considered this error")
	}

	if !IsErrType(herr{code: ErrParseForm}, ErrParseForm) {
		t.Fatal("expected the error to match the code")
	}
}

func TestUsage(t *testing.T) {
	hio := NewJSON(schema.NewDecoder(), nil)
	r, _ := http.NewRequest("GET", "?Name=my-name", nil)
	w := httptest.NewRecorder()

	in := &capitalizeInput{}
	if render, ok := hio.Parse(w, r, in, &capitalizeOutput{}); ok {
		render(capitalize(in))
	}

	if w.Body.String() != `{"name":"MY-NAME"}`+"\n" {
		t.Fatalf("expected name to be capitalized, got: %s", w.Body)
	}
}
