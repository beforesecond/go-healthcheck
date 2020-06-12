package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// RequestJSON is response website checking
type RequestJSON struct {
	TotalWebsites int   `json:"total_websites"`
	Success       int   `json:"success"`
	Failure       int   `json:"failure"`
	TotalTime     int64 `json:"total_time"`
}

// ReadCsvFile is read csv file
func ReadCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}

// GetAccessToken is get access token from LINE refresh token
func GetAccessToken() (string, error) {
	clientID := viper.GetString("line.client_id")
	clientSecret := viper.GetString("line.client_secret")
	refreshToken := viper.GetString("line.refresh_token")
	endPointToken := viper.GetString("line.end_point_token")
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	req, err := http.NewRequest("POST", endPointToken, strings.NewReader(data.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return "", err
	}
	if resp.StatusCode == http.StatusOK {
		var data map[string]interface{}
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			panic(err.Error())
		}
		json.Unmarshal(body, &data)

		return data["access_token"].(string), nil
	}
	defer resp.Body.Close()
	return "", err
}

//RequestToReportAPI Send Request result
func RequestToReportAPI(requestJSON *RequestJSON, token string) bool {
	endPointReport := viper.GetString("line.end_point_report")
	totalWebsites := strconv.Itoa(requestJSON.TotalWebsites)
	success := strconv.Itoa(requestJSON.Success)
	failure := strconv.Itoa(requestJSON.Failure)
	totalTime := strconv.FormatInt(requestJSON.TotalTime, 10)
	var jsonStr = []byte(`{
		"total_websites":` + totalWebsites + `,
		"success":` + success + `,
		"failure":` + failure + `,
		"total_time":` + totalTime + `
		}`)
	req, _ := http.NewRequest("POST", endPointReport, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Message : 400 , Fail")
		return false
	}

	if resp.StatusCode == http.StatusOK {
		log.Print("Message : 200 OK, Success")
		return true
	} else {
		log.Print("Message : 400 , Fail")
		return false
	}
}

// GenerateReport is Generate Report Health Check
func GenerateReport() (*RequestJSON, error) {
	records := ReadCsvFile("test.csv")
	var websiteArray []string
	for _, items := range records {
		websiteArray = append(websiteArray, items[0])
	}

	success := 0
	failed := 0
	totalTime := int64(0)

	var wg sync.WaitGroup

	for _, web := range websiteArray {
		wg.Add(1)
		go func(web string) {
			defer wg.Done()
			web = strings.TrimSpace(web)
			httpStr := "http://"
			if strings.Contains(web, "https://") {
				httpStr = "https://"
			}
			web1 := strings.ReplaceAll(web, "http://", "")
			web2 := strings.ReplaceAll(web1, "https://", "")
			start := time.Now()
			req, err := http.NewRequest("GET", httpStr+web2, nil)

			if err != nil {
				log.Print(err)
			}
			client := &http.Client{Timeout: 5 * time.Second}
			_, err = client.Do(req)
			if err != nil {
				failed = failed + 1
			} else {
				success = success + 1
			}
			end := time.Since(start)
			totalTime = totalTime + int64(end)
		}(web)
	}
	// Wait Http Check
	wg.Wait()
	requestJSON := RequestJSON{
		TotalWebsites: len(websiteArray),
		Success:       success,
		Failure:       failed,
		TotalTime:     totalTime,
	}
	return &requestJSON, nil
}

func main() {
	viper.SetConfigFile("./configs/env.dev.yaml")
	err := viper.ReadInConfig()

	// Handle errors reading the config file
	if err != nil {
		log.Fatal("Fatal error config file", err)
	}

	log.Print("Perform website checking...")
	requestJSON, err := GenerateReport()
	if err != nil {
		log.Print(err)
	}
	log.Print("Done!")
	log.Print("\n")
	log.Print("Checked webistes: ", requestJSON.TotalWebsites)
	log.Print("Successful websites: ", requestJSON.Success)
	log.Print("Failure websites: ", requestJSON.Failure)
	log.Print("Total times to finished checking website: ", requestJSON.TotalTime)
	token, err := GetAccessToken()
	if err != nil {
		log.Print(err)
	}
	RequestToReportAPI(requestJSON, token)
}
