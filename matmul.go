package main

import (
	"fmt"
	"math/rand"
	"time"
	"sync"
)

func main() {

	var n = 1000

	var matA = genmats(n, n)

	var matB = genmats(n, n)

	// fmt.Println("Matrix A:")
	// for _, row := range matA {
	// 	fmt.Println(row)
	// }

	// fmt.Println("Matrix B:")
	// for _, row := range matB {
	// 	fmt.Println(row)
	// }

	start := time.Now()
	//var resultsequent = 
	matmul(matA, matB)
	elapsed := time.Since(start)
	
	// fmt.Println("Result Matrix Sequential:")
	// for _, row := range resultsequent {
	// 	fmt.Println(row)
	// }

	fmt.Println("Time:", elapsed)

	start = time.Now()
	//var resultparallel = 
	matmulparallel(matA, matB)
	elapsed = time.Since(start)

	// fmt.Println("Result Matrix Parallel:")
	// for _, row := range resultparallel {
	// 	fmt.Println(row)
	// }

	fmt.Println("Time:", elapsed)
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
	} // repeate for next row of matA
	return result // return new matrix to main
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


// parallel operations
/*package main

import (
    "fmt"
    "sync"
)

func process(id int, wg *sync.WaitGroup) {
    defer wg.Done()
    // Perform some time-consuming operation
    fmt.Println("Processing:", id)
}

func main() {
    var wg sync.WaitGroup
    numTasks := 5

    for i := 0; i < numTasks; i++ {
        wg.Add(1)
        go process(i, &wg)
    }

    wg.Wait()
    fmt.Println("All tasks complete")
}*/

func process(matA, matB [][]int, result [][]int, i, j int, wg *sync.WaitGroup) { // break off matrix multiplication loop into a separate function
	defer wg.Done()
	sum := 0
	for k := 0; k < len(matA[0]); k++ {
		sum += matA[i][k] * matB[k][j] // calculate result of each matrix entry
	}
	result[i][j] = sum // store result into matrix entry
}

func matmulparallel(matA, matB [][]int) ([][]int) {
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

	for i := 0; i < rowA; i++ {
		for j := 0; j < colB; j++ {
			wg.Add(1) // for each matrix entry create a new process
			go process(matA, matB, result, i, j, &wg) // call process funtion to calc result of each matrix entry
		}
	}
	wg.Wait() // wait for all child processes to terminate
	return result // return completed matrix
}