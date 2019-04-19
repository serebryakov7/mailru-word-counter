package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	SearchPattern    = "Go"
	HttpSchemePrefix = "http"
	WorkersCount     = 5
)

func getHttp(url string) ([]byte, error) {
	res, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return ioutil.ReadAll(res.Body)
}

func count(source string) (int, error) {
	var (
		data []byte
		err  error
	)

	if strings.HasPrefix(source, HttpSchemePrefix) {
		data, err = getHttp(source)
	} else {
		data, err = ioutil.ReadFile(source)
	}

	if err != nil {
		return 0, err
	}

	return strings.Count(string(data), SearchPattern), nil
}

func main() {
	var (
		sources   = make(chan string)
		semaphore = make(chan struct{}, WorkersCount)
		total     uint64
		wg        sync.WaitGroup
	)

	go func() {
		for source := range sources {
			wg.Add(1)

			go func(source string) {
				semaphore <- struct{}{}

				defer func() {
					<-semaphore
					wg.Done()
				}()

				count, err := count(source)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error for %s: %s\n", source, err)
					return
				}

				fmt.Fprintf(os.Stdout, "Count for %s: %v\n", source, count)
				atomic.AddUint64(&total, uint64(count))

			}(source)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		for _, src := range strings.Split(scanner.Text(), "\n") {
			sources <- src
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln("Reading error: ", err)
	}

	close(sources)
	wg.Wait()

	fmt.Printf("Total: %v\n", total)
}
