/*
 * || Setup Instructions ||
 * 1. F1 -> Click on "Terminal: Create New Terminal" -> Click on "Command Prompt"
 * 2. cd FOLDERNAME
 *
 * || Relevant Terminal Commands ||
 * go run FILENAME.go (Runs file); go build (Build and format files in current directory);
 * gofmt (Format files into standard format);
 *
 * WARNING: Every folder needs a "go.mod" if there isn't one, type "go mod init NAME" in terminal
 *
 * || Instructions ||
 * Write Matrix Multiplication using go routines to accelerate computation and compare the
 * sequential vs. accelerated execution time for a matrix of at least 10000x10000 float random
 * numbers.
 */

package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

func stringCounterSequence(pathname, userString string) (int, error) {
	file, err := os.Open(pathname) // open file for segmented reading
	if err != nil {                // if an error occurs while opening file
		return 0, err
	}
	defer file.Close() // closes file at the end of the function

	count := 0

	scanner := bufio.NewScanner(file) // entity that checks for userString in buffer
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()                          // get the current word
		if strings.Contains(word, userString) == true { // check if word matches the string to be found
			count++
		}
	}

	err = scanner.Err() // retrieves any error that may have occurred in scanner
	if err != nil {     // if an error occurs during scanning
		return 0, err
	}

	return count, nil
}

func stringCounterParallel(pathname, userString string) (int, error) {
	file, err := os.Open(pathname)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var wg sync.WaitGroup
	totalCount := 0                        // tracks total count among all the goroutines
	numWorkers := runtime.NumCPU()         // limit child process to # of cores
	chunk := make(chan []byte, numWorkers) // create a channel to pass file chunks to workers
	bufferSize := 1024 * 64                // size of each chunk in bytes

	// Start worker goroutines //
	for i := 0; i < numWorkers; i++ {
		wg.Add(1) // increment counter
		go func() {
			defer wg.Done() // decrement counter at the end of the go function
			for chunk := range chunk {
				chunkCount := 0 // word counter for chunk

				scanner := bufio.NewScanner(strings.NewReader(string(chunk)))
				scanner.Split(bufio.ScanWords) // reader the buffer word by word

				for scanner.Scan() {
					word := scanner.Text() // get the current word
					if strings.Contains(word, userString) {
						chunkCount++
					}
				}
				totalCount += chunkCount
			}
		}()
	}

	// Read file in chunks and send to workers //
	buffer := make([]byte, bufferSize)
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil && err.Error() != "EOF" { // if an error occurs and it's not an EOF
			return 0, err
		}

		if bytesRead == 0 { // end of file
			break
		}
		chunk <- append([]byte{}, buffer[:bytesRead]...) // send a copy of the buffer to the channel (to avoid overwriting)
	}

	close(chunk) // close the channel to signal workers to stop
	wg.Wait()    // wait for all workers to finish
	return totalCount, nil
}

func main() {
	// Prints command format if the # of arguements != 3 //
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run WordCount.go <Search Term> <Filename>")
		return
	}

	userString := os.Args[1] // get search term from cmd line
	filePath := os.Args[2]   // get filename from cmd line

	/// Sequential Search ///
	fmt.Println("Running sequential search")
	start := time.Now() // start timer for a sequential search
	sequentialCount, err := stringCounterSequence(filePath, userString)

	if err != nil { // if stringCounterSeq() returns an error, print message
		fmt.Println("Error:", err)
		return
	}
	sequentialTime := time.Since(start) // end timer for a sequential search
	fmt.Printf("Sequential Search: Found %d occurrences of %s in %v\n", sequentialCount, userString, sequentialTime)

	/// Parallel Search ///
	fmt.Println("Running parallel search")
	start = time.Now() // start timer for parallel search

	parallelCount, err := stringCounterParallel(filePath, userString)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	parallelTime := time.Since(start) // end timer for a parallel search
	fmt.Printf("Parallel Search: Found %d occurrences of %s in %v\n", parallelCount, userString, parallelTime)
}
