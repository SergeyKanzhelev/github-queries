package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
)

var access_token = ""

func main() {

	access_token = os.Getenv("ACCESS_TOKEN")

	if len(access_token) == 0 {
		log.Fatal("access_token is needed")
		os.Exit(2)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getColumnID(ctx context.Context, client *github.Client, org string, projectNumber int, columnsName string) (int64, error) {
	projects, _, err := client.Organizations.ListProjects(ctx, org, &github.ProjectListOptions{State: "open", ListOptions: github.ListOptions{Page:1, PerPage: 100} })

	var targetProject *github.Project

	for _, p := range projects {
		//fmt.Printf("Project: %d %s %s %d\n", *p.ID, *p.Name, *p.HTMLURL, *p.Number)
		if *p.Number == projectNumber {
			targetProject = p
			break
		}
	}

	if targetProject == nil {
		fmt.Printf("Project not found")
		return -1, errors.New("Project not found")
	}

	columns, _, err := client.Projects.ListProjectColumns(ctx, *targetProject.ID, &github.ListOptions{Page:1, PerPage: 100})

	if err != nil {
		fmt.Printf("Projects.ListProjectColumns returned error: %v", err)
		return -1, errors.New("Projects.ListProjectColumns returned error")
	}

	fmt.Printf("Project: %s\n", *targetProject.URL)


	var targetColumn *github.ProjectColumn
	for _, c := range columns {
		//fmt.Printf("Column: %d %s\n", *c.ID, *c.Name)

		if *c.Name == columnsName {
			targetColumn = c
		}
	}

	if targetColumn == nil {
		fmt.Printf("Column not found")
		return -1, errors.New("Column not found")
	}

	fmt.Printf("Column: %d %s\n", *targetColumn.ID, *targetColumn.Name)

	return *targetColumn.ID, nil

}

func addIssuesToColumn(ctx context.Context, client *github.Client, query string, columnID int64) error {
	opts := &github.SearchOptions {
		Sort: "forks",
		Order: "desc",
		ListOptions: github.ListOptions{Page: 1, PerPage: 100},
	}

	result, _, err := client.Search.Issues(ctx, query, opts)
	if err != nil {
		fmt.Printf("Search.Issues returned error: %v", err)
		return err
	}

	for _, issue := range result.Issues {
		fmt.Printf("Issue: %d %s %s %d\n", *issue.ID, *issue.NodeID, *issue.Title, *issue.Number)


		if err != nil {
			fmt.Printf("Organizations.ListProjects returned error: %v", err)
			return err
		}

		input := &github.ProjectCardOptions{
			ContentID:   *issue.ID,
			ContentType: "Issue",
		}

		card, resp, err := client.Projects.CreateProjectCard(ctx, columnID, input)

		if err != nil {
			fmt.Printf("Projects.CreateProjectCard returned error: %v, %q", err, resp)
			return err
		}

		fmt.Printf("Card: %s\n", *card.URL)
	}

	return nil

}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: access_token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	columnID, err := getColumnID(ctx, client, "kubernetes", 43, "Triage")

	if err != nil {
		fmt.Fprintf(w, "something wrong: %q", err)
		return
	}

	addIssuesToColumn(ctx, client, "is:pr is:open label:sig/node -project:kubernetes/43 repo:kubernetes/test-infra", columnID)
	addIssuesToColumn(ctx, client, "is:open label:sig/node+-project:kubernetes/43+repo:kubernetes/test-infra", columnID)
	addIssuesToColumn(ctx, client, "is:open label:sig/node is:pr label:area/test -project:kubernetes/43 repo:kubernetes/kubernetes", columnID)
	addIssuesToColumn(ctx, client, "is:issue is:open label:sig/node  label:area/test -project:kubernetes/43 repo:kubernetes/kubernetes", columnID)
	addIssuesToColumn(ctx, client, "is:open label:sig/node is:pr label:kind/failing-test -project:kubernetes/43 repo:kubernetes/kubernetes", columnID)
	addIssuesToColumn(ctx, client, "is:issue is:open label:sig/node label:kind/failing-test -project:kubernetes/43 repo:kubernetes/kubernetes", columnID)

	columnID, err = getColumnID(ctx, client, "kubernetes", 59, "Triage")

	if err != nil {
		fmt.Fprintf(w, "something wrong: %q", err)
		return
	}

	addIssuesToColumn(ctx, client, "is:open label:sig/node is:issue label:kind/bug org:kubernetes -project:kubernetes/59", columnID)

	columnID, err = getColumnID(ctx, client, "kubernetes", 49, "Triage")

	if err != nil {
		fmt.Fprintf(w, "something wrong: %q", err)
		return
	}

	addIssuesToColumn(ctx, client, "is:open label:sig/node is:pr org:kubernetes -project:kubernetes/49", columnID)

	fmt.Fprintf(w, "Hello, World!\n")
}
