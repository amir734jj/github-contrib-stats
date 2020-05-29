package main

import (
	"context"
	_ "context"
	underscore "github.com/ahl5esoft/golang-underscore"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
	_ "os"
	"sort"
)

type Pair struct {
	a, b interface{}
}

func main() {
	username := "amir734jj"
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_API_KEY")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	//opt := &github.RepositoryListOptions{Type: "public", ListOptions: github.ListOptions{PerPage: 200}}
	repos, _, _ := client.Repositories.List(ctx, username, nil)

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

		return Pair{repo.Name, commits}
	}).Value(&result)

	sort.Slice(result, func(i, j int) bool {
		c1 := result[i].b.(int)
		c2 := result[j].b.(int)
		return c1 > c2
	})

	println(result)
}
