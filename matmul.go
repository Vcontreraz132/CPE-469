package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {

	var matA = genmats(3, 3)

	var matB = genmats(3, 3)

	fmt.Println("Matrix A:")
	for _, row := range matA {
		fmt.Println(row)
	}

	fmt.Println("Matrix B:")
	for _, row := range matB {
		fmt.Println(row)
	}

	var result = matmul(matA, matB)

	fmt.Println("Result Matrix:")
	for _, row := range result {
		fmt.Println(row)
	}
}

func matmul (matA [][]int, matB [][]int) ([][]int) { // (input parameter list) (output parameter list)
	rowA := len(matA) // gets number of rows in A
	rowB := len(matB) // number of rows in B
	colsA := len(matA[0])
	colB := len(matB[0]) // number of columns in B = number of entries in first row of B

	if colsA != rowB {
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