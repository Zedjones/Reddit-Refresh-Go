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

const devicesURL = "https://api.pushbullet.com/v2/devices"
const pushesURL = "https://api.pushbullet.com/v2/pushes"
const searchURL = "https://www.reddit.com/%s/search.json?q=%s&sort=new&restrict_sr=on&limit=1"

//UserInfo hold infos about a user used for pushes
type UserInfo struct {
	Token string
}

//SubResult holds information about a search result
type SubResult struct {
	Url   string
	Title string
}

//ProgramConfig just holds the refresh interval
type ProgramConfig struct {
	Interval float32
}

//RRConfig holds all the information needs to get and push results
type RRConfig struct {
	UserInfo      UserInfo
	LastResult    map[string]string
	Subreddits    map[string][]string
	Devices       map[string]string
	ProgramConfig ProgramConfig
}

//GetConfig loads a config from a JSON file provided
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

//GetDevices gets the Pushbullet devices for a user
//using their API access token
func GetDevices(token string) map[string]string {
	var devicesMap map[string]string
	req, err := http.NewRequest("GET", devicesURL, nil)
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

//SendPushLink sends a link as a push to the specified device
//using the provided API access token
func SendPushLink(device string, token string, result SubResult) {
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
	req, err := http.NewRequest("POST", pushesURL, bytes.NewBuffer(json))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not construct HTTP request.")
		panic(err)
	}
	req.Header.Add("Access-Token", token)
	req.Header.Set("Content-Type", "application/json")
	client.Do(req)
}

//GetResult checks a subreddit for the given search and
//returns the latest result
func GetResult(sub string, search string) SubResult {
	if !strings.Contains(sub, "/r") {
		sub = fmt.Sprintf("r/%s", sub)
	}
	if strings.Contains(search, " ") {
		search = strings.Replace(search, " ", "+", -1)
	}
	searchURL := fmt.Sprintf(searchURL, sub, search)
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
		fmt.Println(sub, search)

		fmt.Fprintln(os.Stderr, "Invalid subreddit provided.")
		return SubResult{}
	}
	item := results[0].(map[string]interface{})
	perma := item["data"].(map[string]interface{})["permalink"].(string)
	link := fmt.Sprintf("https://www.reddit.com%s", perma)
	title := item["data"].(map[string]interface{})["title"].(string)
	return SubResult{link, title}
}
