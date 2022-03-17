// 電通大で行われたコンピュータ囲碁講習会をGolangで追う
package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	komi      = 6.5
	BoardSize = 19
	Width     = (BoardSize + 2)
	BoardMax  = (Width * Width)
	MaxMoves  = 1000
)

type Board struct {
	board [BoardMax]int
	ko_z  int
	moves int
}

var dir4 = [4]int{1, Width, -1, -Width}
var all_playouts int64
var flag_test_playout int
var record [MaxMoves]int

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
		return string(ax) + strconv.Itoa(BoardSize+1-y)
	}
}
func flip_color(col int) int {
	return 3 - col
}

func (b *Board) count_liberty_sub(tz int, color int, p_liberty *int, p_stone *int, check_board *[BoardMax]int) {
	check_board[tz] = 1
	*p_stone++
	for i := 0; i < 4; i++ {
		z := tz + dir4[i]
		if check_board[z] != 0 {
			continue
		}
		if b.board[z] == 0 {
			check_board[z] = 1
			*p_liberty++
		}
		if b.board[z] == color {
			b.count_liberty_sub(z, color, p_liberty, p_stone, check_board)
		}
	}

}
func (b *Board) count_liberty(tz int, p_liberty *int, p_stone *int) {
	*p_liberty = 0
	*p_stone = 0
	var check_board = [BoardMax]int{}
	for i := 0; i < BoardMax; i++ {
		check_board[i] = 0
	}
	b.count_liberty_sub(tz, b.board[tz], p_liberty, p_stone, &check_board)
}

func (b *Board) take_stone(tz int, color int) {
	b.board[tz] = 0
	for i := 0; i < 4; i++ {
		z := tz + dir4[i]
		if b.board[z] == color {
			b.take_stone(z, color)
		}
	}
}

const (
	FILL_EYE_ERR = 1
	FILL_EYE_OK  = 0
)

func (b *Board) put_stone(tz int, color int, fill_eye_err int) int {
	var around = [4][3]int{}
	var liberty, stone int
	un_col := flip_color(color)
	space := 0
	wall := 0
	mycol_safe := 0
	capture_sum := 0
	ko_maybe := 0

	if tz == 0 {
		b.ko_z = 0
		return 0
	}
	for i := 0; i < 4; i++ {
		around[i][0] = 0
		around[i][1] = 0
		around[i][2] = 0
		z := tz + dir4[i]
		c := b.board[z]
		if c == 0 {
			space++
		}
		if c == 3 {
			wall++
		}
		if c == 0 || c == 3 {
			continue
		}
		b.count_liberty(z, &liberty, &stone)
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
	if tz == b.ko_z {
		return 2
	}
	if wall+mycol_safe == 4 && fill_eye_err == FILL_EYE_ERR {
		return 3
	}
	if b.board[tz] != 0 {
		return 4
	}

	for i := 0; i < 4; i++ {
		lib := around[i][0]
		c := around[i][2]
		if c == un_col && lib == 1 && b.board[tz+dir4[i]] != 0 {
			b.take_stone(tz+dir4[i], un_col)
		}
	}

	b.board[tz] = color

	b.count_liberty(tz, &liberty, &stone)
	if capture_sum == 1 && stone == 1 && liberty == 1 {
		b.ko_z = ko_maybe
	} else {
		b.ko_z = 0
	}
	return 0
}

//   var usi_koma_kanji = [20]string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九",
//   "十", "⑪", "⑫", "⑬", "⑭", "⑮", "⑯", "⑰", "⑱","⑲"}
var usi_koma_kanji = [20]string{"　零", "　一", "　二", "　三", "　四", "　五", "　六", "　七", "　八", "　九",
	"　十", "十一", "十二", "十三", "十四", "十五", "十六", "十七", "十八", "十九"}

func (b *Board) PrintBoard() {
	var str = [4]string{"・", "●", "○", "＃"}
	fmt.Printf("\n　　 ")
	for x := 0; x < BoardSize; x++ {
		fmt.Printf("%2d", x+1)
	}
	fmt.Printf("\n　　+")
	for x := 0; x < BoardSize; x++ {
		fmt.Printf("--")
	}
	fmt.Printf("+\n")
	for y := 0; y < BoardSize; y++ {
		fmt.Printf("%s|", usi_koma_kanji[y+1])
		for x := 0; x < BoardSize; x++ {
			fmt.Printf("%s", str[b.board[x+1+Width*(y+1)]])
		}
		fmt.Printf("|")
		if y == 4 {
			fmt.Printf("  ko_z=%d,moves=%d", get81(b.ko_z), b.moves)
		}
		fmt.Printf("\n")
	}
	fmt.Printf("　　+")
	for x := 0; x < BoardSize; x++ {
		fmt.Printf("--")
	}
	fmt.Printf("+\n")
}

func (b *Board) count_score(turn_color int) int {
	var mk = [4]int{}
	var kind = [3]int{0, 0, 0}
	var score, black_area, white_area, black_sum, white_sum int

	for y := 0; y < BoardSize; y++ {
		for x := 0; x < BoardSize; x++ {
			z := get_z(x+1, y+1)
			c := b.board[z]
			kind[c]++
			if c != 0 {
				continue
			}
			mk[1] = 0
			mk[2] = 0
			for i := 0; i < 4; i++ {
				mk[b.board[z+dir4[i]]]++
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
		win = 1 - win
	} // gogo07

	// fmt.Printf("black_sum=%2d, (stones=%2d, area=%2d)\n", black_sum, kind[1], black_area)
	// fmt.Printf("white_sum=%2d, (stones=%2d, area=%2d)\n", white_sum, kind[2], white_area)
	// fmt.Printf("score=%d, win=%d\n", score, win)
	return win
}

func (b *Board) playout(turn_color int) int {
	color := turn_color
	previous_z := 0
	loop_max := BoardSize*BoardSize + 200

	// all_playouts++
	atomic.AddInt64(&all_playouts, 1)
	for loop := 0; loop < loop_max; loop++ {
		var empty = [BoardMax]int{}
		var empty_num, r, z int
		for y := 0; y <= BoardSize; y++ {
			for x := 0; x < BoardSize; x++ {
				z = get_z(x+1, y+1)
				if b.board[z] != 0 {
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
			err := b.put_stone(z, color, FILL_EYE_ERR)
			if err == 0 {
				break
			}
			empty[r] = empty[empty_num-1]
			empty_num--
		}
		if flag_test_playout != 0 {
			record[b.moves] = z
			b.moves++
		}
		if z == 0 && previous_z == 0 {
			break
		}
		previous_z = z
		// PrintBoard()
		// fmt.Printf("loop=%d,z=%d,c=%d,empty_num=%d,ko_z=%d\n",
		// 	loop, get81(z), color, empty_num, get81(ko_z))
		color = flip_color(color)
	}
	return b.count_score(turn_color)
}

func (b *Board) clone() *Board {
	// copyb := *b
	var copyb Board
	copyb.ko_z = b.ko_z
	copy(copyb.board[:], b.board[:])
	return &copyb
}

func (b *Board) primitive_monte_calro(color int) int {
	try_num := 8
	threads := 8
	best_z := 0
	var best_value, win_rate float64
	bc := b.clone()
	best_value = -100.0

	for y := 0; y <= BoardSize; y++ {
		for x := 0; x < BoardSize; x++ {
			z := get_z(x+1, y+1)
			if b.board[z] != 0 {
				continue
			}
			err := b.put_stone(z, color, FILL_EYE_ERR)
			if err != 0 {
				continue
			}

			win_sum := 0
			c := make(chan int)
			var wg sync.WaitGroup
			for i := 0; i < threads; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < try_num/threads; i++ {
						bc2 := b.clone()
						win := -bc2.playout(flip_color(color))
						c <- win
					}
				}()
				// win_sum += win
				// b.ko_z = bc2.ko_z
				// copy(b.board[:], bc2.board[:])
			}
			go func() {
				wg.Wait()
				close(c)
			}()
			// for win := range c {
			// 	win_sum += win
			// }
			for i := 0; i < try_num; i++ {
				win := <-c
				win_sum += win
			}
			win_rate = float64(win_sum) / float64(try_num)
			if win_rate > best_value {
				best_value = win_rate
				best_z = z
				// fmt.Printf("best_z=%d,color=%d,v=%5.3f,try_num=%d\n", get81(best_z), color, best_value, try_num)
			}
			b.ko_z = bc.ko_z
			copy(b.board[:], bc.board[:])
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

var mu2 sync.Mutex

func (b *Board) create_node() int {
	if node_num == NodeMax {
		fmt.Printf("node over Err\n")
		os.Exit(0)
	}
	mu2.Lock()
	pN := &node[node_num]
	pN.Child_num = 0
	pN.ChildGameSum = 0
	for y := 0; y <= BoardSize; y++ {
		for x := 0; x < BoardSize; x++ {
			z := get_z(x+1, y+1)
			if b.board[z] != 0 {
				continue
			}
			add_child(pN, z)
		}
	}
	add_child(pN, 0)
	node_num++
	mu2.Unlock()
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
		fmt.Printf("Err! select\n")
		os.Exit(0)
	}
	return select_i
}

var mu sync.Mutex

func (b *Board) search_uct(color int, node_n int) int {
	pN := &node[node_n]
	var c *Child
	var win int
	for {
		select_i := select_best_ucb(node_n)
		c = &pN.Children[select_i]
		z := c.Z
		if z < 0 {
			continue
		}
		err := b.put_stone(z, color, FILL_EYE_ERR)
		if err == 0 {
			break
		}
		c.Z = ILLEGAL_Z
		// fmt.Printf("ILLEGAL:z=%2d\n", get81(z))
	}
	if c.Games <= 0 {
		win = 1 - b.playout(flip_color(color))
	} else {
		if c.Next == NODE_EMPTY {
			c.Next = b.create_node()
		}
		win = 1 - b.search_uct(flip_color(color), c.Next)
	}
	mu.Lock()
	c.Rate = (c.Rate*float64(c.Games) + float64(win)) / float64(c.Games+1)
	c.Games++
	pN.ChildGameSum++
	mu.Unlock()
	return win
}

func (b *Board) get_best_uct(color int) int {
	max := -999
	node_num = 0
	// uct_loop := 3000 * 2 * 2
	uct_loop := 8000
	threads := 8
	var best_i = -1
	next := b.create_node()
	var wg sync.WaitGroup
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < uct_loop/threads; i++ {
				bc := b.clone()
				bc.search_uct(color, next)
			}
		}()
	}
	wg.Wait()
	pN := &node[next]
	for i := 0; i < pN.Child_num; i++ {
		c := &pN.Children[i]
		if c.Games > max {
			best_i = i
			max = c.Games
		}
		// fmt.Printf("%2d:z=%2d,rate=%.4f,games=%3d\n", i, get81(c.Z), c.Rate, c.Games)
	}
	best_z := pN.Children[best_i].Z
	fmt.Printf("best_z=%d,rate=%.4f,games=%d,playouts=%d,nodes=%d\n",
		get81(best_z), pN.Children[best_i].Rate, max, all_playouts, node_num)
	return best_z
}

func (b *Board) init_board() {
	for i := 0; i < BoardMax; i++ {
		b.board[i] = 3
	}
	for y := 0; y < BoardSize; y++ {
		for x := 0; x < BoardSize; x++ {
			b.board[get_z(x+1, y+1)] = 0
		}
	}
	b.moves = 0
	b.ko_z = 0
}

func (b *Board) add_moves(z int, color int) {
	err := b.put_stone(z, color, FILL_EYE_OK)
	if err != 0 {
		fmt.Printf("Err!\n")
		os.Exit(0)
	}
	record[b.moves] = z
	b.moves++
	// b.PrintBoard()
}

func (b *Board) get_computer_move(color int, fUCT int) int {
	var z int
	st := time.Now()
	all_playouts = 0
	if fUCT != 0 {
		z = b.get_best_uct(color)
	} else {
		z = b.primitive_monte_calro(color)
	}
	t := time.Since(st).Seconds()
	fmt.Printf("%.1f sec, %.0f playout/sec, play_z=%2d,moves=%d,color=%d,playouts=%d\n",
		t, float64(all_playouts)/t, get81(z), b.moves, color, all_playouts)
	return z
}
func undo() {

}
func (b *Board) print_sgf() {
	fmt.Printf("(;GM[1]SZ[%d]KM[%.1f]PB[]PW[]\n", BoardSize, komi)
	for i := 0; i < b.moves; i++ {
		z := record[i]
		y := z / Width
		x := z - y*Width
		var sStone = [2]string{"B", "W"}
		fmt.Printf(";%s", sStone[i&1])
		if z == 0 {
			fmt.Printf("[]")
		} else {
			fmt.Printf("[%c%c]", x+'a'-1, y+'a'-1)
		}
		if ((i + 1) % 10) == 0 {
			fmt.Printf("\n")
		}
	}
	fmt.Printf(")\n")
}
func selfplay() {
	color := 1
	var b Board
	b.init_board()

	for {
		fUCT := 1
		if color == 1 {
			fUCT = 0
		}
		z := b.get_computer_move(color, fUCT)
		b.add_moves(z, color)
		if z == 0 && b.moves > 1 && record[b.moves-2] == 0 {
			break
		}
		if b.moves > 300 {
			break
		} // too long
		color = flip_color(color)
	}
	fmt.Printf("Color: %d  Score: %d\n", color, b.count_score(color))
	b.print_sgf()
}

func test_playout() {
	flag_test_playout = 1
	var b Board
	b.init_board()
	b.playout(1)
	b.PrintBoard()
	b.print_sgf()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	// test_playout()
	// selfplay()
	var b Board
	b.init_board()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()
		str := strings.Split(command, " ")
		switch str[0] {
		case "boardsize":
			fmt.Printf("= \n\n")
		case "clear_board":
			b.init_board()
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
			z := b.get_computer_move(color, 1)
			b.add_moves(z, color)
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
			// y := int(ax[1] - '0')
			y, _ := strconv.Atoi(ax[1:])
			z := get_z(int(x), BoardSize-y+1)
			fmt.Fprintf(os.Stderr, "x=%d y=%d z=%d\n", x, y, get81(z))
			if ax == "pass" {
				z = 0
			}
			b.add_moves(z, color)
			fmt.Printf("= \n\n")
		default:
			fmt.Printf("? unknown_command\n\n")
		}
	}
}
