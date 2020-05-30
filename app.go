package main

import (
	"context"
	_ "context"
	"encoding/json"
	_ "encoding/json"
	underscore "github.com/ahl5esoft/golang-underscore"
	"github.com/google/go-github/github"
	"github.com/hoisie/web"
	_ "github.com/hoisie/web"
	"golang.org/x/oauth2"
	"os"
	_ "os"
	"sort"
	"time"
	_ "time"
)

type Pair struct {
	a, b interface{}
}

type Stats struct {
	Repo        string `json:"repo"`
	Commits     int    `json:"commits"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

func stats(ctx context.Context) func(web *web.Context, username string) {
	return func(web *web.Context, username string) {
		deadline := time.Now().Add(10 * time.Second)
		child, _ := context.WithDeadline(ctx, deadline)
		client := client(child)

		opt := &github.RepositoryListOptions{Type: "public", ListOptions: github.ListOptions{PerPage: 200}}
		repos, _, _ := client.Repositories.List(ctx, username, opt)

		result := make([]Pair, 0)

		underscore.Chain(repos).Where(func(repo github.Repository, _ int) bool {
			return !*repo.Fork
		}).Select(func(repo github.Repository, _ int) Pair {
			cont, _, _ := client.Repositories.ListContributors(ctx, username, *repo.Name, nil)

			return Pair{repo, cont}
		}).Select(func(pair Pair, _ int) Pair {
			repo := pair.a.(github.Repository)
			commits := 0
			underscore.Chain(pair.b.([]*github.Contributor)).Select(func(cont github.Contributor, _ int) int {
				return *cont.Contributions
			}).Aggregate(0, func(memo int, n, _ int) int {
				return memo + n
			}).Value(&commits)

			return Pair{repo, commits}
		}).Value(&result)

		sort.Slice(result, func(i, j int) bool {
			c1 := result[i].b.(int)
			c2 := result[j].b.(int)
			return c1 > c2
		})

		response := make([]Stats, 0)
		underscore.Chain(result).Select(func(pair Pair, _ int) Stats {
			repo := pair.a.(github.Repository)
			commit := pair.b.(int)
			return Stats{Repo: *repo.Name, Commits: commit, Url: *repo.HTMLURL, Description: *repo.Description}
		}).Value(&response)

		_ = json.NewEncoder(web.ResponseWriter).Encode(response)
	}
}

func client(ctx context.Context) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_API_KEY")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client
}

func main() {
	ctx := context.Background()

	web.Get("/(.*)", stats(ctx))
	web.Run("0.0.0.0:5000")
}
