package main

import (
	"fmt"

	"github.com/zedjones/Reddit-Refresh-Go/reddit_refresh_go/reddit_refresh"
)

const SETTINGS_FILE = "../Settings.json"

func main() {
	config := reddit_refresh.GetConfig(SETTINGS_FILE)
	fmt.Println(config)
	devices := reddit_refresh.GetDevices(config.UserInfo.Token)
	fmt.Println(devices)
	result := reddit_refresh.GetResult("gamedeals", "Battlefield 1")
	fmt.Println(result)
}
