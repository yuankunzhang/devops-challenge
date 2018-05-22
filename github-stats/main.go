package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var (
	accessToken string
	showDetail  bool
)

func init() {
	detail := flag.Bool("detail", false, "show detailed statistics")
	flag.Parse()

	accessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	if accessToken == "" {
		panic(fmt.Errorf("GITHUB_ACCESS_TOKEN not set"))
	}

	showDetail = *detail
}

func main() {
	// in is the data input channel.
	in := input(os.Stdin)

	// out is the result output channel.
	// errc collects all errors occurred when querying.
	out, errc := query(in)

	// done will be closed once all output are flushed.
	done := output(os.Stdout, out, errc)

	<-done
}

func input(r io.Reader) <-chan string {
	in := make(chan string)
	go func() {
		defer close(in)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			in <- strings.TrimSpace(scanner.Text())
		}
	}()
	return in
}

func query(in <-chan string) (<-chan *RepoStats, <-chan error) {
	out := make(chan *RepoStats)
	errc := make(chan error)

	go func() {
		defer close(out)
		defer close(errc)
		var wg sync.WaitGroup

		client := NewClient(context.Background(), accessToken)

		for s := range in {
			wg.Add(1)
			go func(s string) {
				defer wg.Done()

				item := strings.TrimSpace(s)

				// Empty?
				if item == "" {
					return
				}

				fields := strings.Split(s, "/")

				// Valid?
				if len(fields) != 2 {
					errc <- fmt.Errorf("invalid input")
					return
				}

				stats, err := client.Query(fields[0], fields[1])
				if err != nil {
					errc <- err
					return
				}
				out <- stats
			}(s)
		}

		wg.Wait()
	}()
	return out, errc
}

func output(w io.Writer, out <-chan *RepoStats, errc <-chan error) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup

		writer := csv.NewWriter(w)
		var records [][]string
		var errors []error

		wg.Add(1)
		go func() {
			defer wg.Done()
			for o := range out {
				records = append(records, o.CsvRecord())
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range errc {
				errors = append(errors, e)
			}
		}()

		wg.Wait()

		if len(records) > 0 {
			writer.Write(CsvHeader())
			writer.WriteAll(records)
			writer.Flush()
		}

		if showDetail {
			fmt.Printf("\n\nQueries: %d\nErrors: %d\n", len(records), len(errors))
		}
	}()
	return done
}
