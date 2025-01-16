/*
 * || Setup Instructions ||
 * 1. F1 -> Click on "Terminal: Create New Terminal" -> Click on "Command Prompt"
 * 2. cd FOLDERNAME
 *
 * || Relevant Terminal Commands ||
 * go run FILENAME.go (Runs file); go build (Build and format files in current directory);
 * gofmt FILENAME.go (Format files into standard format)
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
	"time"
)

func main() {
	start := time.Now() // Start time

	// Define two 3x3 matrices //
	matrixOne := [3][3]int16{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	matrixTwo := [3][3]int16{
		{9, 8, 7},
		{6, 5, 4},
		{3, 2, 1},
	}

	// Channel to receive the result //
	resultChannel := make(chan [3][3]int16)
	sizeChannel := make(chan int16)

	// Goroutine for matrix multiplication //
	go func( /*PARAMETERS HERE*/ ) {
		resultMatrix, sizeOfMatrix := multiplyMatrix(matrixOne, matrixTwo)
		resultChannel <- resultMatrix
		sizeChannel <- sizeOfMatrix
	}( /*ARGUEMENTS HERE*/ )

	// Retrieve result from channel //
	resultMatrix := <-resultChannel
	sizeOfMatrix := <-sizeChannel

	// Print the result matrix //
	printMatrix(resultMatrix, sizeOfMatrix)

	end := time.Now()                                  //End time
	fmt.Printf("Execution time: %v\n", end.Sub(start)) //Calculate and print execution time
}

func multiplyMatrix(matrixOne, matrixTwo [3][3]int16) ([3][3]int16, int16) {
	var resultMatrix [3][3]int16

	rows := int16(3)
	cols := int16(3)

	// Matrix multiplication logic
	for i := int16(0); i < rows; i++ { //Iterate over rows of the first matrix
		go func(row int16) { //Goroutine for every row
			for j := int16(0); j < cols; j++ { //Iterate over columns of the second matrix
				sum := int16(0)
				for k := int16(0); k < cols; k++ { //Dot product calculation
					sum += matrixOne[i][k] * matrixTwo[k][j]
				}
				resultMatrix[i][j] = sum
			}
		}(i)
	}
	return resultMatrix, rows
}

func printMatrix(resultMatrix [3][3]int16, sizeOfMatrix int16) {
	for i := int16(0); i < sizeOfMatrix; i++ {
		fmt.Println(resultMatrix[i])
	}
}
