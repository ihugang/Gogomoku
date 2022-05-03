package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
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
func nextStep(data []int, side Piece) int {
	width := int(math.Sqrt(float64(len(data))))
	data2 := make([][]Point, width)
	for i := 0; i < width; i++ {
		data2[i] = make([]Point, width)
		for j := 0; j < width; j++ {
			p := Point{x: i, y: j, value: data[i*width+j]}
			data2[i][j] = p
		}
	}

	x, y := compute(data2, side)
	return x*width + y
}

// JudgeWin : 判断是否赢局
func JudgeWin(data []int) (Piece, []Point) {
	width := int(math.Sqrt(float64(len(data))))
	fmt.Println("matrix width:", width)
	data0 := make([][]Point, width)
	for i := 0; i < width; i++ {
		data0[i] = make([]Point, width)
		for j := 0; j < width; j++ {
			p := Point{x: i, y: j, value: data[i*width+j]}
			data0[i][j] = p
		}
	}

	matrixLength := len(data0)
	rowsLength := matrixLength*6 - 2
	fmt.Println("need compute rows:", rowsLength)

	var rows [][]Point

	for i := 0; i < matrixLength; i++ {
		row := data0[i]
		rows = append(rows, row)
	}

	data2 := rotate45Matrix(data0)
	for i := 0; i < len(data2); i++ {
		row := data2[i]
		rows = append(rows, row)
	}

	data1 := rotateMatrix(data0)
	for i := 0; i < matrixLength; i++ {
		row := data1[i]
		rows = append(rows, row)
	}

	data3 := rotateReverse45Matrix(data0)
	for i := 0; i < len(data3); i++ {
		row := data3[i]
		rows = append(rows, row)
	}

	for i := 0; i < rowsLength; i++ {
		fmt.Println("line ", i)
		row := rows[i]
		printRow(row)
		win, side, p := judgeRowWin(row)
		if win {
			return side, p
		}
	}

	return Available, nil
}

type lineWeight map[int]int
type DirectionLineWeight []lineWeight

type RowWeight struct {
	RowNo  int
	Weight int
}

func Print(data DirectionLineWeight) {
	for i := 0; i < len(data); i++ {
		fmt.Printf("\n\r%d :\n\r", i)
		for k1, v1 := range data[i] {
			fmt.Printf(" %d %d\n\r", k1, v1)
		}
	}
	fmt.Println()
}

func printRow(data []Point) {
	fmt.Println("row begin")
	for i := 0; i < len(data); i++ {
		fmt.Printf("%2d ", data[i].value)
	}
	fmt.Printf("\n\r")
	fmt.Println("row end")
}

// compute : 计算
// 1、将所有的行可能性集中在一起
// 2、然后逐行计算当前权值
// 3、判断对方有没有四连的，如有防堵
func compute(data [][]Point, side Piece) (int, int) {
	matrixLength := len(data)
	rowsLength := matrixLength*6 - 2
	fmt.Println("need compute rows:", rowsLength)

	fmt.Println("data:")
	var rows [][]Point
	otherWeights := make([]RowWeight, rowsLength)
	weights := make([]RowWeight, rowsLength)

	for i := 0; i < matrixLength; i++ {
		row := data[i]
		rows = append(rows, row)
	}
	printMatrix(data)

	fmt.Println("data2 45度")
	data2 := rotate45Matrix(data)
	for i := 0; i < len(data2); i++ {
		row := data2[i]
		rows = append(rows, row)
	}

	fmt.Println("data1 90度")
	data1 := rotateMatrix(data)
	for i := 0; i < matrixLength; i++ {
		row := data1[i]
		rows = append(rows, row)
	}
	printMatrix(data1)

	fmt.Println("data3 -45度")
	data3 := rotateReverse45Matrix(data)
	for i := 0; i < len(data3); i++ {
		row := data3[i]
		rows = append(rows, row)
	}

	otherWeight := 0
	weight := 0

	printRows(rows)

	for i := 0; i < rowsLength; i++ {
		fmt.Println("Compute line ", i)
		row := rows[i]
		printRow(row)
		log.Printf("line %d : %#v ", i, row)
		b, w := computeRowWeight(row)
		if side == Black {
			if otherWeight < w {
				otherWeight = w
			}
			if weight < b {
				weight = b
			}
			otherWeights[i] = RowWeight{RowNo: i, Weight: w}
			weights[i] = RowWeight{RowNo: i, Weight: b}
		} else {
			if otherWeight < b {
				otherWeight = b
			}
			if weight < w {
				weight = w
			}

			otherWeights[i] = RowWeight{RowNo: i, Weight: b}
			weights[i] = RowWeight{RowNo: i, Weight: w}

		}
		fmt.Printf("%4d %6d %6d\n\r", i, b, w)
	}

	fmt.Println("weight:", weight, "otherWeight:", otherWeight)

	sort.SliceStable(otherWeights, func(i, j int) bool {
		return otherWeights[i].Weight > otherWeights[j].Weight
	})

	fmt.Println("开始计算...")
	// 开始计算
	if otherWeight >= 1000 {
		for k := 0; k < len(otherWeights); k++ {
			firstRow := otherWeights[0]
			fmt.Println(firstRow)
			printRow(rows[firstRow.RowNo])

			newWeight := otherWeight
			newSelfWeight := weight
			lastMinIndex := -1
			lastMaxIndex := -1

			for i := 0; i < len(rows[firstRow.RowNo]); i++ {
				crow := make([]Point, len(rows[firstRow.RowNo]))
				copy(crow, rows[firstRow.RowNo])
				if crow[i].value == 0 {
					crow[i].value = int(side)
					printRow(crow)

					b, w := computeRowWeight(crow)
					if side == Black {
						if newWeight > w {
							newWeight = w
							lastMinIndex = i
						}
						if newSelfWeight < b {
							newSelfWeight = b
							lastMaxIndex = i
						}
					} else {
						if newWeight > b {
							newWeight = b
							lastMinIndex = i
						}
						if newSelfWeight < w {
							newSelfWeight = w
							lastMaxIndex = i
						}
					}
				}
			}

			if lastMinIndex > -1 {
				point := rows[firstRow.RowNo][lastMinIndex]
				fmt.Println("影响对家 New Step:", point.x, point.y)
				return point.x, point.y
			} else {
				// 对对方没有影响
				if lastMaxIndex > -1 {
					point := rows[firstRow.RowNo][lastMinIndex]
					fmt.Println("利于自己 New Step:", point.x, point.y)
					return point.x, point.y
				}
			}
		}

	} else {
		// 对方暂时没有威胁
		sort.SliceStable(weights, func(i, j int) bool {
			return weights[i].Weight > weights[j].Weight
		})

		for k := 0; k < len(weights); k++ {
			firstRow := weights[k]
			lastMaxIndex := -1
			newSelfWeight := 0
			for i := 0; i < len(rows[firstRow.RowNo]); i++ {
				crow := make([]Point, len(rows[firstRow.RowNo]))
				copy(crow, rows[firstRow.RowNo])
				if crow[i].value == 0 {
					crow[i].value = int(side)
					printRow(crow)

					b, w := computeRowWeight(crow)
					if side == Black {

						if newSelfWeight < b {
							newSelfWeight = b
							lastMaxIndex = i
						}
					} else {
						if newSelfWeight < w {
							newSelfWeight = w
							lastMaxIndex = i
						}
					}
				}
			}

			if lastMaxIndex > -1 {
				point := rows[firstRow.RowNo][lastMaxIndex]
				fmt.Println("没有威胁时 利于自己 New Step:", point.x, point.y)
				return point.x, point.y
			}
		}

	}

	// 最优选择如果没有时
	// 随机选择一个位置
	for k := 1; k < matrixLength; k++ {
		for j := 0; j < len(rows[k]); j++ {
			point := rows[k][j]
			if point.value == 0 {
				fmt.Println("随机选择")
				return point.x, point.y
			}
		}
	}

	return -1, -1
}

// computeWeight :计算所有方向的权重
func computeWeight(data [][]Point, side Piece, x int, y int) int {
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
	data0 := data

	printMatrix(data0)

	// 计算横向
	for i := 0; i < len(data0); i++ {
		row := data0[i]
		b, w := computeRowWeight(row)
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
		b, w := computeRowWeight(row)
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
	data1 := rotateMatrix(data)
	printMatrix(data1)

	for i := 0; i < len(data1); i++ {
		row := data1[i]
		b, w := computeRowWeight(row)
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
		b, w := computeRowWeight(row)
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

	// 有四点已经一线了，判断是否有机会堵
	if otherWeight >= 10000 {

	}

	// 计算纵向
	// 计算斜向 45度
	// 计算逆斜向 -45度

	fmt.Println(fmt.Sprintf("weight:%d, other weight:%d", weight, otherWeight))
	return weight
}

// 旋转矩阵
func rotateMatrix(matrix [][]Point) [][]Point {
	r := make([][]Point, len(matrix))
	for i := 0; i < len(matrix); i++ {
		r[i] = make([]Point, len(matrix))
	}

	// transpose it
	for i := 0; i < len(matrix); i++ {
		for j := 0; j <= i; j++ {
			r[i][j], r[j][i] = matrix[j][i], matrix[i][j]
		}
	}
	return r
}

func reverse(x []Point) []Point {
	var r []Point
	for i := len(x) - 1; i >= 0; i-- {
		r = append(r, x[i])
	}

	return r
}

func rotate45Matrix(matrix [][]Point) [][]Point {
	var m [][]Point

	for line := 0; line < 2*len(matrix)-1; line++ {
		var r []Point
		for i := 0; i < len(matrix); i++ {
			for j := 0; j < len(matrix[i]); j++ {
				if i+j == line {
					r = append(r, matrix[i][j])
					break
				}
			}
		}

		r = reverse(r)
		//fmt.Printf("%#v\n\r", r)
		m = append(m, r)
	}
	return m
}

func rotateReverse45Matrix(matrix [][]Point) [][]Point {
	var m, m2 [][]Point

	// Mirror
	for i := 0; i < len(matrix); i++ {
		var newRow []Point
		for j := len(matrix[i]); j > 0; j-- {
			newRow = append(newRow, matrix[i][j-1])
		}
		m = append(m, newRow)
	}

	for line := 0; line < 2*len(m)-1; line++ {
		var r []Point
		for i := 0; i < len(m); i++ {
			for j := 0; j < len(m[i]); j++ {
				if i+j == line {
					r = append(r, m[i][j])
					break
				}
			}
		}

		m2 = append(m2, r)
	}
	return m2
}

type pieceTimeWeight map[int]int

type pieceWeight map[int]pieceTimeWeight

// judgeRowWin : 判断某行是否赢
func judgeRowWin(row []Point) (bool, Piece, []Point) {
	times := 0
	c := -1
	lastC := -1

	p := make([]Point, 5)

	for i := 0; i < len(row); i++ {
		c = row[i].value

		if lastC != c {
			if times >= 5 {
				for j := 0; j < 5; j++ {
					p[j] = row[(i-5)+j]
				}
				return true, Piece(lastC), p
			}
			times = 1
		} else {
			times++
		}
		lastC = c
	}
	return false, 0, nil
}

func printRows(rows [][]Point) {
	fmt.Println("print data rows:")
	for i := 0; i < len(rows); i++ {
		fmt.Printf("Line %2d : ", i)
		for j := 0; j < len(rows[i]); j++ {
			fmt.Printf(" %d ", rows[i][j].value)
		}
		fmt.Printf("\n\r")
	}
	fmt.Println("print data rows end.")
}

// computeRowWeight :计算单行权重
func computeRowWeight(row []Point) (int, int) {
	p := make(map[int]pieceTimeWeight)
	p[0] = make(map[int]int)
	p[1] = make(map[int]int)
	p[2] = make(map[int]int)

	times := 0
	spaces := 0
	c := -1
	lastC := -1
	lastCc := -1
	lastTimes := 0
	whiteWeight := 0
	blackWeight := 0
	data := row

	// 0, 0, 0, 0, 1, 1, 2, 0, 1, 2, 0
	// 1, 2, 1, 2, 0, 1, 1, 2, 2, 2, 1

	for i := 0; i < len(data); i++ {
		c = data[i].value

		if c != lastC {
			if lastC > 0 {
				_, ok := p[lastC][times]
				if !ok {
					p[lastC][times] = 0
				}

				if (spaces > 0) || (c == 0) {
					if lastCc == lastC || lastCc == -1 {
						times += lastTimes
						lastTimes = 0
						spaces = 0
					}
					if c == 0 && times > 1 {
						times++
					}
					fmt.Println("i:", i, "space: ", spaces, "times:", times, "c:", lastC, "next:", c)
					p[lastC][times]++
					spaces = 0
					lastCc = lastC
					lastTimes = times
				} else {
					times = 0
					lastTimes = 0
				}
			}
			times = 1
			if c == 0 {
				spaces = 1
			}
		} else {
			times++
			if c == 0 {
				spaces++
			}
		}

		lastC = c
	}

	if lastC > 0 && spaces > 0 {
		_, ok := p[lastC][times]
		if !ok {
			p[lastC][times] = 0
		}
		if lastCc == lastC || lastCc == -1 {
			times += lastTimes
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

var _gameData0 = GameData{
	data: [][]int{
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 6},
		{2, 2, 1, 2, 0, 1, 1, 2, 2, 2, 1},
		{3, 1, 1, 1, 0, 2, 2, 1, 2, 2, 0},
		{4, 1, 0, 1, 1, 2, 2, 2, 2, 1, 0},
		{5, 0, 1, 1, 0, 2, 2, 2, 0, 2, 1},
		{6, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{8, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0},
		{8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	}}

type GOMOKU_GAME_DATA struct {
	Data []int `json:"data" binding:"required"`
}

type GOMOKU_GAME_JUDGE_DATA struct {
	data [][]int
}

type Point struct {
	x     int
	y     int
	value int
}

var _gameData [][]Point

func initData() {
	for i := 0; i < len(_gameData0.data); i++ {
		var r []Point
		for j := 0; j < len(_gameData0.data[i]); j++ {
			p := Point{x: i, y: j, value: _gameData0.data[i][j]}
			r = append(r, p)
		}
		_gameData = append(_gameData, r)
	}
}

func printMatrix(data [][]Point) {
	fmt.Println("matrix begin")
	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data[i]); j++ {
			fmt.Printf("(%d,%d) %d ", data[i][j].x, data[i][j].y, data[i][j].value)
		}
		fmt.Println()
	}
	fmt.Println("matrix end")
	fmt.Println()
}

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func main() {
	fmt.Println("Gomoku Game Simple algorithm, code by shrek 2022")
	file, err := openLogFile("./mylog.log")
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	//initData()
	//data := _gameData
	//printMatrix(data)
	//
	//data2 := rotateReverse45Matrix(data)
	//fmt.Println("data2")
	//printMatrix(data2)
	//
	//data3 := rotate45Matrix(data)
	//fmt.Println("data3")
	//printMatrix(data3)
	//os.Exit(0)
	//initData()
	//data := _gameData
	//
	//compute(data, Black)

	//x, y := computeRowWeight(data[0])
	//fmt.Println("x=", x, "y=", y)
	//x, y = computeRowWeight(data[1])
	//fmt.Println("x=", x, "y=", y)
	//x, y = computeRowWeight(data[2])
	//fmt.Println("x=", x, "y=", y)
	//x, y = computeRowWeight(data[3])
	//fmt.Println("x=", x, "y=", y)
	//x, y = computeRowWeight(data[4])
	//fmt.Println("x=", x, "y=", y)
	//os.Exit(1)

	//v := computeWeight(data, Black, 0, 0)
	//fmt.Println(v)

	//os.Exit(0)

	r := gin.Default()
	r.GET("/api/v1/hello", func(c *gin.Context) {
		c.String(200, "Hello, Gomoku")
	})

	// POST
	r.POST("/api/nextstep", func(c *gin.Context) {
		json := GOMOKU_GAME_DATA{}
		c.BindJSON(&json)
		log.Printf("%v", &json)
		fmt.Println(json)

		x := nextStep(json.Data, 2)

		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"position": x,
		})
	})

	// POST
	r.POST("/api/GomokuWin", func(c *gin.Context) {
		json := GOMOKU_GAME_DATA{}
		c.BindJSON(&json)
		log.Printf("%v", &json)
		fmt.Println(json.Data)
		x, data := JudgeWin(json.Data)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"win":     x,
			"data":    data,
		})
	})
	r.Run()
}
