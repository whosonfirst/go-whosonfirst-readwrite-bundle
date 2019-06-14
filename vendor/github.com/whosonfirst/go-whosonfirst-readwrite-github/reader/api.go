package reader

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	wof_reader "github.com/whosonfirst/go-whosonfirst-readwrite/reader"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	_ "log"
	"strings"
	"time"
)

type GitHubAPIReader struct {
	wof_reader.Reader
	owner    string
	repo     string
	branch   string
	client   *github.Client
	context  context.Context
	throttle <-chan time.Time
}

func NewGitHubAPIReader(ctx context.Context, owner string, repo string, branch string, token string) (wof_reader.Reader, error) {

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// https://github.com/golang/go/wiki/RateLimiting

	rate := time.Second / 3
	throttle := time.Tick(rate)

	r := GitHubAPIReader{
		repo:     repo,
		owner:    owner,
		branch:   branch,
		throttle: throttle,
		client:   client,
		context:  ctx,
	}

	return &r, nil
}

func (r *GitHubAPIReader) Read(key string) (io.ReadCloser, error) {

	<-r.throttle

	url := r.URI(key)

	opts := &github.RepositoryContentGetOptions{}

	rsp, _, _, err := r.client.Repositories.GetContents(r.context, r.owner, r.repo, url, opts)

	if err != nil {
		return nil, err
	}

	body, err := rsp.GetContent()

	if err != nil {
		return nil, err
	}

	br := strings.NewReader(body)
	fh := ioutil.NopCloser(br)

	return fh, nil
}

func (r *GitHubAPIReader) URI(key string) string {

	return fmt.Sprintf("data/%s", key)
}
