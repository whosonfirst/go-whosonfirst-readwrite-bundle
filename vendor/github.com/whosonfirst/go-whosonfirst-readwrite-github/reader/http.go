package reader

// maybe also make a GH API reader...
// https://developer.github.com/v3/repos/contents/#get-contents

import (
	"errors"
	"fmt"
	wof_reader "github.com/whosonfirst/go-whosonfirst-readwrite/reader"
	"io"
	_ "log"
	"net/http"
	"time"
)

type GitHubReader struct {
	wof_reader.Reader
	repo     string
	branch   string
	throttle <-chan time.Time
}

func NewGitHubReader(repo string, branch string) (wof_reader.Reader, error) {

	// https://github.com/golang/go/wiki/RateLimiting

	rate := time.Second / 3
	throttle := time.Tick(rate)

	r := GitHubReader{
		repo:     repo,
		branch:   branch,
		throttle: throttle,
	}

	return &r, nil
}

func (r *GitHubReader) Read(key string) (io.ReadCloser, error) {

	<-r.throttle

	url := r.URI(key)

	// log.Println("READ", key, url)

	rsp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != 200 {
		return nil, errors.New(rsp.Status)
	}

	return rsp.Body, nil
}

func (r *GitHubReader) URI(key string) string {

	return fmt.Sprintf("https://raw.githubusercontent.com/whosonfirst-data/%s/%s/data/%s", r.repo, r.branch, key)
}
