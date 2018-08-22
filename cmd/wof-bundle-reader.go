package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"github.com/whosonfirst/go-whosonfirst-readwrite-bundle"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"log"
	"os"
	"strconv"
)

func main() {

	var dsn_flags flags.MultiDSNString
	flag.Var(&dsn_flags, "dsn", "DSN strings MUST contain a 'reader=SOURCE' pair followed by any additional pairs required by that reader. Supported reader sources are: fs, http, mysql, s3, sqlite.")

	flag.Parse()

	r, err := bundle.NewMultiReaderFromFlags(dsn_flags)

	if err != nil {
		log.Fatal(err)
	}

	for _, str_id := range flag.Args() {

		id, err := strconv.ParseInt(str_id, 10, 64)

		if err != nil {
			log.Fatal(err)
		}

		path, err := uri.Id2RelPath(id)

		if err != nil {
			log.Fatal(err)
		}

		fh, err := r.Read(path)

		if err != nil {
			log.Fatal(err)
		}

		defer fh.Close()

		io.Copy(os.Stdout, fh)
	}
}
