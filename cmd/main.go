package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"net/http"
	"os"
)

type Piece int

const (
	Available Piece = 0
	Black           = 1
	White           = 2
)

type GameData struct {
	data [][]int `json:"data"`
}

// nextstep :下一步走哪里？
func nextStep(gameData GameData, side Piece) (int, int) {
	x, y := 0, 0
	width := len(gameData.data)
	fmt.Println("game board width:", width)
	fmt.Println(fmt.Sprintf("x= %d,y= %d", x, y))
	maxWeight := 0
	for i := 0; i < len(gameData.data); i++ {
		for j := 0; i < len(gameData.data); i++ {
			c := gameData.data[i][j]
			if c > 0 {
				continue
			}
			gameData.data[i][j] = int(side)
			t := computeWeight(gameData, side, i, j)
			if maxWeight < t {
				maxWeight = t
				x, y = i, j
			}
		}
	}
	fmt.Println("weight:", maxWeight)
	return x, y
}

type lineWeight map[int]int
type DirectionLineWeight []lineWeight

func Print(data DirectionLineWeight) {
	for i := 0; i < len(data); i++ {
		fmt.Printf("\n\r%d :\n\r", i)
		for k1, v1 := range data[i] {
			fmt.Printf(" %d %d\n\r", k1, v1)
		}
	}
	fmt.Println()
}

// computeWeight :计算所有方向的权重
func computeWeight(data GameData, side Piece, x int, y int) int {
	otherWeights := make([]lineWeight, 4)
	otherWeights[0] = make(map[int]int)
	otherWeights[1] = make(map[int]int)
	otherWeights[2] = make(map[int]int)
	otherWeights[3] = make(map[int]int)

	weights := make([]lineWeight, 4)
	weights[0] = make(map[int]int)
	weights[1] = make(map[int]int)
	weights[2] = make(map[int]int)
	weights[3] = make(map[int]int)

	weight, otherWeight := 0, 0
	data0 := data.data

	printMatrix(data0)

	// 计算横向
	for i := 0; i < len(data0); i++ {
		row := data0[i]
		b, w := computeRowWeight(&row)
		if side == Black {
			if otherWeight < w {
				otherWeight = w
			}
			if weight < b {
				weight = b
			}
			otherWeights[0][i] = w
			weights[0][i] = b
		} else {
			if otherWeight < b {
				otherWeight = b
			}
			if weight < w {
				weight = w
			}

			otherWeights[0][i] = b
			weights[0][i] = w
		}
	}

	fmt.Println(data0)

	data2 := rotate45Matrix(data0)
	for i := 0; i < len(data2); i++ {
		row := data2[i]
		b, w := computeRowWeight(&row)
		if side == Black {
			if otherWeight < w {
				otherWeight = w
			}
			if weight < b {
				weight = b
			}
			otherWeights[2][i] = w
			weights[2][i] = b
		} else {
			if otherWeight < b {
				otherWeight = b
			}
			if weight < w {
				weight = w
			}
			otherWeights[2][i] = b
			weights[2][i] = w
		}
	}

	// 旋转90度计算
	data1 := rorateMatrix(data.data)
	printMatrix(data1)

	for i := 0; i < len(data1); i++ {
		row := data1[i]
		b, w := computeRowWeight(&row)
		if side == Black {
			if otherWeight < w {
				otherWeight = w
			}
			if weight < b {
				weight = b
			}
			otherWeights[1][i] = w
			weights[1][i] = b
		} else {
			if otherWeight < b {
				otherWeight = b
			}
			if weight < w {
				weight = w
			}
			otherWeights[1][i] = b
			weights[1][i] = w
		}
	}

	// 斜向计算
	// 顺时45度
	fmt.Println("data2")

	fmt.Println("data3")
	data3 := rotate45Matrix(data1)
	for i := 0; i < len(data3); i++ {
		row := data3[i]
		b, w := computeRowWeight(&row)
		if side == Black {
			if otherWeight < w {
				otherWeight = w
			}
			if weight < b {
				weight = b
			}
			otherWeights[3][i] = w
			weights[3][i] = b
		} else {
			if otherWeight < b {
				otherWeight = b
			}
			if weight < w {
				weight = w
			}
			otherWeights[3][i] = b
			weights[3][i] = w
		}
	}

	Print(weights)
	Print(otherWeights)

	// 有三点已经一线了，需要堵上
	if otherWeight >= 300 {

	}

	// 计算纵向
	// 计算斜向 45度
	// 计算逆斜向 -45度

	fmt.Println(fmt.Sprintf("weight:%d, other weight:%d", weight, otherWeight))
	return weight
}

// 旋转矩阵
func rorateMatrix(matrix [][]int) [][]int {
	for i, j := 0, len(matrix)-1; i < j; i, j = i+1, j-1 {
		matrix[i], matrix[j] = matrix[j], matrix[i]
	}

	// transpose it
	for i := 0; i < len(matrix); i++ {
		for j := 0; j < i; j++ {
			matrix[i][j], matrix[j][i] = matrix[j][i], matrix[i][j]
		}
	}
	return matrix
}

func reverse(x []int) []int {
	var r []int
	for i := len(x) - 1; i >= 0; i-- {
		r = append(r, x[i])
	}

	return r
}

func rotate45Matrix(matrix [][]int) [][]int {
	var m [][]int

	for line := 0; line < 2*len(matrix)-1; line++ {
		var r []int
		for i := 0; i < len(matrix); i++ {
			for j := 0; j < len(matrix[i]); j++ {
				if i+j == line {
					r = append(r, matrix[i][j])
					break
				}
			}
		}

		r = reverse(r)
		fmt.Printf("%#v\n\r", r)
		m = append(m, r)
	}

	return m
}

type pieceTimeWeight map[int]int

type pieceWeight map[int]pieceTimeWeight

// computeRowWeight :计算单行权重
func computeRowWeight(row *[]int) (int, int) {
	p := make(map[int]pieceTimeWeight)
	p[0] = make(map[int]int)
	p[1] = make(map[int]int)
	p[2] = make(map[int]int)

	times := 0
	c := -1
	lastC := -1
	whiteWeight := 0
	blackWeight := 0
	data := *row

	for i := 0; i < len(data); i++ {
		c = data[i]
		if c != lastC {
			if lastC > 0 {
				_, ok := p[lastC][times]
				if !ok {
					p[lastC][times] = 0
				}
				p[lastC][times]++
			}
			times = 1
		} else {
			times++
		}
		lastC = c
	}

	if lastC > 0 {
		_, ok := p[lastC][times]
		if !ok {
			p[lastC][times] = 0
		}
		p[lastC][times]++
	}

	for k, v := range p {
		if k <= 0 {
			continue
		}

		weight := 0

		for k1, v1 := range v {
			weight = weight + int(math.Pow(10, float64(k1)-1))*v1
		}

		if k == 1 {
			blackWeight = weight
		} else {
			whiteWeight = weight
		}
	}

	return blackWeight, whiteWeight
}

var _gameData = GameData{
	data: [][]int{
		{0, 0, 0, 0, 1, 1, 2, 0, 1, 2, 0},
		{1, 2, 1, 2, 0, 1, 1, 2, 2, 2, 1},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}}

func printMatrix(data [][]int) {
	fmt.Println("matrix begin")
	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data[i]); j++ {
			fmt.Printf("%d ", data[i][j])
		}
		fmt.Println()
	}
	fmt.Println("matrix end")
	fmt.Println()
}

func main() {
	fmt.Println("Gomoku Game Simple algorithm, code by shrek 2022")

	//row := _gameData.data[0]
	//
	//fmt.Println(row)
	//
	//b, w := computeRowWeight(&row)
	//fmt.Println(b, w)
	//
	//row = _gameData.data[1]
	//
	//fmt.Println(row)
	//
	//b, w = computeRowWeight(&row)
	//fmt.Println(b, w)
	data := _gameData
	//printMatrix(data.data)
	//
	//var _ = rotate45Matrix(data.data)
	//os.Exit(1)

	v := computeWeight(data, Black, 0, 0)
	fmt.Println(v)

	os.Exit(0)

	r := gin.Default()
	r.GET("/api/v1/hello", func(c *gin.Context) {
		c.String(200, "Hello, Gomoku")
	})

	// POST
	r.POST("/api/v1/nextstep", func(c *gin.Context) {
		json := GameData{}
		c.BindJSON(&json)
		log.Printf("%v", &json)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"x":       0,
			"y":       0,
		})
	})
	r.Run()
}
