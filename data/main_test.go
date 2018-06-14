package main

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestMapRows(t *testing.T) {
	var rows []GoogleSheetsRow
	json.Unmarshal([]byte(`
		[
			{
        "id": {
          "$t": "https://spreadsheets.google.com/feeds/list/1jNWarfCxUCnjby0fKPViY9j5h2XTB4k0InK9OWdd1-s/od6/public/values/clrrx"
        },
        "updated": {
          "$t": "2017-11-28T22:35:19.634Z"
        },
        "gsx$state": {
          "$t": "New South Wales"
        },
        "gsx$population": {
          "$t": "7,757,800.00"
        },
        "gsx$percent": {
          "$t": "32.03%"
        },
        "gsx$capital": {
          "$t": "Sydney"
        }
      },
      {
        "id": {
          "$t": "https://spreadsheets.google.com/feeds/list/1jNWarfCxUCnjby0fKPViY9j5h2XTB4k0InK9OWdd1-s/od6/public/values/cyevm"
        },
        "updated": {
          "$t": "2017-11-28T22:35:19.634Z"
        },
        "gsx$state": {
          "$t": "Australian Capital Territory"
        },
        "gsx$population": {
          "$t": "398,300.00"
        },
        "gsx$percent": {
          "$t": "1.64%"
        },
        "gsx$capital": {
          "$t": "Canberra"
        }
      }
		]
	`), &rows)
	result := MapRows(rows)
	assert.Len(t, result, 2)
}

func TestMapRow(t *testing.T) {
	var row GoogleSheetsRow
	json.Unmarshal([]byte(`
		{
			"id": {
				"$t": "https://spreadsheets.google.com/feeds/list/1jNWarfCxUCnjby0fKPViY9j5h2XTB4k0InK9OWdd1-s/od6/public/values/cyevm"
			},
			"updated": {
				"$t": "2017-11-28T22:35:19.634Z"
			},
			"gsx$state": {
				"$t": "Australian Capital Territory"
			},
			"gsx$population": {
				"$t": "398,300.00"
			},
			"gsx$capital": {
				"$t": "Canberra"
			}
		}
	`), &row)
	result := MapRow(row)
	assert.Contains(t, result, "capital")
	assert.Contains(t, result, "population")
	assert.Contains(t, result, "state")
	assert.Len(t, result, 3)
}

func TestHandler(t *testing.T) {
	tests := []struct {
		request events.APIGatewayProxyRequest
		expect  string
		err     error
	}{
		{
			request: events.APIGatewayProxyRequest{QueryStringParameters: nil},
			expect:  "Required query parameter 'sheet' is missing",
			err:     nil,
		},
	}

	for _, test := range tests {
		response, err := Handler(test.request)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.expect, response.Body)
	}
}
