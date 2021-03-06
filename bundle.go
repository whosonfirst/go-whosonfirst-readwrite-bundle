package bundle

import (
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"github.com/whosonfirst/go-whosonfirst-github/organizations"
	github_reader "github.com/whosonfirst/go-whosonfirst-readwrite-github/reader"
	http_reader "github.com/whosonfirst/go-whosonfirst-readwrite-http/reader"
	mysql_reader "github.com/whosonfirst/go-whosonfirst-readwrite-mysql/reader"
	s3_config "github.com/whosonfirst/go-whosonfirst-readwrite-s3/config"
	s3_reader "github.com/whosonfirst/go-whosonfirst-readwrite-s3/reader"
	s3_writer "github.com/whosonfirst/go-whosonfirst-readwrite-s3/writer"
	sqlite_reader "github.com/whosonfirst/go-whosonfirst-readwrite-sqlite/reader"
	sqlite_writer "github.com/whosonfirst/go-whosonfirst-readwrite-sqlite/writer"
	"github.com/whosonfirst/go-whosonfirst-readwrite/reader"
	"github.com/whosonfirst/go-whosonfirst-readwrite/writer"
	_ "log"
	"path/filepath"
	"strings"
)

func ValidReadersString() string {
	valid := ValidReaders()
	return strings.Join(valid, ", ")
}

func ValidReaders() []string {

	readers := []string{
		"fs",
		"github",
		"http",
		"mysql",
		"repo",
		"s3",
		"sqlite",
	}

	return readers
}

func NewMultiReaderFromFlags(dsn_flags flags.MultiDSNString) (reader.Reader, error) {

	readers := make([]reader.Reader, 0)

	for _, dsn := range dsn_flags {

		source, ok := dsn["reader"]

		if !ok {
			return nil, errors.New("Missing reader key in DSN string")
		}

		var r reader.Reader
		var e error

		switch strings.ToUpper(source) {

		case "FS":

			// something something something filesystem globbing...
			// (20180822/thisisaaronland)

			r, e = newFSReader(dsn)
		case "GITHUB":

			repo, ok := dsn["repo"]

			if !ok {
				return nil, errors.New("Missing repo pair in DSN string")
			}

			if strings.HasSuffix(repo, "*") {

				token, _ := dsn["access_token"]

				opts := organizations.NewDefaultListOptions()
				opts.Prefix = strings.Replace(repo, "*", "", -1)
				opts.AccessToken = token
				opts.NotForked = true

				repos, err := organizations.ListRepos("whosonfirst-data", opts)

				if err != nil {
					return nil, err
				}

				for _, repo := range repos {

					dsn := map[string]string{
						"repo": repo,
					}

					r, e = newGitHubReader(dsn)

					if e != nil {
						return nil, e
					}

					readers = append(readers, r)
				}

				continue

			} else {
				r, e = newGitHubReader(dsn)
			}

		case "HTTP":
			r, e = newFSReader(dsn)
		case "MYSQL":
			r, e = newMySQLReader(dsn)
		case "REPO":
			dsn["root"] = filepath.Join(dsn["root"], "data")
			r, e = newFSReader(dsn)
		case "S3":
			r, e = newS3Reader(dsn)
		case "SQLITE":
			r, e = newSQLiteReader(dsn)
		default:
			return nil, errors.New("Unsupported reader")
		}

		if e != nil {
			return nil, e
		}

		readers = append(readers, r)
	}

	if len(readers) == 0 {
		return nil, errors.New("You forgot to specify any sources")
	}

	return reader.NewMultiReader(readers...)
}

func newFSReader(dsn map[string]string) (reader.Reader, error) {

	root, ok := dsn["root"]

	if !ok {
		return nil, errors.New("FS reader DSN missing a root={PATH} pair")
	}

	return reader.NewFSReader(root)
}

func newGitHubReader(dsn map[string]string) (reader.Reader, error) {

	repo, ok := dsn["repo"]

	if !ok {
		return nil, errors.New("GitHub reader DSN missing a repo={REPO} pair")
	}

	branch := "master"

	_, ok = dsn["branch"]

	if ok {
		branch = dsn["branch"]
	}

	return github_reader.NewGitHubReader(repo, branch)
}

func newHTTPReader(dsn map[string]string) (reader.Reader, error) {

	root, ok := dsn["root"]

	if !ok {
		return nil, errors.New("HTTP reader DSN missing a root={PATH} pair")
	}

	return http_reader.NewHTTPReader(root)
}

func newMySQLReader(dsn map[string]string) (reader.Reader, error) {
	str_dsn := dsnToString(dsn)
	return mysql_reader.NewMySQLGeoJSONReader(str_dsn)
}

func newS3Reader(dsn map[string]string) (reader.Reader, error) {

	str_dsn := dsnToString(dsn)

	cfg, err := s3_config.NewS3ConfigFromString(str_dsn)

	if err != nil {
		return nil, err
	}

	return s3_reader.NewS3Reader(cfg)
}

func newSQLiteReader(dsn map[string]string) (reader.Reader, error) {
	str_dsn := dsnToString(dsn)
	return sqlite_reader.NewSQLiteReader(str_dsn)
}

func dsnToString(dsn map[string]string) string {

	str_dsn := ""

	for k, v := range dsn {
		str_dsn = fmt.Sprintf("%s %s=%v", str_dsn, k, v)
	}

	str_dsn = strings.Trim(str_dsn, " ")
	return str_dsn
}

func ValidWritersString() string {

	valid := ValidWriters()
	return strings.Join(valid, ", ")
}

func ValidWriters() []string {

	writers := []string{
		"fs",
		// "github",
		// "http",
		// "mysql",
		"null",
		"repo",
		"s3",
		"sqlite",
		"stdout",
	}

	return writers
}

func NewMultiWriterFromFlags(dsn_flags flags.MultiDSNString) (writer.Writer, error) {

	writers := make([]writer.Writer, 0)

	for _, dsn := range dsn_flags {

		source, ok := dsn["writer"]

		if !ok {
			return nil, errors.New("Missing writer key in DSN string")
		}

		var w writer.Writer
		var e error

		switch strings.ToUpper(source) {

		case "FS":
			w, e = newFSWriter(dsn)
		case "NULL":
			w, e = writer.NewNullWriter()
		case "REPO":
			dsn["root"] = filepath.Join(dsn["root"], "data")
			w, e = newFSWriter(dsn)
		case "S3":
			w, e = newS3Writer(dsn)
		case "SQLITE":
			w, e = newSQLiteWriter(dsn)
		case "STDOUT":
			w, e = writer.NewStdoutWriter()
		default:
			return nil, errors.New("Unsupported writer")
		}

		if e != nil {
			return nil, e
		}

		writers = append(writers, w)
	}

	if len(writers) == 0 {
		return nil, errors.New("You forgot to specify any sources")
	}

	return writer.NewMultiWriter(writers...)
}

func newFSWriter(dsn map[string]string) (writer.Writer, error) {

	root, ok := dsn["root"]

	if !ok {
		return nil, errors.New("FS writer DSN missing a root={PATH} pair")
	}

	return writer.NewFSWriter(root)
}

func newS3Writer(dsn map[string]string) (writer.Writer, error) {

	str_dsn := dsnToString(dsn)

	cfg, err := s3_config.NewS3ConfigFromString(str_dsn)

	if err != nil {
		return nil, err
	}

	return s3_writer.NewS3Writer(cfg)
}

func newSQLiteWriter(dsn map[string]string) (writer.Writer, error) {
	str_dsn := dsnToString(dsn)
	return sqlite_writer.NewSQLiteWriter(str_dsn)
}
