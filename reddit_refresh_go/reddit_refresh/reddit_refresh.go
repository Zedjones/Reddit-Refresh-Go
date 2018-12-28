package reddit_refresh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const DEVICES_URL = "https://api.pushbullet.com/v2/devices"
const PUSHES_URL = "https://api.pushbullet.com/v2/pushes"
const SEARCH_URL = "https://www.reddit.com/%s/search.json?q=%s&sort=new&restrict_sr=on&limit=1"

type UserInfo struct {
	Token string
}

type SubResult struct {
	Url   string
	Title string
}

type ProgramConfig struct {
	Interval float32
}

type RRConfig struct {
	UserInfo      UserInfo
	LastResult    map[string]string
	Subreddits    map[string][]string
	Devices       map[string]string
	ProgramConfig ProgramConfig
}

func GetConfig(file_name string) RRConfig {
	content, err := ioutil.ReadFile(file_name)
	if err != nil {
		panic(err)
	}
	configuration := RRConfig{}
	err = json.Unmarshal(content, &configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	return configuration
}

func GetDevices(token string) map[string]string {
	var devicesMap map[string]string
	req, err := http.NewRequest("GET", DEVICES_URL, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not construct HTTP request.")
		panic(err)
	}
	client := &http.Client{}
	req.SetBasicAuth(token, "")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error sending HTTP request.")
		panic(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	devicesMap = make(map[string]string)
	for _, device := range result["devices"].([]interface{}) {
		var nickname string
		var iden string
		deviceData := device.(map[string]interface{})
		if deviceData["nickname"] != nil {
			nickname = deviceData["nickname"].(string)
		} else {
			continue
		}
		iden = deviceData["iden"].(string)
		devicesMap[nickname] = iden
	}
	return devicesMap
}

func SendPushLink(devices []string, token string, result SubResult) {
	for _, device := range devices {
		client := &http.Client{}
		data := make(map[string]string)
		data["title"] = result.Title
		data["url"] = result.Url
		data["type"] = "link"
		data["device_iden"] = device
		json, err := json.Marshal(data)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error converting data map in JSON string.")
			panic(err)
		}
		req, err := http.NewRequest("POST", PUSHES_URL, bytes.NewBuffer(json))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not construct HTTP request.")
			panic(err)
		}
		req.Header.Add("Access-Token", token)
		req.Header.Set("Content-Type", "application/json")
		client.Do(req)
	}
}

func GetResult(sub string, search string) SubResult {
	if !strings.Contains(sub, "/r") {
		sub = fmt.Sprintf("r/%s", sub)
	}
	if strings.Contains(search, " ") {
		search = strings.Replace(search, " ", "+", -1)
	}
	searchURL := fmt.Sprintf(SEARCH_URL, sub, search)
	client := &http.Client{}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not construct HTTP request.")
		panic(err)
	}
	req.Header.Set("User-Agent", "reddit-refresh-go-1.0")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting search results.")
		panic(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	results := result["data"].(map[string]interface{})["children"].([]interface{})
	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "Invalid subreddit provided.")
		return SubResult{}
	}
	item := results[0].(map[string]interface{})
	perma := item["data"].(map[string]interface{})["permalink"].(string)
	link := fmt.Sprintf("https://www.reddit.com%s", perma)
	title := item["data"].(map[string]interface{})["title"].(string)
	return SubResult{link, title}
}
