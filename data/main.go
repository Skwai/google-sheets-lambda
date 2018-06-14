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

// GoogleSheetsText is the content of a single Google Sheets cell
type GoogleSheetsText struct {
	Text string `json:"$t"`
}

// GoogleSheetsRow represents a single row in a Google Sheet
type GoogleSheetsRow map[string]interface{}

// GoogleSheetsFeed is the structure of the data feed returned from Google Sheets
type GoogleSheetsFeed struct {
	ID      GoogleSheetsText  `json:"id"`
	Updated GoogleSheetsText  `json:"updated"`
	Entry   []GoogleSheetsRow `json:"entry"`
}

// GoogleSheetsResponse is the structure of the data returned from Google Sheets
type GoogleSheetsResponse struct {
	Version  string `json:"version"`
	Encoding string `json:"encoding"`
	Title    struct {
		Type string `json:"type"`
		GoogleSheetsText
	}
	Feed GoogleSheetsFeed `json:"feed"`
}

// SheetsErrorResponse is the structure of a sheets error response
type SheetsErrorResponse struct {
	Error SheetsError `json:"error"`
}

// SheetsError is the structure of a sheets error
type SheetsError struct {
	Message string `json:"message"`
}

// Handler responds to the lamba
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sheet := request.QueryStringParameters["sheet"]
	response := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}
	if sheet == "" {
		e := SheetsErrorResponse{Error: SheetsError{Message: "Required query parameter 'sheet' is missing"}}
		j, _ := json.Marshal(e)
		response.Body = string(j)
		response.StatusCode = 422

		return response, nil
	}
	data, err := GetSheetDataFromAPI(sheet)
	if err != nil {
		e := SheetsErrorResponse{Error: SheetsError{Message: "There was an error retrieving data from sheet"}}
		j, _ := json.Marshal(e)
		response.Body = string(j)
		response.StatusCode = 400

		return response, nil
	}
	rows := MapRows(data.Feed.Entry)
	j, err := json.Marshal(rows)
	if err != nil {
		e := SheetsErrorResponse{Error: SheetsError{Message: "There was an error parsing data from sheet"}}
		j, _ := json.Marshal(e)
		response.Body = string(j)
		response.StatusCode = 400

		return response, err
	}
	response.Body = string(j)
	response.StatusCode = 200

	return response, nil
}

func main() {
	lambda.Start(Handler)
}

// MapRows maps all of the rows in the Google Sheet data
func MapRows(rows []GoogleSheetsRow) []map[string]interface{} {
	var mappedRows []map[string]interface{}
	for _, v := range rows {
		mappedRows = append(mappedRows, MapRow(v))
	}

	return mappedRows
}

// MapRow maps a single Google Sheet row to an interface with
func MapRow(row GoogleSheetsRow) map[string]interface{} {
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

// GetSheetDataFromAPI gets the JSON data from a Google Sheet
func GetSheetDataFromAPI(sheetID string) (GoogleSheetsResponse, error) {
	var data GoogleSheetsResponse
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
