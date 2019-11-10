// 電通大で行われたコンピュータ囲碁講習会をGolangで追う
package main

import (
	// "bufio"
	"fmt"
	// "log"
	// "math"
	// "math/rand"
	// "os"
	// "sort"
	// "strconv"
	// "strings"
	// "sync"
	// "time"
	// "unicode"
	// "unsafe"
)


const (
	BoardSize  = 9
    Width    =  (BoardSize + 2)
	BoardMax  =(Width * Width)
)
var board=[BoardMax] int {
	3,3,3,3,3,3,3,3,3,3,3,
	3,2,1,1,0,1,0,0,0,0,3,
	3,2,2,1,1,0,1,2,0,0,3,
	3,2,0,2,1,2,2,1,1,0,3,
	3,0,2,2,2,1,1,1,0,0,3,
	3,0,0,0,2,1,2,1,0,0,3,
	3,0,0,2,0,2,2,1,0,0,3,
	3,0,0,0,0,2,1,1,0,0,3,
	3,0,0,0,0,2,2,1,0,0,3,
	3,0,0,0,0,0,2,1,0,0,3,
	3,3,3,3,3,3,3,3,3,3,3,
  }
//   var usi_koma_kanji = [20]string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九",
//   "十", "⑪", "⑫", "⑬", "⑭", "⑮", "⑯", "⑰", "⑱","⑲"}
  var usi_koma_kanji = [20]string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九",
  "❿", "⓫", "⓬", "⓭", "⓮", "⓯", "⓰", "⓱", "⓲","⓳"}

func PrintBoard(){
	var str=[4]string{"・","●","○","＃"}
	fmt.Printf("\n   ")
	for x := 0; x < BoardSize; x++ {
		fmt.Printf("%2d",x+1)}
	fmt.Printf("\n  +")
	for x := 0; x < BoardSize; x++ {
		fmt.Printf("--")}
		fmt.Printf("+\n")
		for y := 0; y < BoardSize; y++ {
		fmt.Printf("%s|", usi_koma_kanji[y+1])
		for x := 0; x < BoardSize; x++ {
			fmt.Printf("%s", str[board[x+1+Width* (y+1)]])
		}
		fmt.Printf("|\n")
	}
	fmt.Printf("  +")
	for x := 0; x < BoardSize; x++ {
		fmt.Printf("--")}
		fmt.Printf("+\n")
}

func main() {
	PrintBoard()
}
