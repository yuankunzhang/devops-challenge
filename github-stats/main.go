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
	showSummary bool
	showError   bool
)

type inputError struct {
	input string
	error error
}

func init() {
	summaryFlag := flag.Bool("summary", false, "show statistical summary")
	errorFlag := flag.Bool("error", false, "show errors")
	flag.Parse()

	accessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	if accessToken == "" {
		panic(fmt.Errorf("GITHUB_ACCESS_TOKEN not set"))
	}

	showSummary = *summaryFlag
	showError = *errorFlag
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

func query(in <-chan string) (<-chan *RepoStats, <-chan *inputError) {
	out := make(chan *RepoStats)
	errc := make(chan *inputError)

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
					errc <- &inputError{s, fmt.Errorf("invalid input: Should be in format of $orgname/$repo")}
					return
				}

				stats, err := client.Query(fields[0], fields[1])
				if err != nil {
					errc <- &inputError{s, err}
					return
				}
				out <- stats
			}(s)
		}

		wg.Wait()
	}()
	return out, errc
}

func output(w io.Writer, out <-chan *RepoStats, errc <-chan *inputError) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup

		writer := csv.NewWriter(w)
		var total = 0
		var records [][]string
		var inputErrors []inputError

		wg.Add(1)
		go func() {
			defer wg.Done()
			for o := range out {
				total += 1
				records = append(records, o.CsvRecord())
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range errc {
				total += 1
				inputErrors = append(inputErrors, *e)
			}
		}()

		wg.Wait()

		if len(records) > 0 {
			writer.Write(CsvHeader())
			writer.WriteAll(records)
			writer.Flush()
		}

		if showError {
			fmt.Printf("\n\nErrors:\n")
			for _, ie := range inputErrors {
				fmt.Printf("  <%s> %s\n", ie.input, ie.error)
			}
		}

		if showSummary {
			fmt.Printf("\n\nSummary:\n")
			fmt.Printf("  Total Inputs (not including empty lines): %d\n  Succeeded: %d\n  Failed: %d\n", total, len(records), len(inputErrors))
		}
	}()
	return done
}
