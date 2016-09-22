package gobbler

import (
	"net/http"
	"net/url"
	"testing"
)

func TestEmptyValuesAndEmptySecret(t *testing.T) {
	values := map[string]string{}
	expected := "d41d8cd98f00b204e9800998ecf8427e"
	s := gobblerSignature(values, "")
	if s != expected {
		t.Fatalf("Expected %s, got %s", expected, s)
	}
}

func TestSignature(t *testing.T) {
	tc := map[string]map[string]string{
		"96a9c2d1038f634570bc0e5270d0a5e2": map[string]string{
			"do":   "you",
			"like": "this",
			"test": "?",
		},
		"912ec803b2ce49e4a541068d495ab570": map[string]string{},
		"3f0bf720b8c73fc1409b029aa5ce7b13": map[string]string{
			"some_utf8": "öäåáýíčůščé",
		},
	}
	secret := "asdf"

	for k, v := range tc {
		s := gobblerSignature(v, secret)
		if s != k {
			t.Fatalf("Expected %s, got %s", k, s)
		}
	}
}

type TestStruct struct {
	A string
	B string
}

type TestStruct2 struct {
	A string
	B int
}

func TestStructToMap(t *testing.T) {
	d := &TestStruct{A: "hej", B: "då"}
	expected := map[string]string{
		"a": "hej",
		"b": "då",
	}
	got, err := structToMap(d)
	if err != nil {
		t.Error(err)
	}
	for k, v := range expected {
		if got[k] != v {
			t.Fatalf("Expected key %s to contain %s, not %s", k, v, got[k])
		}
	}
}

func TestStructToMapFail(t *testing.T) {
	d := &TestStruct2{A: "hej", B: 10}
	got, err := structToMap(d)
	if err == nil {
		t.Fatalf("Expected error, got none")
	}
	if got != nil {
		t.Fatalf("Expected return value to be nil, is %s", got)
	}
}

type TestClient struct{}

func (tc *TestClient) PostForm(url string, data url.Values) (*http.Response, error) {
	return nil, nil
}

func TestPost(t *testing.T) {
	tc := &TestClient{}
	g := &Gobbler{
		Client: tc,
	}
	g.post("Hej", &TestStruct{A: "hej", B: "ho"})
	// g := New("key", "secret")

	// g.post("method",
	// TODO
}
