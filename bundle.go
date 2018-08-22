package bundle

import (
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	// github_reader "github.com/whosonfirst/go-whosonfirst-readwrite-github/reader"
	http_reader "github.com/whosonfirst/go-whosonfirst-readwrite-http/reader"
	mysql_reader "github.com/whosonfirst/go-whosonfirst-readwrite-mysql/reader"
	s3_config "github.com/whosonfirst/go-whosonfirst-readwrite-s3/config"
	s3_reader "github.com/whosonfirst/go-whosonfirst-readwrite-s3/reader"
	sqlite_reader "github.com/whosonfirst/go-whosonfirst-readwrite-sqlite/reader"
	"github.com/whosonfirst/go-whosonfirst-readwrite/reader"
	"strings"
)

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
			r, e = newFSReader(dsn)
		case "HTTP":
			r, e = newFSReader(dsn)
		case "MYSQL":
			r, e = newMySQLReader(dsn)
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