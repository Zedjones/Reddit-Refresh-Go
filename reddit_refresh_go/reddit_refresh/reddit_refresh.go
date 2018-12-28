package reddit_refresh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const DEVICES_URL = "https://api.pushbullet.com/v2/devices"
const PUSHES_URL = "https://api.pushbullet.com/v2/pushes"

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
	var devices_map map[string]string
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
	devices_map = make(map[string]string)
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
		devices_map[nickname] = iden
	}
	return devices_map
}

func SendPushLink(devices []string, token string, result SubResult) {
	for _, device := range devices {
		client := &http.Client{}
		test_token := "o.fVHr05C1TTUIjLyF54Fn3cFpeWvSpe62"
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
		req.Header.Add("Access-Token", test_token)
		req.Header.Set("Content-Type", "application/json")
		client.Do(req)
	}
}
