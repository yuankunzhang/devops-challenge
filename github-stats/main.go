package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

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

func query(in <-chan string) (<-chan string, <-chan error) {
	out := make(chan string)
	errc := make(chan error)
	go func() {
		defer close(out)
		var wg sync.WaitGroup

		for i := range in {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(5 * time.Second)
				out <- i
			}()
		}

		wg.Wait()
	}()
	return out, errc
}

func output(w io.Writer, out <-chan string) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)

		var s []string
		for o := range out {
			s = append(s, o)
		}

		fmt.Printf("%v\n", s)
	}()
	return done
}
