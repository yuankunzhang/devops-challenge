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

func init() {
	summaryFlag := flag.Bool("s", false, "show summaries")
	errorFlag := flag.Bool("e", false, "show errors")
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
	// ie is the channel that collects input errors.
	in, ie := input(os.Stdin)

	// out is the result output channel.
	// qe is the channel that collects query errors.
	out, qe := query(in)

	// done will be closed once all output are flushed.
	done := output(os.Stdout, out, ie, qe)

	// Wait until result outputted.
	<-done
}

type inputError struct {
	input string
	error error
}

type queryError struct {
	input string
	error error
}

func input(r io.Reader) (<-chan string, <-chan inputError) {
	in := make(chan string)
	errc := make(chan inputError)
	go func() {
		defer close(in)
		defer close(errc)

		uniqueMap := make(map[string]struct{})

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			s := strings.TrimSpace(scanner.Text())

			// Empty?
			if s == "" {
				continue
			}

			// Invalid?
			if len(strings.Split(s, "/")) != 2 {
				errc <- inputError{s, fmt.Errorf("invalid input: should be in format of $orgname/$repo")}
				continue
			}

			// Duplicated?
			if _, ok := uniqueMap[s]; ok {
				continue
			}
			uniqueMap[s] = struct{}{}
			in <- s
		}
	}()
	return in, errc
}

func query(in <-chan string) (<-chan *RepoStats, <-chan queryError) {
	out := make(chan *RepoStats)
	errc := make(chan queryError)

	go func() {
		defer close(out)
		defer close(errc)
		var wg sync.WaitGroup

		client := NewClient(context.Background(), accessToken)

		for s := range in {
			wg.Add(1)
			go func(s string) {
				defer wg.Done()
				fields := strings.Split(s, "/")
			stats, err := client.Query(fields[0], fields[1])
				if err != nil {
					errc <- queryError{s, err}
					return
				}
				out <- stats
			}(s)
		}

		wg.Wait()
	}()
	return out, errc
}

func output(w io.Writer, out <-chan *RepoStats, ie <-chan inputError, qe <-chan queryError) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup

		var total = 0
		var csvRecords [][]string
		var inputErrors []inputError
		var queryErrors []queryError

		wg.Add(3)

		go func() {
			defer wg.Done()
			for o := range out {
				total += 1
				csvRecords = append(csvRecords, o.CsvRecord())
			}
		}()

		go func() {
			defer wg.Done()
			for e := range ie {
				total += 1
				inputErrors = append(inputErrors, e)
			}
		}()

		go func() {
			defer wg.Done()
			for e := range qe {
				total += 1
				queryErrors = append(queryErrors, e)
			}
		}()

		wg.Wait()

		if len(csvRecords) > 0 {
			writer := csv.NewWriter(w)
			writer.Write(CsvHeader())
			writer.WriteAll(csvRecords)
			writer.Flush()
		}

		if showError {
			if len(inputErrors) > 0 {
				fmt.Printf("\n\nInput Errors:\n")
				for _, ie := range inputErrors {
					fmt.Printf("  <%s> %s\n", ie.input, ie.error)
				}
			}

			if len(queryErrors) > 0 {
				fmt.Printf("\n\nQuery Errors:\n")
				for _, ie := range queryErrors {
					fmt.Printf("  <%s> %s\n", ie.input, ie.error)
				}
			}
		}

		if showSummary {
			fmt.Printf("\n\nSummaries:\n")
			fmt.Printf("  Total Unique Inputs (not including empty lines): %d\n", total)
			fmt.Printf("  Succeeded: %d\n", len(csvRecords))
			fmt.Printf("  Failed: %d\n", len(inputErrors) + len(queryErrors))
		}
	}()
	return done
}
