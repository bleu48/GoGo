// 電通大で行われたコンピュータ囲碁講習会をGolangで追う
package main

import (
	// "bufio"
	"bufio"
	"fmt"
	"math"
	"strings"

	// "log"

	"math/rand"
	"os"

	// "sort"
	// "strconv"
	// "strings"
	// "sync"
	"time"
	// "unicode"
	// "unsafe"
)

const (
	komi      = 6.5
	BoardSize = 9
	Width     = (BoardSize + 2)
	BoardMax  = (Width * Width)
	MaxMoves  = 1000
)

var board = [BoardMax]int{
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3,
	3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3,
	3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3,
	3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3,
	3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3,
	3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3,
	3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3,
	3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3,
	3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
}

var dir4 = [4]int{1, Width, -1, -Width}
var ko_z int
var moves, all_playouts, flag_test_playout int
var record [MaxMoves]int
var record_time [MaxMoves]float64

func get_z(x int, y int) int {
	return y*Width + x
}
func get81(z int) int {
	y := z / Width
	x := z - y*Width
	if z == 0 {
		return 0
	}
	return x*10 + y
}
func get_char_z(z int) string {
	if z == 0 {
		return "pass"
	} else {
		y := z / Width
		x := z - y*Width
		ax := 'A' + x - 1
		if ax >= 'I' {
			ax++
		}
		return string(ax) + string(BoardSize+1-y+'0')
	}
}
func flip_color(col int) int {
	return 3 - col
}

var check_board = [BoardMax]int{}

func count_liberty_sub(tz int, color int, p_liberty *int, p_stone *int) {
	check_board[tz] = 1
	*p_stone++
	for i := 0; i < 4; i++ {
		z := tz + dir4[i]
		if check_board[z] != 0 {
			continue
		}
		if board[z] == 0 {
			check_board[z] = 1
			*p_liberty++
		}
		if board[z] == color {
			count_liberty_sub(z, color, p_liberty, p_stone)
		}
	}

}
func count_liberty(tz int, p_liberty *int, p_stone *int) {
	*p_liberty = 0
	*p_stone = 0
	for i := 0; i < BoardMax; i++ {
		check_board[i] = 0
	}
	count_liberty_sub(tz, board[tz], p_liberty, p_stone)
}

func take_stone(tz int, color int) {
	board[tz] = 0
	for i := 0; i < 4; i++ {
		z := tz + dir4[i]
		if board[z] == color {
			take_stone(z, color)
		}
	}
}

const (
	FILL_EYE_ERR = 1
	FILL_EYE_OK  = 0
)

func put_stone(tz int, color int, fill_eye_err int) int {
	var around = [4][3]int{}
	var liberty, stone int
	un_col := flip_color(color)
	space := 0
	wall := 0
	mycol_safe := 0
	capture_sum := 0
	ko_maybe := 0

	if tz == 0 {
		ko_z = 0
		return 0
	}
	for i := 0; i < 4; i++ {
		around[i][0] = 0
		around[i][1] = 0
		around[i][2] = 0
		z := tz + dir4[i]
		c := board[z]
		if c == 0 {
			space++
		}
		if c == 3 {
			wall++
		}
		if c == 0 || c == 3 {
			continue
		}
		count_liberty(z, &liberty, &stone)
		around[i][0] = liberty
		around[i][1] = stone
		around[i][2] = c
		if c == un_col && liberty == 1 {
			capture_sum += stone
			ko_maybe = z
		}
		if c == color && liberty >= 2 {
			mycol_safe++
		}

	}
	if capture_sum == 0 && space == 0 && mycol_safe == 0 {
		return 1
	}
	if tz == ko_z {
		return 2
	}
	if wall+mycol_safe == 4 && fill_eye_err == FILL_EYE_ERR {
		return 3
	}
	if board[tz] != 0 {
		return 4
	}

	for i := 0; i < 4; i++ {
		lib := around[i][0]
		c := around[i][2]
		if c == un_col && lib == 1 && board[tz+dir4[i]] != 0 {
			take_stone(tz+dir4[i], un_col)
		}
	}

	board[tz] = color

	count_liberty(tz, &liberty, &stone)
	if capture_sum == 1 && stone == 1 && liberty == 1 {
		ko_z = ko_maybe
	} else {
		ko_z = 0
	}
	return 0
}

//   var usi_koma_kanji = [20]string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九",
//   "十", "⑪", "⑫", "⑬", "⑭", "⑮", "⑯", "⑰", "⑱","⑲"}
// var usi_koma_kanji = [20]string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九",
// 	"❿", "⓫", "⓬", "⓭", "⓮", "⓯", "⓰", "⓱", "⓲", "⓳"}
var usi_koma_kanji = [20]string{" 0", " 1", " 2", " 3", " 4", " 5", " 6", " 7", " 8", " 9",
	"❿", "⓫", "⓬", "⓭", "⓮", "⓯", "⓰", "⓱", "⓲", "⓳"}

func PrintBoard() {
	// var str = [4]string{"・", "●", "○", "＃"}
	var str = [4]string{" .", " *", " o", " #"}
	fmt.Fprintf(os.Stderr, "\n   ")
	for x := 0; x < BoardSize; x++ {
		fmt.Fprintf(os.Stderr, "%2d", x+1)
	}
	fmt.Fprintf(os.Stderr, "\n  +")
	for x := 0; x < BoardSize; x++ {
		fmt.Fprintf(os.Stderr, "--")
	}
	fmt.Fprintf(os.Stderr, "+\n")
	for y := 0; y < BoardSize; y++ {
		fmt.Fprintf(os.Stderr, "%s|", usi_koma_kanji[y+1])
		for x := 0; x < BoardSize; x++ {
			fmt.Fprintf(os.Stderr, "%s", str[board[x+1+Width*(y+1)]])
		}
		fmt.Fprintf(os.Stderr, "|")
		if y == 4 {
			fmt.Fprintf(os.Stderr, "  ko_z=%d,moves=%d", get81(ko_z), moves)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}
	fmt.Fprintf(os.Stderr, "  +")
	for x := 0; x < BoardSize; x++ {
		fmt.Fprintf(os.Stderr, "--")
	}
	fmt.Fprintf(os.Stderr, "+\n")
}

func count_score(turn_color int) int {
	var mk = [4]int{}
	var kind = [3]int{0, 0, 0}
	var score, black_area, white_area, black_sum, white_sum int

	for y := 0; y < BoardSize; y++ {
		for x := 0; x < BoardSize; x++ {
			z := get_z(x+1, y+1)
			c := board[z]
			kind[c]++
			if c != 0 {
				continue
			}
			mk[1] = 0
			mk[2] = 0
			for i := 0; i < 4; i++ {
				mk[board[z+dir4[i]]]++
			}
			if mk[1] != 0 && mk[2] == 0 {
				black_area++
			}
			if mk[2] != 0 && mk[1] == 0 {
				white_area++
			}
		}
	}
	black_sum = kind[1] + black_area
	white_sum = kind[2] + white_area
	score = black_sum - white_sum
	win := 0
	if float64(score)-komi > 0 {
		win = 1
	}
	if turn_color == 2 {
		win = -win
	} // gogo07

	// fmt.Fprintf(os.Stderr,"black_sum=%2d, (stones=%2d, area=%2d)\n", black_sum, kind[1], black_area)
	// fmt.Fprintf(os.Stderr,"white_sum=%2d, (stones=%2d, area=%2d)\n", white_sum, kind[2], white_area)
	// fmt.Fprintf(os.Stderr,"score=%d, win=%d\n", score, win)
	return win
}

func playout(turn_color int) int {
	color := turn_color
	previous_z := 0
	loop_max := BoardSize*BoardSize + 200

	all_playouts++
	for loop := 0; loop < loop_max; loop++ {
		var empty = [BoardMax]int{}
		var empty_num, r, z int
		for y := 0; y <= BoardSize; y++ {
			for x := 0; x < BoardSize; x++ {
				z = get_z(x+1, y+1)
				if board[z] != 0 {
					continue
				}
				empty[empty_num] = z
				empty_num++
			}
		}
		r = 0
		for {
			if empty_num == 0 {
				z = 0
			} else {
				r = rand.Intn(empty_num)
				z = empty[r]
			}
			err := put_stone(z, color, FILL_EYE_ERR)
			if err == 0 {
				break
			}
			empty[r] = empty[empty_num-1]
			empty_num--
		}
		if flag_test_playout != 0 {
			record[moves] = z
			moves++
		}
		if z == 0 && previous_z == 0 {
			break
		}
		previous_z = z
		// PrintBoard()
		// fmt.Fprintf(os.Stderr,"loop=%d,z=%d,c=%d,empty_num=%d,ko_z=%d\n",
		// 	loop, get81(z), color, empty_num, get81(ko_z))
		color = flip_color(color)
	}
	return count_score(turn_color)
}

func primitive_monte_calro(color int) int {
	try_num := 30
	best_z := 0
	var best_value, win_rate float64
	var board_copy = [BoardMax]int{}
	ko_z_copy := ko_z
	copy(board_copy[:], board[:])
	best_value = -100.0

	for y := 0; y <= BoardSize; y++ {
		for x := 0; x < BoardSize; x++ {
			z := get_z(x+1, y+1)
			if board[z] != 0 {
				continue
			}
			err := put_stone(z, color, FILL_EYE_ERR)
			if err != 0 {
				continue
			}

			win_sum := 0
			for i := 0; i < try_num; i++ {
				var board_copy2 = [BoardMax]int{}
				ko_z_copy2 := ko_z
				copy(board_copy2[:], board[:])
				win := -playout(flip_color(color))
				win_sum += win
				ko_z = ko_z_copy2
				copy(board[:], board_copy2[:])
			}
			win_rate = float64(win_sum) / float64(try_num)
			if win_rate > best_value {
				best_value = win_rate
				best_z = z
				// fmt.Fprintf(os.Stderr,"best_z=%d,color=%d,v=%5.3f,try_num=%d\n", get81(best_z), color, best_value, try_num)
			}
			ko_z = ko_z_copy
			copy(board[:], board_copy[:])
		}
	}
	return best_z
}

// UCT
const (
	Childrenize = BoardSize*BoardSize + 1
	NodeMax     = 10000
	NODE_EMPTY  = -1
	ILLEGAL_Z   = -1
)

type Child struct {
	Z     int
	Games int
	Rate  float64
	Next  int
}
type Node struct {
	Child_num    int
	Children     [Childrenize]Child
	ChildGameSum int
}

var node = [NodeMax]Node{}
var node_num = 0

func add_child(pN *Node, z int) {
	n := pN.Child_num
	pN.Children[n].Z = z
	pN.Children[n].Games = 0
	pN.Children[n].Rate = 0.0
	pN.Children[n].Next = NODE_EMPTY
	pN.Child_num++
}

func create_node() int {
	if node_num == NodeMax {
		fmt.Fprintf(os.Stderr, "node over Err\n")
		os.Exit(0)
	}
	pN := &node[node_num]
	pN.Child_num = 0
	pN.ChildGameSum = 0
	for y := 0; y <= BoardSize; y++ {
		for x := 0; x < BoardSize; x++ {
			z := get_z(x+1, y+1)
			if board[z] != 0 {
				continue
			}
			add_child(pN, z)
		}
	}
	add_child(pN, 0)
	node_num++
	return node_num - 1
}

func select_best_ucb(node_n int) int {
	pN := &node[node_n]
	select_i := -1
	max_ucb := -999.0
	ucb := 0.0
	for i := 0; i < pN.Child_num; i++ {
		c := &pN.Children[i]
		if c.Z == ILLEGAL_Z {
			continue
		}
		if c.Games == 0 {
			ucb = 10000.0 + 32768.0*rand.Float64()
		} else {
			ucb = c.Rate + 1.0*math.Sqrt(math.Log(float64(pN.ChildGameSum))/float64(c.Games))
		}
		if ucb > max_ucb {
			max_ucb = ucb
			select_i = i
		}
	}
	if select_i == -1 {
		fmt.Fprintf(os.Stderr, "Err! select\n")
		os.Exit(0)
	}
	return select_i
}

func search_uct(color int, node_n int) int {
	pN := &node[node_n]
	var c *Child
	var win int
	for {
		select_i := select_best_ucb(node_n)
		c = &pN.Children[select_i]
		z := c.Z
		err := put_stone(z, color, FILL_EYE_ERR)
		if err == 0 {
			break
		}
		c.Z = ILLEGAL_Z
		// fmt.Fprintf(os.Stderr,"ILLEGAL:z=%2d\n", get81(z))
	}
	if c.Games <= 0 {
		win = -playout(flip_color(color))
	} else {
		if c.Next == NODE_EMPTY {
			c.Next = create_node()
		}
		win = -search_uct(flip_color(color), c.Next)
	}
	c.Rate = (c.Rate*float64(c.Games) + float64(win)) / float64(c.Games+1)
	c.Games++
	pN.ChildGameSum++
	return win
}

func get_best_uct(color int) int {
	max := -999
	node_num = 0
	uct_loop := 10000
	var best_i = -1
	next := create_node()
	for i := 0; i < uct_loop; i++ {
		var board_copy = [BoardMax]int{}
		ko_z_copy := ko_z
		copy(board_copy[:], board[:])

		search_uct(color, next)

		ko_z = ko_z_copy
		copy(board[:], board_copy[:])
	}
	pN := &node[next]
	for i := 0; i < pN.Child_num; i++ {
		c := &pN.Children[i]
		if c.Games > max {
			best_i = i
			max = c.Games
		}
		// fmt.Fprintf(os.Stderr,"%2d:z=%2d,rate=%.4f,games=%3d\n", i, get81(c.Z), c.Rate, c.Games)
	}
	best_z := pN.Children[best_i].Z
	fmt.Fprintf(os.Stderr, "best_z=%d,rate=%.4f,games=%d,playouts=%d,nodes=%d\n",
		get81(best_z), pN.Children[best_i].Rate, max, all_playouts, node_num)
	return best_z
}

func init_board() {
	for i := 0; i < BoardMax; i++ {
		board[i] = 3
	}
	for y := 0; y < BoardSize; y++ {
		for x := 0; x < BoardSize; x++ {
			board[get_z(x+1, y+1)] = 0
		}
	}
	moves = 0
	ko_z = 0
}

func add_moves(z int, color int, sec float64) {
	err := put_stone(z, color, FILL_EYE_OK)
	if err != 0 {
		fmt.Fprintf(os.Stderr, "Err!\n")
		os.Exit(0)
	}
	record[moves] = z
	record_time[moves] = sec
	moves++
	PrintBoard()
}

func play_computer_move(color int, fUCT int) int {
	var z int
	st := time.Now()
	all_playouts = 0
	if fUCT != 0 {
		z = get_best_uct(color)
	} else {
		z = primitive_monte_calro(color)
	}
	t := time.Since(st).Seconds()
	fmt.Fprintf(os.Stderr, "%.1f sec, %.0f playout/sec, play_z=%2d,moves=%d,color=%d,playouts=%d\n",
		t, float64(all_playouts)/t, get81(z), moves, color, all_playouts)
	add_moves(z, color, t)
	return z
}
func undo() {

}
func print_sgf() {
	fmt.Fprintf(os.Stderr, "(;GM[1]SZ[%d]KM[%.1f]PB[]PW[]\n", BoardSize, komi)
	for i := 0; i < moves; i++ {
		z := record[i]
		y := z / Width
		x := z - y*Width
		var sStone = [2]string{"B", "W"}
		fmt.Fprintf(os.Stderr, ";%s", sStone[i&1])
		if z == 0 {
			fmt.Fprintf(os.Stderr, "[]")
		} else {
			fmt.Fprintf(os.Stderr, "[%c%c]", x+'a'-1, y+'a'-1)
		}
		if ((i + 1) % 10) == 0 {
			fmt.Fprintf(os.Stderr, "\n")
		}
	}
	fmt.Fprintf(os.Stderr, ")\n")
}
func selfplay() {
	color := 1

	for {
		fUCT := 1
		if color == 1 {
			fUCT = 0
		}
		z := play_computer_move(color, fUCT)
		if z == 0 && moves > 1 && record[moves-2] == 0 {
			break
		}
		if moves > 300 {
			break
		} // too long
		color = flip_color(color)
	}

	print_sgf()
}

func test_playout() {
	flag_test_playout = 1
	playout(1)
	PrintBoard()
	print_sgf()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	init_board()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()
		str := strings.Split(command, " ")
		switch str[0] {
		case "boardsize":
			fmt.Printf("= \n\n")
		case "clear_board":
			init_board()
			fmt.Printf("= \n\n")
		case "quit":
			os.Exit(0)
		case "protocol_version":
			fmt.Printf("= 2\n\n")
		case "name":
			fmt.Printf("= GoGo\n\n")
		case "version":
			fmt.Printf("= 0.0.1\n\n")
		case "list_commands":
			fmt.Printf("= boardsize\nclear_board\nquit\nprotocol_version\nundo\n" +
				"name\nversion\nlist_commands\nkomi\ngenmove\nplay\n\n")
		case "komi":
			fmt.Printf("= 6.5\n\n")
		case "undo":
			undo()
			fmt.Printf("= \n\n")
		case "genmove":
			color := 1
			if strings.ToLower(str[1]) == "w" {
				color = 2
			}
			z := play_computer_move(color, 1)
			fmt.Printf("= %s\n\n", get_char_z(z))
		case "play":
			color := 1
			if strings.ToLower(str[1]) == "w" {
				color = 2
			}
			ax := strings.ToLower(str[2])
			fmt.Fprintf(os.Stderr, "ax=%s\n", ax)
			x := ax[0] - 'a' + 1
			if ax[0] >= 'i' {
				x--
			}
			y := int(ax[1] - '0')
			z := get_z(int(x), BoardSize-y+1)
			fmt.Fprintf(os.Stderr, "x=%d y=%d z=%d\n", x, y, get81(z))
			if ax == "pass" {
				z = 0
			}
			add_moves(z, color, 0)
			fmt.Printf("= \n\n")
		default:
			fmt.Printf("? unknown_command\n\n")
		}
	}
}
