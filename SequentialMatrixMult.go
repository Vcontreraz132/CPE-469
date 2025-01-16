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

// import (
// 	"fmt"
// 	"time"
// )

// func main() {
// 	start := time.Now() // Start time

// 	// Define two 3x3 matrices
// 	matrixOne := [3][3]int16{
// 		{1, 2, 3},
// 		{4, 5, 6},
// 		{7, 8, 9},
// 	}
// 	matrixTwo := [3][3]int16{
// 		{9, 8, 7},
// 		{6, 5, 4},
// 		{3, 2, 1},
// 	}

// 	// Resultant matrix to store the result of multiplication
// 	var resultMatrix [3][3]int16

// 	// Dimensions of the matrices
// 	rows := 3
// 	cols := 3

// 	// Matrix multiplication logic
// 	for i := 0; i < rows; i++ { // Iterate over rows of the first matrix
// 		for j := 0; j < cols; j++ { // Iterate over columns of the second matrix
// 			sum := int16(0)
// 			for k := 0; k < cols; k++ { // Dot product calculation
// 				sum += matrixOne[i][k] * matrixTwo[k][j]
// 			}
// 			resultMatrix[i][j] = sum
// 		}
// 	}

// 	for i := 0; i < rows; i++ {
// 		fmt.Println(resultMatrix[i])
// 	}

// 	end := time.Now()                                  // End time
// 	fmt.Printf("Execution time: %v\n", end.Sub(start)) // Calculate and print execution time
// }
