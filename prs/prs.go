package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
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
		fmt.Errorf("Failed to get PRs: %v", err)
		return -1, err
	}
	defer resp.Body.Close()

	var result = prs{}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Errorf("Failed to parse JSON PRs: %v", err)
		return -1, err
	}

	return result.TotalCount, nil
}

func main() {

	// see documentation
	// https://developer.github.com/v3/search/#search-issues-and-pull-requests
	// https://docs.github.com/en/github/searching-for-information-on-github/searching-issues-and-pull-requests

	baseQuery := "repo:kubernetes/kubernetes type:pr is:open label:sig/node "
	baseMasterQuery := baseQuery + "base:master "

	columns := []column{
		column{"total", baseQuery},
		//column{"kind api-change", baseMasterQuery + "label:kind/api-change"},
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
	result := fmt.Sprintf("%s", time.Now().Format("2006-01-02T15:04:05.999999-07:00"))
	for _, v := range columns {
		count, err := getPRsCount(v.Labels)
		if err != nil {
			fmt.Printf("Error for query %s: %v", v.Labels, err)
			os.Exit(1)
		}
		result += fmt.Sprintf(", %d", count)
		header += fmt.Sprintf(", \"%s\"", v.ColumnName)

		q := url.Values{}
		q.Add("q", v.Labels)

		fmt.Printf("\"%s\", \"https://github.com/kubernetes/kubernetes/pulls?%s\"\n", v.ColumnName, q.Encode())
	}

	fmt.Println(header)
	fmt.Println(result)

}
