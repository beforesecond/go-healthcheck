package main

import (
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/spf13/viper"
)

var tests = []func(t *testing.T){
	func(t *testing.T) {
		t.Run("it should check access token", func(t *testing.T) {
			token, err := GetAccessToken()
			if err != nil {
				t.Errorf("Error get access token from LINE API Refresh Token.")
			}

			if token == "" {
				t.Errorf("token is null.")
			}
		})
	},
	func(t *testing.T) {
		t.Run("it should check function generate report", func(t *testing.T) {
			requestJSON, err := GenerateReport()

			if err != nil {
				t.Errorf("Error " + err.Error())
			}

			if requestJSON.TotalWebsites == 0 {
				t.Errorf("Don't have url website")
			}
		})

	},
	func(t *testing.T) {
		t.Run("it should check request api", func(t *testing.T) {
			token, _ := GetAccessToken()
			requestJSON, _ := GenerateReport()

			ok := RequestToReportAPI(requestJSON, token)

			if !ok {
				t.Errorf("Error check request api")
			}
		})

	},
}

func setup() {
	viper.SetConfigFile("./configs/env.dev.yaml")
	err := viper.ReadInConfig()

	// Handle errors reading the config file
	if err != nil {
		log.Fatal("Fatal error config file", err)
	}
	fmt.Printf("\033[1;36m%s\033[0m", "> Setup completed\n")
}

func teardown() {
	// Do something here.
	fmt.Printf("\033[1;36m%s\033[0m", "> Teardown completed")
	fmt.Printf("\n")
}

func TestEverything(t *testing.T) {
	for i, fn := range tests {
		setup()
		t.Run(strconv.Itoa(i), fn)
		teardown()
	}
}
