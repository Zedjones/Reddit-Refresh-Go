package reddit_refresh 

import (
	"os"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type UserInfo struct {
	Token	string
}

type ProgramConfig struct {
	Interval	float32
}

type RRConfig struct {
	User_info		UserInfo
	Last_result 	map[string]string 
	Subreddits		map[string][]string
	Devices			map[string]string
	Program_config	ProgramConfig
}

func Add(a int, b int) int {
	return a + b
}

func GetConfig(file_name string) RRConfig {
	file, _ := os.Open(file_name)
	defer file.Close()
	content, _ := ioutil.ReadFile(file_name)
	configuration := RRConfig{}
	err := json.Unmarshal(content, &configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	return configuration
}