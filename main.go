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
		semaphore = make(chan struct{}, WorkersCount)
		total     	uint64
		wg        	sync.WaitGroup
	)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		for _, source := range strings.Split(scanner.Text(), "\n") {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(source string) {
				defer func() {
					wg.Done()
					<-semaphore
				}()

				count, err := count(source)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error for %s: %s\n", source, err)
					return
				}

				fmt.Fprintf(os.Stdout, "Count for %s: %d\n", source, count)
				atomic.AddUint64(&total, uint64(count))

			}(source)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln("Reading error: ", err)
	}

	wg.Wait()
	fmt.Printf("Total: %d\n", total)
}
