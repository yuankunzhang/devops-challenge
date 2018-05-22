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

var accessToken string

func init() {
	token := flag.String("token", "", "the access token used when querying graphql")
	flag.Parse()

	if *token == "" {
		panic(fmt.Errorf("please provide a token"))
	}

	accessToken = *token
}

func main() {
	// in is the data input channel.
	in := input(os.Stdin)

	// out is the result output channel.
	out, _ := query(in)

	// done will be closed once all output are flushed.
	done := output(os.Stdout, out)

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

		fmt.Printf("\n\nProcessing, this may take a while...\n\n")
	}()
	return in
}

func query(in <-chan string) (<-chan *RepoStats, <-chan error) {
	out := make(chan *RepoStats)
	errc := make(chan error)

	go func() {
		defer close(out)
		var wg sync.WaitGroup

		client := NewClient(context.Background(), accessToken)

		for s := range in {
			wg.Add(1)
			go func(s string) {
				defer wg.Done()
				fields := strings.Split(s, "/")
				if len(fields) != 2 {
					// TODO(yuankun): handle this
					fmt.Printf("invalid input")
					return
				}
				stats, err := client.Query(fields[0], fields[1])
				if err != nil {
					// TODO(yuankun): handle this
					fmt.Printf("query failed: %s", err)
					return
				}
				out <- stats
			}(s)
		}

		wg.Wait()
	}()
	return out, errc
}

func output(w io.Writer, out <-chan *RepoStats) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)

		writer := csv.NewWriter(w)
		var records [][]string

		records = append(records, CsvHeader())

		for o := range out {
			records = append(records, o.CsvRecord())
		}

		writer.WriteAll(records)
		writer.Flush()
	}()
	return done
}
