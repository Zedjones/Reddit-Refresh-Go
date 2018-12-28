package main

import (
	"fmt"

	"../reddit_refresh"
)

const SETTINGS_FILE = "../Settings.json"

func main() {
	config := reddit_refresh.GetConfig(SETTINGS_FILE)
	fmt.Println(config)
	devices := reddit_refresh.GetDevices(config.UserInfo.Token)
	fmt.Println(devices)
}
