package gobbler

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

const API_ROOT string = "https://ws.audioscrobbler.com/2.0/"

type ApiClient interface {
	PostForm(url string, data url.Values) (*http.Response, error)
}

type Gobbler struct {
	LoggedIn   bool
	SessionKey string
	ApiKey     string
	Secret     string
	Client     ApiClient
}

type LoginRequest struct {
	Username string
	Password string
}

type LoginResponse struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
	Session struct {
		Name       string `json:"name"`
		Key        string `json:"key"`
		Subscriber string `json:"subscriber"`
	} `json:"session"`
}

type ScrobbleRequest struct {
	Artist    string
	Track     string
	Album     string
	Timestamp string
	Sk        string
}

type NameResponse struct {
	Text      string `json:"#text"`
	Corrected string `json:"corrected"`
}

type ScrobbleResponse struct {
	Scrobbles struct {
		Scrobble struct {
			Track       NameResponse `json:"track"`
			Artist      NameResponse `json:"artist"`
			Album       NameResponse `json:"album"`
			AlbumArtist NameResponse `json:"albumArtist"`
			Timestamp   string       `json:"timestamp"`
		} `json:"scrobble"`
	} `json:"scrobbles"`
}

func New(apiKey, Secret string) *Gobbler {
	client := &http.Client{}
	return &Gobbler{
		ApiKey: apiKey,
		Secret: Secret,
		Client: client,
	}
}

// Use this method to authenticate against the Last.fm API
// The returned session key will be stored in Gobbler's SessionKey
// field. Depending on the result of the operation, the LoggedIn
// field is either set to false or true
func (g *Gobbler) Login(username, password string) (bool, error) {
	c := &LoginRequest{username, password}
	r := &LoginResponse{}
	data, _ := g.post("auth.getMobileSession", c)
	err := json.Unmarshal(data, &r)
	if err != nil {
		return false, err
	}
	if r.Session.Key != "" {
		g.LoggedIn = true
		g.SessionKey = r.Session.Key
		return g.LoggedIn, nil
	} else {
		return false, errors.New(fmt.Sprintf("Got an error while logging in (%d): %s", r.Error, r.Message))
	}
}

// Use this method to scrobble individual tracks. Artist
// and track are required fields, album can be empty string.
// For return value fields check ScrobbleResponse.
func (g *Gobbler) Scrobble(artist, track, album string) (*ScrobbleResponse, error) {
	if !g.LoggedIn {
		return nil, errors.New("You need to be logged in")
	}
	if artist == "" || track == "" {
		return nil, errors.New("Artist and track must be filled in")
	}
	c := &ScrobbleRequest{
		Artist:    artist,
		Track:     track,
		Timestamp: strconv.FormatInt(time.Now().Unix(), 10),
		Sk:        g.SessionKey,
	}
	if album != "" {
		c.Album = album
	}
	r := &ScrobbleResponse{}
	data, _ := g.post("track.scrobble", c)
	err := json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// This method will take a struct containing only string fields
// and return a map of strings with the keys being lowercased
// names of the fields.
// If the struct contains other fields than strings, this returns
// an error.
func structToMap(data interface{}) (map[string]string, error) {
	m := map[string]string{}
	r := reflect.ValueOf(data).Elem()
	t := r.Type()
	for i := 0; i < r.NumField(); i++ {
		if t := r.Field(i).Kind(); t != reflect.String {
			return nil, errors.New(fmt.Sprintf("The type can only be string, not %s", t))
		}
		m[strings.ToLower(t.Field(i).Name)] = r.Field(i).String()
	}
	return m, nil
}

func (g *Gobbler) post(method string, data interface{}) ([]byte, error) {
	params, _ := structToMap(data)
	params["method"] = method
	params["api_key"] = g.ApiKey

	sig := gobblerSignature(params, g.Secret)
	params["api_sig"] = sig
	params["format"] = "json"

	pv := url.Values{}
	for k, v := range params {
		pv.Add(k, v)
	}

	result, err := g.Client.PostForm(API_ROOT, pv)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func gobblerSignature(pv map[string]string, secret string) string {
	sk := make([]string, 0, len(pv))
	sts := ""
	// Sort the keys in alphabetic order
	for k, _ := range pv {
		sk = append(sk, k)
	}
	sort.Strings(sk)
	// Create the signing string by concatenating <key><value> pairs
	for _, k := range sk {
		sts += fmt.Sprintf("%s%s", k, pv[k])
	}
	// Finally append the secret
	sts += secret
	// Return a MD5 hash of it
	return fmt.Sprintf("%x", md5.Sum([]byte(sts)))
}
