package main

import (
	"fmt"
	"math/rand"
	"time"
	"sync"
	"runtime"
)

func main() {

	var n = 10000
	var block = 200 // partition matrices into blocks

	fmt.Println("Generating matrices")
	var matA = genmats(n, n)

	var matB = genmats(n, n)
	fmt.Println("Generation Complete")


	fmt.Println("Sequential Start")
	start := time.Now()
	//var resultsequent = 
	matmulSequentialBlock(matA, matB, block)
	elapsed := time.Since(start)
	fmt.Println("Sequential Time:", elapsed)

	fmt.Println("Parallel Start")
	start = time.Now()
	//var resultparallel = 
	matmulparallel(matA, matB, block)
	elapsed = time.Since(start)
	fmt.Println("Parallel Time:", elapsed)
}

func matmul (matA [][]int, matB [][]int) ([][]int) { // (input parameter list) (output parameter list)
	rowA := len(matA) // gets number of rows in A
	rowB := len(matB) // number of rows in B
	colA := len(matA[0])
	colB := len(matB[0]) // number of columns in B = number of entries in first row of B

	if colA != rowB {
		panic("not correct dimensions")
	}

	var i, j, k int // row, column counters for matA and matB
	temp := 0
	result := make([][]int, rowA)
	for i := range result {
		result[i] = make([]int, colB)
	}

	for i = 0; i < rowA; i++ { // for each entry in current row of matA
		for j = 0; j < colB; j++ { // for each entry in current column of matB
			for k = 0; k < rowB; k++ {
				temp = temp + matA[i][k]*matB[k][j] // multiply entries and add result of multi to previous value, repeate for each entry in current row of matB
			}
			result[i][j] = temp // store result of calc into resulting matrix entry
			temp = 0 // clear calc
		} // repeate for next column of matB
		//fmt.Println("Entry:", i, "Done")
	} // repeate for next row of matA
	return result // return new matrix to main
}

func matmulSequentialBlock(matA, matB [][]int, block int) [][]int {
	rowA := len(matA)
	colB := len(matB[0])
	colA := len(matA[0])

	if colA != len(matB) {
		panic("Matrix dimensions do not match for multiplication")
	}

	// Initialize the result matrix
	result := make([][]int, rowA)
	for i := range result {
		result[i] = make([]int, colB)
	}

	// Perform block-based multiplication
	for i := 0; i < rowA; i += block { // Iterate over blocks horizontally in A
		for j := 0; j < colB; j += block { // Iterate over blocks vertically in B
			for k := 0; k < colA; k += block { // Iterate over blocks vertically in A
				// Process the current block sequentially
				for ii := i; ii < i+block && ii < rowA; ii++ { // Move horizontally within the block in A
					for jj := j; jj < j+block && jj < colB; jj++ { // Move vertically within the block in B
						sum := 0
						for kk := k; kk < k+block && kk < colA; kk++ { // Move within the block
							sum += matA[ii][kk] * matB[kk][jj]
						}
						result[ii][jj] += sum
					}
				}
			}
		}
		fmt.Println("Block", i, "Done")
	}
	return result
}


func genmats(rows, cols int) ([][]int) {
	rand.Seed(time.Now().UnixNano()) // seed rand generator with current time
	matrix := make([][]int, rows) // resulting matrix, 2d array
	for i := 0; i < rows; i++ {
		matrix[i] = make([]int, cols)
		for j := 0; j < cols; j++ {
			matrix[i][j] = rand.Intn(10) // populate each entry with number between 0 and 10
		}
	}
	return matrix
}

func matmulparallel(matA, matB [][]int, block int) ([][]int) {
	rowA := len(matA) // gets number of rows in A
	rowB := len(matB) // number of rows in B
	colA := len(matA[0])
	colB := len(matB[0]) // number of columns in B = number of entries in first row of B

	if colA != rowB {
		panic("not correct dimensions")
	}

	result := make([][]int, rowA)
	for i := range result {
		result[i] = make([]int, colB)
	}

	var wg sync.WaitGroup

	numCPU := runtime.NumCPU()

	channel := make(chan struct{}, numCPU) 

	for i := 0; i < rowA; i += block { // i moves to next block horizontally through A
		for j := 0; j < colB; j += block { // j moves to next block vertically through B
			for k := 0; k < colA; k += block { // k moves to next block vertically through A

				channel <- struct{}{} // Acquire a worker slot

				wg.Add(1) // for each matrix block create a new process

				go func (i, j, k int) { // create goroutine funtion to calc result of each matrix block
					defer wg.Done()

					defer func() {
					<-channel // receive and close channel
					}()

					for ii := i; ii < i + block && ii < rowA; ii++ { // move horizontally through the A block
						for jj := j; jj < j + block && jj < colB; jj++ { // move vertically through B block
							sum := 0
							for kk := k; kk < k + block && kk < colA; kk++ { // move vertically thorugh A block
								sum += matA[ii][kk] * matB[kk][jj] // calculate result of each matrix entry
							}
							result[ii][jj] += sum
						}
					}
				}(i, j, k)
			}
		}
		fmt.Println("Block,", i, "Done")
	}
	wg.Wait() // wait for all child processes to terminate
	return result // return completed matrix
}
