package main

import "fmt"

func main() {

	var matA = [][]int {
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
	}

	var matB = [][]int {
		{9, 10, 11},
		{12, 13, 14},
		{15, 16, 17},
	}

	var result = matmul(matA, matB)

	fmt.Println(result)
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
