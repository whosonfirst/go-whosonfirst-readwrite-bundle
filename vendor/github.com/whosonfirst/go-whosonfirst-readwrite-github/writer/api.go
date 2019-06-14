package writer

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	wof_writer "github.com/whosonfirst/go-whosonfirst-readwrite/writer"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	_ "log"
	"time"
)

type GitHubAPIWriterCommitTemplates struct {
	New    string
	Update string
}

type GitHubAPIWriter struct {
	wof_writer.Writer
	owner     string
	repo      string
	branch    string
	client    *github.Client
	context   context.Context
	user      *github.User
	throttle  <-chan time.Time
	templates *GitHubAPIWriterCommitTemplates
}

func NewGitHubAPIWriter(ctx context.Context, owner string, repo string, branch string, token string, templates *GitHubAPIWriterCommitTemplates) (wof_writer.Writer, error) {

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	users := client.Users
	user, _, err := users.Get(ctx, "")

	if err != nil {
		return nil, err
	}

	// https://github.com/golang/go/wiki/RateLimiting

	rate := time.Second / 3
	throttle := time.Tick(rate)

	r := GitHubAPIWriter{
		repo:      repo,
		owner:     owner,
		branch:    branch,
		throttle:  throttle,
		client:    client,
		user:      user,
		templates: templates,
		context:   ctx,
	}

	return &r, nil
}

func (r *GitHubAPIWriter) Write(path string, fh io.ReadCloser) error {

	<-r.throttle

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return err
	}

	url := r.URI(path)

	commit_msg := fmt.Sprintf(r.templates.New, url)
	name := *r.user.Login
	email := fmt.Sprintf("%s@localhost", name)

	update_opts := &github.RepositoryContentFileOptions{
		Message: github.String(commit_msg),
		Content: body,
		Branch:  github.String(r.branch),
		Committer: &github.CommitAuthor{
			Name:  github.String(name),
			Email: github.String(email),
		},
	}

	get_opts := &github.RepositoryContentGetOptions{}

	get_rsp, _, _, err := r.client.Repositories.GetContents(r.context, r.owner, r.repo, url, get_opts)

	if err == nil {
		commit_msg = fmt.Sprintf(r.templates.Update, url)
		update_opts.Message = github.String(commit_msg)
		update_opts.SHA = get_rsp.SHA
	}

	_, _, err = r.client.Repositories.UpdateFile(r.context, r.owner, r.repo, url, update_opts)

	if err != nil {
		return err
	}

	return nil
}

func (r *GitHubAPIWriter) URI(key string) string {

	return fmt.Sprintf("data/%s", key)
}
