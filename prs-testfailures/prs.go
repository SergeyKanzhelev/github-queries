package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
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

func getPRs() error {
	// see documentation
	// https://developer.github.com/v3/search/#search-issues-and-pull-requests
	// https://docs.github.com/en/github/searching-for-information-on-github/searching-issues-and-pull-requests

	columns := []column{
		column{"Test-infra sig/node: PRs", "repo:kubernetes/test-infra is:pr is:open label:sig/node "},
		column{"Test-infra sig/node: issues", "repo:kubernetes/test-infra is:issue is:open label:sig/node "},
		column{"k/k sig node area/test: PRs", "repo:kubernetes/kubernetes is:open label:sig/node label:area/test is:pr "},
		column{"k/k sig node area/test: PRs (approved)", "repo:kubernetes/kubernetes is:open label:sig/node label:area/test is:pr label:approved"},
		column{"k/k sig node area/test: issues", "repo:kubernetes/kubernetes is:open label:sig/node label:area/test is:issue "},
		column{"k/k sig node kind/failing-test: PRs", "repo:kubernetes/kubernetes is:open label:sig/node is:pr label:kind/failing-test "},
		column{"k/k sig node kind/failing-test: PRs (approved)", "repo:kubernetes/kubernetes is:open label:sig/node is:pr label:kind/failing-test label:approved"},
		column{"k/k sig node kind/failing-test", "repo:kubernetes/kubernetes is:open label:sig/node is:issue label:kind/failing-test "},
	}

	for _, v := range columns {
		count, err := getPRsCount(v.Labels)
		if err != nil {
			return fmt.Errorf("error for query %s: %v", v.Labels, err)
		}

		q := url.Values{}
		q.Add("q", v.Labels)
		query := fmt.Sprintf("https://github.com/issues?%s", q.Encode())

		fmt.Printf("- %s: [%d](%s)\n", v.ColumnName, count, query)
	}

	return nil
}

func main() {

	err := getPRs()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

}
