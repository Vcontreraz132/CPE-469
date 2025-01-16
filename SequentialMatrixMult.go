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
	"fmt"
	"math/rand"
	"time"
)

func main() {
	start := time.Now() // Start time

	// Declare matrices //
	matrixOne := [8][8]int16{}
	matrixTwo := [8][8]int16{}
	resultMatrix := [8][8]int16{}

	// Dimensions of the matrices //
	rows := 8
	cols := 8

	// Initialize matrices with random values from 0 - 4 (inclusive) //
	for i := 0; i < rows; i++ {
		for j := 0; j < rows; j++ {
			matrixOne[i][j] = int16(rand.Intn(5))
			matrixTwo[i][j] = int16(rand.Intn(5))
		}
	}

	// Matrix multiplication //
	for i := 0; i < rows; i++ { //Iterate over rows of the first matrix
		for j := 0; j < cols; j++ { //Iterate over columns of the second matrix
			sum := int16(0)
			for k := 0; k < cols; k++ { //Dot product calculation
				sum += matrixOne[i][k] * matrixTwo[k][j]
			}
			resultMatrix[i][j] = sum
		}
	}

	fmt.Println("|| Matrix One ||")
	for i := 0; i < rows; i++ {
		fmt.Println(matrixOne[i])
	}

	fmt.Println("\n|| Matrix Two ||")
	for i := 0; i < rows; i++ {
		fmt.Println(matrixTwo[i])
	}

	fmt.Println("\n|| Result Matrix ||")
	for i := 0; i < rows; i++ {
		fmt.Println(resultMatrix[i])
	}

	end := time.Now()                                  // End time
	fmt.Printf("Execution time: %v\n", end.Sub(start)) // Calculate and print execution time
}
