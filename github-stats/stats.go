package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	graphqlURL  = "https://api.github.com/graphql"
	contentType = "application/json"
)

// Client manages communications with the Github GraphQL API.
type Client struct {
	httpClient    *http.Client
	queryTemplate *template.Template
}

// RepoStats represents the repository information that we are interested in.
type RepoStats struct {
	Name       string
	URL        string
	CommitDate string
	AuthorName string
}

// CsvRecords converts the RepoStats object to a valid csv record, which is
// actually a string array.
func (stats *RepoStats) CsvRecord() []string {
	return []string{
		stats.Name,
		stats.URL,
		stats.CommitDate,
		stats.AuthorName,
	}
}

// CsvHeader returns a string array which represents the header record of a
// list of csv records.
func CsvHeader() []string {
	return []string{"Name", "Clone URL", "Date of Latest Commit", "Name of Latest Author"}
}

// NewClient returns a new Github GraphQL API client.
func NewClient(ctx context.Context, accessToken string) *Client {
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	httpClient := oauth2.NewClient(ctx, src)

	tmpl, _ := template.New("query").Parse(queryTemplate)

	return &Client{
		httpClient:    httpClient,
		queryTemplate: tmpl,
	}
}

const queryTemplate = `query{
	repository(owner: "{{ .Owner }}", name: "{{ .Name }}") {
		name
		url
		defaultBranchRef {
			target {
				... on Commit {
          			history(first: 1) {
            			edges {
              				node {
                				message
                				author {
                  					name
                  					date
								}
              				}
						}
          			}
        		}
			}
		}
	}
	rateLimit {
    	limit
    	cost
    	remaining
    	resetAt
  	}
}`

// QueryResult receives the result returns from Github GraphQL API after
// a GraphQL query is issued.
type QueryResult struct {
	Data struct {
		Repository struct {
			Name string `json:"name"`
			URL  string `json:"url"`
			DefaultBranchRef  struct {
				Target struct {
					History struct {
						Edges []struct {
							Node struct {
								Message string `json:"message"`
								Author  struct {
									Name string `json:"name"`
									Date string `json:"date"`
								} `json:"author"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"history"`
				} `json:"target"`
			} `json:"defaultBranchRef"`
		} `json:"repository"`
		RateLimit struct {
			Limit     int       `json:"limit"`
			Cost      int       `json:"cost"`
			Remaining int       `json:"remaining"`
			ResetAt   time.Time `json:"resetAt"`
		} `json:"rateLimit"`
	} `json:"data"`
	Errors []struct {
		Message   string
		Locations []struct {
			Line   int
			Column int
		}
	} `json:"errors"`
}

// Query queries repository information for the given owner & name pair.
func (client *Client) Query(owner, name string) (*RepoStats, error) {
	var queryStmt bytes.Buffer
	err := client.queryTemplate.Execute(&queryStmt, &struct {
		Owner  string
		Name   string
		Branch string
	}{
		Owner:  owner,
		Name:   name,
		Branch: "master",
	})
	if err != nil {
		return nil, err
	}

	in := struct {
		Query string `json:"query"`
	}{
		Query: queryStmt.String(),
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return nil, errors.Wrap(err, "json encode failed")
	}

	resp, err := client.httpClient.Post(graphqlURL, contentType, &buf)
	if err != nil {
		return nil, errors.Wrap(err, "post request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %v", resp.Status)
	}

	var out QueryResult
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil && err != io.EOF {
		return nil, errors.Wrap(err, "json decode failed")
	}

	if len(out.Errors) > 0 {
		// Return message of the first error.
		return nil, errors.Errorf("query error: %v", out.Errors[0].Message)
	}

	if len(out.Data.Repository.DefaultBranchRef.Target.History.Edges) == 0 {
		return nil, errors.Errorf("query error: empty commit history")
	}

	repo := out.Data.Repository
	author := repo.DefaultBranchRef.Target.History.Edges[0].Node.Author

	return &RepoStats{
		Name:       repo.Name,
		URL:        repo.URL,
		CommitDate: author.Date,
		AuthorName: author.Name,
	}, nil
}
