package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"io"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

type prs struct {
	TotalCount int `json:"total_count"`
}

type column struct {
	ColumnName string
	Labels     string
}

var apiRequestsCount = 0

func getPRsCount(query string) (int, error) {
	q := url.Values{}
	q.Add("q", query)
	q.Add("per_page", "1")

	resp, err := http.Get("https://api.github.com/search/issues?" + q.Encode())

	if apiRequestsCount == 9 {
		time.Sleep(1 * time.Minute)
		apiRequestsCount = 0
	} else {
		apiRequestsCount += 1
	}

	if err != nil {
		return -1, fmt.Errorf("failed to get PRs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return -1, fmt.Errorf("status code is not 200: %v, %v", resp.StatusCode, string(b))
	}

	var result = prs{}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Errorf("Failed to parse JSON PRs: %v", err)
		return -1, err
	}

	return result.TotalCount, nil
}

func getPRs() ([]interface{}, error) {
	// see documentation
	// https://developer.github.com/v3/search/#search-issues-and-pull-requests
	// https://docs.github.com/en/github/searching-for-information-on-github/searching-issues-and-pull-requests

	baseQuery := "repo:kubernetes/kubernetes type:pr is:open label:sig/node "
	baseMasterQuery := baseQuery + "base:master "

	columns := []column{
		column{"total", baseQuery},
		// column{"kind api-change", baseMasterQuery + "label:kind/api-change"},
		column{"kind bug", baseMasterQuery + "label:kind/bug"},
		column{"kind cleanup", baseMasterQuery + "label:kind/cleanup"},
		column{"kind deprecation", baseMasterQuery + "label:kind/deprecation"},
		column{"kind design", baseMasterQuery + "label:kind/design"},
		column{"kind documentation", baseMasterQuery + "label:kind/documentation"},
		column{"kind failing-test", baseMasterQuery + "label:kind/failing-test"},
		column{"kind feature", baseMasterQuery + "label:kind/feature"},
		column{"other", baseMasterQuery + "-label:kind/bug -label:kind/cleanup -label:kind/deprecation -label:kind/design -label:kind/documentation -label:kind/failing-test -label:kind/feature"},
		column{"cherry picks", baseQuery + "-base:master"},
	}

	// shrug: " -label:¯\\_(ツ)_/¯ "

	header := "time"
	result := []interface{}{}
	result = append(result, fmt.Sprintf("%s", time.Now().Format("01/02/2006 15:04")))
	//result = append(result, time.Now())
	for _, v := range columns {
		count, err := getPRsCount(v.Labels)
		if err != nil {
			return nil, fmt.Errorf("error for query %s: %v", v.Labels, err)
		}
		result = append(result, count)
		header += fmt.Sprintf(", \"%s\"", v.ColumnName)

		// q := url.Values{}
		// q.Add("q", v.Labels)
		// fmt.Printf("\"%s\", \"https://github.com/kubernetes/kubernetes/pulls?%s\"\n", v.ColumnName, q.Encode())
	}

	return result, nil
}


func getBugs() ([]interface{}, error) {
	// see documentation
	// https://developer.github.com/v3/search/#search-issues-and-pull-requests
	// https://docs.github.com/en/github/searching-for-information-on-github/searching-issues-and-pull-requests

	baseQuery := "repo:kubernetes/kubernetes is:issue is:open label:sig/node "

	var dateNow = time.Now().UTC()

	var dateNowStr = dateNow.Format("2006-01-02T15:04:05-0700")
	var dateRange1day = dateNow.AddDate(0, 0, -1).UTC().Format("2006-01-02T15:04:05-0700") + ".." + dateNowStr
	var dateRange10days = dateNow.AddDate(0, 0, -10).UTC().Format("2006-01-02T15:04:05-0700") + ".." + dateNowStr
	var dateRange90days = dateNow.AddDate(0, 0, -90).UTC().Format("2006-01-02T15:04:05-0700") + ".." + dateNowStr
	var dateOver90days = dateNow.AddDate(0, 0, -90).UTC().Format("2006-01-02T15:04:05-0700")

	columns := []column{
		column{"total", baseQuery},
		column{"kind bug", baseQuery + "label:kind/bug"},
		column{"kind cleanup", baseQuery + "label:kind/cleanup"},
		column{"kind deprecation", baseQuery + "label:kind/deprecation"},
		column{"kind documentation", baseQuery + "label:kind/documentation"},
		column{"kind failing-test", baseQuery + "label:kind/failing-test"},
		column{"kind feature", baseQuery + "label:kind/feature"},
		column{"kind support", baseQuery + "label:kind/support"},
		column{"kind flake", baseQuery + "label:kind/flake"},
		column{"kind other", baseQuery + "-label:kind/bug -label:kind/cleanup -label:kind/deprecation -label:kind/design -label:kind/documentation -label:kind/failing-test -label:kind/feature -label:kind/support -label:kind/flake"},

		column{"updated last day", baseQuery + "updated:" + dateRange1day},
		column{"updated last 10 days", baseQuery + "updated:" + dateRange10days},
		column{"updated last 90 days", baseQuery + "updated:" + dateRange90days},
		column{"updated over 90 days", baseQuery + "updated:<" + dateOver90days},
	}

	header := "time"
	result := []interface{}{}
	result = append(result, fmt.Sprintf("%s", time.Now().Format("01/02/2006 15:04")))
	for _, v := range columns {
		count, err := getPRsCount(v.Labels)
		if err != nil {
			return nil, fmt.Errorf("error for query %s: %v", v.Labels, err)
		}
		result = append(result, count)
		header += fmt.Sprintf(", \"%s\"", v.ColumnName)
	}

	return result, nil
}


func writeToSheet(values []interface{}, sheet string) error {
	// Service account based oauth2 two legged integration
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"), option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	// https://docs.google.com/spreadsheets/d/1VW5_Eq8MzswfDi9xEvfYyP8edF_Ny7MBANIsJXT3VGw/edit
	spreadsheetId := "1VW5_Eq8MzswfDi9xEvfYyP8edF_Ny7MBANIsJXT3VGw"
	readRange := sheet + "!A2:K"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	writeRange := fmt.Sprintf(sheet + "!A%d", len(resp.Values)+2)

	var vr sheets.ValueRange

	vr.Values = append(vr.Values, values)

	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, &vr).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to write data to sheet: %v", err)
	}
	return nil
}

func main() {

	results, err := getPRs()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = writeToSheet(results, "Sheet1")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%v\n", results)

	bugs, err := getBugs()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = writeToSheet(bugs, "Bugs")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%v\n", bugs)
}
