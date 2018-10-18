package main

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-github/organizations"
	"github.com/whosonfirst/iso8601duration"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"
)

func main() {

	org := flag.String("org", "whosonfirst-data", "The name of the organization to clone repositories from")
	prefix := flag.String("prefix", "whosonfirst-data", "Limit repositories to only those with this prefix")
	exclude := flag.String("exclude", "", "Exclude repositories with this prefix")
	updated_since := flag.String("updated-since", "", "A valid Unix timestamp or an ISO8601 duration string (months are currently not supported)")
	forked := flag.Bool("forked", false, "Only include repositories that have been forked")
	not_forked := flag.Bool("not-forked", false, "Only include repositories that have not been forked")
	token := flag.String("token", "", "A valid GitHub API access token")

	flag.Parse()

	opts := organizations.NewDefaultListOptions()

	opts.Prefix = *prefix
	opts.Exclude = *exclude
	opts.Forked = *forked
	opts.NotForked = *not_forked
	opts.AccessToken = *token

	if *updated_since != "" {

		var since time.Time

		is_timestamp, err := regexp.MatchString("^\\d+$", *updated_since)

		if err != nil {
			log.Fatal(err)
		}

		if is_timestamp {

			ts, err := strconv.Atoi(*updated_since)

			if err != nil {
				log.Fatal(err)
			}

			now := time.Now()

			tm := time.Unix(int64(ts), 0)
			since = now.Add(-time.Since(tm))

		} else {

			// maybe also this https://github.com/araddon/dateparse ?

			d, err := duration.FromString(*updated_since)

			if err != nil {
				log.Fatal(err)
			}

			now := time.Now()
			since = now.Add(-d.ToDuration())
		}

		// log.Printf("SINCE %v\n", since)
		// os.Exit(0)

		opts.PushedSince = &since
	}

	repos, err := organizations.ListRepos(*org, opts)

	if err != nil {
		log.Fatal(err)
	}

	for _, name := range repos {
		fmt.Println(name)
	}

	os.Exit(0)
}
