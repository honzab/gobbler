package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
)

const API_ROOT string = "https://ws.audioscrobbler.com/2.0/"

type Gobbler struct {
	ApiKey     string
	Secret     string
	LoggedIn   bool
	SessionKey string
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
	Artist string
	Track  string
	Album  string
}

type ScrobbleResponse struct {
	Scrobbles []struct {
		Blah string
	} `json:"scrobbles"`
}

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

func (g *Gobbler) Scrobble(artist, track, album string) (bool, error) {
	// TODO
	return false, nil
}

func (g *Gobbler) post(method string, data interface{}) ([]byte, error) {
	params := map[string]string{}
	params["method"] = method
	params["api_key"] = g.ApiKey

	r := reflect.ValueOf(data).Elem()
	t := r.Type()
	for i := 0; i < r.NumField(); i++ {
		params[strings.ToLower(t.Field(i).Name)] = r.Field(i).String()
	}

	sig, err := gobblerSignature(params, g.Secret)
	if err != nil {
		panic("Can't sign this")
	}
	params["api_sig"] = sig
	params["format"] = "json"

	pv := url.Values{}
	for k, v := range params {
		pv.Add(k, v)
	}

	result, err := http.PostForm(API_ROOT, pv)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	log.Println(strings.TrimSpace(string(body))) // TODO Remove me maybe?
	return body, nil
}

func gobblerSignature(pv map[string]string, secret string) (string, error) {
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
	return fmt.Sprintf("%x", md5.Sum([]byte(sts))), nil
}
