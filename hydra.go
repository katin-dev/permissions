package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type hydra struct {
	Url string
}

func NewHydra(url string) *hydra {
	h := hydra{Url: url}
	return &h
}

func (h *hydra) getUserIdByAccessToken(accessToken string) (string, error) {
	introspectUrl := h.Url + "/oauth2/introspect"

	resp, err := postRequest(introspectUrl, map[string]string{
		"token": accessToken,
		"scope": "",
	})
	if err != nil {
		return "", err
	}

	if resp["active"] == false {
		return "", errors.New("Invalid token (is not active)")
	}

	return resp["sub"].(string), nil
}

func postRequest(postUrl string, data map[string]string) (map[string]interface{}, error) {
	values := url.Values{}
	for key, value := range data {
		values[key] = []string{value}
	}

	resp, err := http.PostForm(
		postUrl,
		values,
	)

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("Hydra response code is " + string(resp.StatusCode))
		return nil, errors.New("Hydra response status code id not 200")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Read error: %s", err.Error())
		return nil, err
	}

	token := make(map[string]interface{})
	err = json.Unmarshal(body, &token)
	if err != nil {
		log.Printf("Json error: %s", err.Error())
		return nil, err
	}

	return token, nil
}
