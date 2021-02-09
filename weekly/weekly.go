package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
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

func getPRsCount(query string) (int, error) {
	q := url.Values{}
	q.Add("q", query)
	q.Add("per_page", "1")

	resp, err := http.Get("https://api.github.com/search/issues?" + q.Encode())

	if err != nil {
		return -1, fmt.Errorf("failed to get PRs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return -1, fmt.Errorf("status code is not 200: %v", resp.StatusCode)
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

	var dateNow = time.Now().UTC()

	var lastMeeting = dateNow

	if lastMeeting.Weekday() == time.Tuesday && lastMeeting.Hour() < 17 {
		lastMeeting = lastMeeting.AddDate(0, 0, -7).UTC()
	} else {
		for lastMeeting.Weekday() != time.Tuesday {
			lastMeeting = lastMeeting.AddDate(0, 0, -1).UTC()
		}
	}

	year, month, day := lastMeeting.Date()
	lastMeeting = time.Date(year, month, day, 17, 0, 0, 0, time.UTC)

	// if meeting has already started, uncomment below:
	// dateNow = lastMeeting
	lastMeeting = lastMeeting.AddDate(0, 0, -7).UTC()
	//lastMeeting = lastMeeting.AddDate(0, 0, -7).UTC()
	//lastMeeting = lastMeeting.AddDate(0, 0, -7).UTC()
	//lastMeeting = lastMeeting.AddDate(0, 0, -7).UTC()

	baseQuery := "repo:kubernetes/kubernetes type:pr label:sig/node "

	var dateNowStr = dateNow.Format("2006-01-02T15:04:05-0700")
	var lastMeetingDateStr = lastMeeting.Format("2006-01-02T15:04:05-0700")
	var dateRange = lastMeetingDateStr + ".." + dateNowStr

	columns := []column{
		column{"total", baseQuery + "is:open "},
		column{"created", baseQuery + " created:" + dateRange},
		column{"updated", baseQuery + "updated:" + dateRange + " created:<" + lastMeetingDateStr},
		column{"closed", baseQuery + " is:unmerged closed:" + dateRange},
		column{"merged", baseQuery + " merged:" + dateRange},
	}

	// shrug: " -label:¯\\_(ツ)_/¯ "

	header := "from"
	result := []interface{}{}
	result = append(result, lastMeetingDateStr)

	header += "time"
	result = append(result, dateNowStr)
	for _, v := range columns {
		count, err := getPRsCount(v.Labels)
		if err != nil {
			return nil, fmt.Errorf("error for query %s: %v", v.Labels, err)
		}
		//=HYPERLINK("https://github.com/kubernetes/kubernetes/pulls?q=repo%3Akubernetes%2Fkubernetes+type%3Apr+label%3Asig%2Fnode++created%3A%3E%3D2020-08-04T17%3A00%3A00%2B0000", "created")

		q := url.Values{}
		q.Add("q", v.Labels)
		urlStr := fmt.Sprintf("https://github.com/kubernetes/kubernetes/pulls?%s", q.Encode())

		var hyperlinkStr = fmt.Sprintf("=HYPERLINK(\"%s\", \"%d\")", urlStr, count)
		result = append(result, hyperlinkStr)
		header += fmt.Sprintf(", \"%s\"", v.ColumnName)

	}

	return result, nil
}

func writeToSheet(values []interface{}) error {
	// Service account based oauth2 two legged integration
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"), option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	// https://docs.google.com/spreadsheets/d/1VW5_Eq8MzswfDi9xEvfYyP8edF_Ny7MBANIsJXT3VGw/edit
	spreadsheetId := "1VW5_Eq8MzswfDi9xEvfYyP8edF_Ny7MBANIsJXT3VGw"
	readRange := "Weekly!A24:G"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	writeRange := fmt.Sprintf("Weekly!A%d", len(resp.Values)+24)

	var vr sheets.ValueRange

	vr.Values = append(vr.Values, values)

	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, &vr).ValueInputOption("USER_ENTERED").Do()
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

	err = writeToSheet(results)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%v\n", results)
}
