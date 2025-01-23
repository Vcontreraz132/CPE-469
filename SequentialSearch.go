/*
 * || Setup Instructions ||
 * 1. F1 -> Click on "Terminal: Create New Terminal" -> Click on "Command Prompt"
 * 2. cd FOLDERNAME
 *
 * || Relevant Terminal Commands ||
 * go run FILENAME.go (Runs file); go build (Build and format files in current directory);
 * gofmt (Format files into standard format)
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
	"strings"
	"time"
	//"runtime"
)

func stringCounterSeq(pathname, userString string) (int, error) {
	file, err := os.Open(pathname)						// open file for segmented reading
	if err != nil {										// if an error occurs while opening file
		return 0, err
	}
	defer file.Close()									// closes file at the end of the function

	count := 0

	scanner := bufio.NewScanner(file)					// entity that checks for userString in buffer
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text() 							// get the current word
		if strings.Contains(word, userString) == true {	// check if word matches the string to be found
			count++
		} 
	}

	err = scanner.Err()									// retrieves any error that may have occurred in scanner
	if err != nil {										// if an error occurs during scanning
		return 0, err
	}

	return count, nil
}

func stringCounterParallel(pathname, userString string, numWorkers int) (int, error) {
	file, err := os.Open(pathname)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}

	chunkSize := len(lines) / numWorkers
	results := make(chan int, numWorkers)
	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numWorkers - 1 {
			end = len(lines)
		}

		go func(linesChunk []string) {
			count := 0
			for _, line := range linesChunk {
				count += strings.Count(line, userString)
			}
			results <- count
		}(lines[start:end])
	}

	totalCount := 0
	for i := 0; i < numWorkers; i++ {
		totalCount += <-results
	}
	return totalCount, nil
}

func main() {
	
	// Prints command format if the # of arguements != 3 //
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run sequentialSearch.go <Search Term> <Filename>")
		return
	}

	userString := os.Args[1] 										// get search term from cmd line
	filePath := os.Args[2] 											// get filename from cmd line
	//numWorkers := runtime.NumCPU() 									// limit child process to # of cores

	fmt.Println("Running sequential search")
	start := time.Now()												// start timer for a sequential search
	sequentialCount, err := stringCounterSeq(filePath, userString)

	if err != nil {													// if stringCounterSeq() returns an error, print message
		fmt.Println("Error:", err)
		return
	}
	sequentialTime := time.Since(start)								// end timer for a sequential search
	fmt.Printf("Sequential Search: Found %d occurrences of %s in %v\n", sequentialCount, userString, sequentialTime)

	// //time.Sleep(2 * time.Second)

	// fmt.Println("Running parallel search")
	// start = time.Now()
	// parallelCount, err := stringCounterParallel(filePath, userString, numWorkers)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }
	// parallelTime := time.Since(start)
	// fmt.Printf("Parallel Search: Found %d occurrences of %s in %v\n", parallelCount, userString, parallelTime)
}