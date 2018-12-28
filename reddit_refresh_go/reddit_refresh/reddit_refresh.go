package reddit_refresh

import (
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
		device_data := device.(map[string]interface{})
		if device_data["nickname"] != nil {
			nickname = device_data["nickname"].(string)
		} else {
			continue
		}
		iden = device_data["iden"].(string)
		devices_map[nickname] = iden
	}
	return devices_map
}
