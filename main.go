package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const (
	DETECT_PORTAL      = "https://detectportal.firefox.com/success.txt"
	PORTAL_URL_PATTERN = `http://(\d{1,3}\.){3}\d{1,3}:\d{1,5}/fgtauth\?[0-9a-zA-Z]+`
	FAILURE_MESSAGE    = "Authentication Failed"
)

func main() {
	home := os.Getenv("HOME")
	path := home + "/.config/sriherwifi/creds.json"
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Unable to read credentials:", err.Error())
		return
	}

	creds := map[string]string{}
	err = json.Unmarshal(content, &creds)
	if err != nil {
		fmt.Println("Unable to parse credentials:", err.Error())
		return
	}

	USERNAME := creds["username"]
	PASSWORD := creds["password"]

	http_client := http.Client{}

	res, err := http_client.Get(DETECT_PORTAL)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	body, _ := io.ReadAll(res.Body)
	res.Body.Close()

	if string(body) == "success" {
		fmt.Println("No Captive Portal")
		return
	}

	r := regexp.MustCompile(PORTAL_URL_PATTERN)
	portal_url := string(r.Find(body))

	if len(portal_url) == 0 {
		fmt.Println("Unable to find to portal url")
		return
	}

	_, err = http.Get(portal_url)
	if err != nil {
		fmt.Println("Unable Load Portal:", err.Error())
		return
	}

	data := url.Values{}
	data.Add("4Tredir", DETECT_PORTAL)
	data.Add("magic", strings.Split(portal_url, "?")[1])
	data.Add("username", USERNAME)
	data.Add("password", PASSWORD)

	res, err = http_client.PostForm(portal_url, data)
	if err != nil {
		fmt.Println("Unable Authenticate in Portal:", err.Error())
		return
	}
	body, _ = io.ReadAll(res.Body)
	res.Body.Close()

	if res.StatusCode == http.StatusOK {
		if ok, _ := regexp.MatchString(FAILURE_MESSAGE, string(body)); ok {
			fmt.Println(FAILURE_MESSAGE)
		} else {
			fmt.Println("Successfully Authenticated!")
		}
	} else {
		fmt.Println("Response:", res.Status)
	}
}
