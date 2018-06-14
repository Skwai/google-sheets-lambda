package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const sheetBaseURL string = "https://spreadsheets.google.com/feeds/list/%s/od6/public/values?alt=json"
const colPrefix string = "gsx$"

type googleSheetsText struct {
	Text string `json:"$t"`
}

type googleSheetsRow map[string]interface{}

type googleSheetsFeed struct {
	ID      googleSheetsText  `json:"id"`
	Updated googleSheetsText  `json:"updated"`
	Entry   []googleSheetsRow `json:"entry"`
}

type googleSheetsResponse struct {
	Version  string `json:"version"`
	Encoding string `json:"encoding"`
	Title    struct {
		Type string `json:"type"`
		googleSheetsText
	}
	Feed googleSheetsFeed `json:"feed"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sheet := request.QueryStringParameters["sheet"]
	response := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}

	if sheet == "" {
		response.Body = "Required query parameter 'sheet' is missing"
		response.StatusCode = 422

		return response, nil
	}

	data, err := getSheetDataFromAPI(sheet)

	if err != nil {
		response.Body = "There was an error retrieving data from sheet"
		response.StatusCode = 400

		return response, nil
	}

	rows := mapRows(data.Feed.Entry)
	json, err := json.Marshal(rows)

	if err != nil {
		response.Body = "There was an error parsing data from sheet"
		response.StatusCode = 400

		return response, err
	}

	response.Body = string(json)
	response.StatusCode = 200

	return response, nil
}

func main() {
	lambda.Start(handler)
}

func mapRows(rows []googleSheetsRow) []map[string]interface{} {
	var mappedRows []map[string]interface{}

	for _, v := range rows {
		mappedRows = append(mappedRows, mapRow(v))
	}

	return mappedRows
}

func mapRow(row googleSheetsRow) map[string]interface{} {
	data := make(map[string]interface{})

	for k, v := range row {
		if strings.Contains(k, colPrefix) {
			value := v.(map[string]interface{})["$t"]

			if value != nil {
				key := strings.Replace(k, colPrefix, "", -1)
				data[key] = value
			}
		}
	}

	return data
}

func getSheetDataFromAPI(sheetID string) (googleSheetsResponse, error) {
	var data googleSheetsResponse

	url := fmt.Sprintf(sheetBaseURL, sheetID)
	response, err := http.Get(url)

	if err != nil {
		return data, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return data, err
	}

	err = json.Unmarshal(body, &data)

	return data, err
}
