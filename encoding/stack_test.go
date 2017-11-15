package encoding

import "testing"

func TestStack(t *testing.T) {
	stack1 := make(Stack, 0)
	_, err := stack1.Default()
	if err == nil {
		t.Fatal("should throw error on empty stack")
	}

	stack2 := NewStack(&JSON{}, &XML{})
	enc1, err := stack2.Default()
	if err != nil {
		t.Fatal(err)
	}

	if enc1.MimeType() != "application/json" {
		t.Fatal("default encoding should be json")
	}

	enc2 := stack2.Find("application/xml")
	if enc2 == nil {
		t.Fatal("should be able to find xml encoding")
	}

	enc3 := stack2.Find("foo/bar")
	if enc3 != nil {
		t.Fatal("should not be able to find bogus encoding")
	}
}
