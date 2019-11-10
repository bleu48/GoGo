// 電通大で行われたコンピュータ囲碁講習会をGolangで追う
package main

import (
	// "bufio"
	"fmt"
	// "log"
	// "math"
	"math/rand"
	// "os"
	// "sort"
	// "strconv"
	// "strings"
	// "sync"
	"time"
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
	3,0,0,0,0,0,0,0,0,0,3,
	3,0,0,0,0,0,0,0,0,0,3,
	3,0,0,0,0,0,0,0,0,0,3,
	3,0,0,0,0,0,0,0,0,0,3,
	3,0,0,0,0,0,0,0,0,0,3,
	3,0,0,0,0,0,0,0,0,0,3,
	3,0,0,0,0,0,0,0,0,0,3,
	3,0,0,0,0,0,0,0,0,0,3,
	3,0,0,0,0,0,0,0,0,0,3,
	3,3,3,3,3,3,3,3,3,3,3,
	  }

  var dir4=[4]int{1,Width,-1,-Width}
  var ko_z int;
  func get_z(x int,y int)int{
	  return y*Width+x
  }
  func get81(z int)int{
	  y:=z/Width
	  x:=z-y*Width
	  if (z==0){return 0}
	  return x*10+y
  }
  func flip_color(col int)int{
	  return 3-col
  }
  var check_board=[BoardMax]int{}

  func count_liberty_sub(tz int,color int,p_liberty *int,p_stone *int){
	  check_board[tz]=1
	  *p_stone++
	  for i:=0;i<4;i++{
		  z:=tz+dir4[i]
		  if check_board[z]!=0 {
			  continue
		  }
		  if board[z]==0{
			  check_board[z]=1
			  *p_liberty++
		  }
		  if board[z]==color{
			  count_liberty_sub(z,color,p_liberty,p_stone)
		  }
	  }

  }
  func count_liberty(tz int,p_liberty *int,p_stone *int){
	  *p_liberty=0
	  *p_stone=0
	  for i:=0;i<BoardMax;i++{
		  check_board[i]=0
	  }
	  count_liberty_sub(tz,board[tz],p_liberty,p_stone)
  }

  func take_stone(tz int,color int){
	  board[tz]=0
	  for i:=0;i<4;i++{
		  z:=tz+dir4[i]
		  if board[z]== color{take_stone(z,color)}
	  }
  }
  func put_stone(tz int,color int)int{
	  var around=[4][3]int{}
	  var liberty,stone int
	  un_col:=flip_color(color)
	  space:=0
	  wall:=0
	  mycol_safe:=0
	  capture_sum:=0
	  ko_maybe:=0

	  if tz==0{
		  ko_z=0
		  return 0
	  }
	  for i:=0;i<4;i++{
		around[i][0]=0
		around[i][1]=0
		around[i][2]=0
		z:=tz+dir4[i]
		c:=board[z]
		if c==0 {space++}
		if c==3 {wall++}
		if c==0 || c==3{continue}
		count_liberty(z,&liberty,&stone)
		around[i][0]=liberty
		around[i][1]=stone
		around[i][2]=c
		if c==un_col && liberty==1 {
			capture_sum+=stone
			ko_maybe=z
		}
	if c==color && liberty >=2{mycol_safe++}

	}
	if capture_sum==0 && space ==0 && mycol_safe==0 {return 1}
	if tz== ko_z{return 2}
	if wall+mycol_safe==4 {return 3}
	if board[tz]!=0 {return 4}

	for i:=0;i<4;i++{
		lib:=around[i][0]
		c:=around[i][2]
		if c==un_col && lib==1 && board[tz+dir4[i] ]!=0{
			take_stone(tz+dir4[i],un_col)
		}
	}

	board[tz]=color

	count_liberty(tz,&liberty,&stone)
	if capture_sum==1 && stone==1 && liberty==1 {
		ko_z=ko_maybe
	}	else {ko_z=0}
	return 0
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

func get_empty_z()int{
	var x,y,z int
	for{
		x=rand.Intn(9)+1
		y=rand.Intn(9)+1
		z=get_z(x,y)
		if board[z]==0 {break}
	}
	return z
}

func play_one_move(color int)int{
	var z int
	for i:=0;i<100;i++{
		z:=get_empty_z()
		err:=put_stone(z,color)
		if err==0{return z}
	}
	z=0
	put_stone(0,color)
	return z
}

var moves int
var record [1000]int
func main() {
	color:=1
	rand.Seed(time.Now().UnixNano())
	for{
		z:=play_one_move(color)
		fmt.Printf("moves=%4d, color=%d, z=%d\n",moves, color, get81(z));
		PrintBoard()

		record[moves]=z 
		moves++
		if moves==1000{
			fmt.Printf("max moves!\n");
			break
		}
		if z==0 && moves >=2 && record[moves-2]==0 {
			fmt.Printf("two pass\n");
			break
		}
		color=flip_color(color)

	}

}
