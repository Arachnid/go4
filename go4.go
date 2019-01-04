package main

import (
    "bufio"
    "fmt"
    "os"
    "errors"
    "strings"
    "time"
)

const WIDTH uint = 7
const HEIGHT uint = 6
const MIN_SCORE = -int(WIDTH*HEIGHT)/2 + 3
const MAX_SCORE = int(WIDTH*HEIGHT+1)/2 - 3
// BOTTOM_MASK is a bitmask containing one for the bottom slot of each colum
const BOTTOM_MASK = 1 | (1 << 7) | (1 << 14) | (1 << 21) | (1 << 28) | (1 << 35) | (1 << 42)
const BOARD_MASK = BOTTOM_MASK * ((1 << HEIGHT) - 1)

var COLUMN_ORDER [WIDTH]uint = [...]uint{3, 2, 4, 1, 5, 0, 6}

func topMaskCol(col uint) uint64 {
    return (uint64(1) << (HEIGHT - 1)) << (col * (HEIGHT + 1))
}

func bottomMaskCol(col uint) uint64 {
    return uint64(1) << (col * (HEIGHT + 1))
}

func columnMask(col uint) uint64 {
    return ((uint64(1) << HEIGHT) - 1) << (col * (HEIGHT + 1))
}

func alignment(pos uint64) bool {
    // Horizontal
    m := pos & (pos >> (HEIGHT + 1))
    if (m & (m >> (2 * (HEIGHT + 1)))) != 0 {
        return true
    }

    // Diagonal 1
    m = pos & (pos >> HEIGHT)
    if (m & (m >> (2 * HEIGHT))) != 0 {
        return true
    }

    // Diagonal 2
    m = pos & (pos >> (HEIGHT + 2))
    if (m & (m >> (2 * (HEIGHT + 2)))) != 0 {
        return true
    }

    // Vertical
    m = pos & (pos >> 1)
    if (m & (m >> 2)) != 0 {
        return true
    }

    return false
}

/* Position represents a Connect 4 position.
 * Functions are relative to the current player to play.
 * Position containing aligment are not supported by this class.
 *
 * A binary bitboard representationis used.
 * Each column is encoded on HEIGHT+1 bits.
 *
 * Example of bit order to encode for a 7x6 board
 * .  .  .  .  .  .  .
 * 5 12 19 26 33 40 47
 * 4 11 18 25 32 39 46
 * 3 10 17 24 31 38 45
 * 2  9 16 23 30 37 44
 * 1  8 15 22 29 36 43
 * 0  7 14 21 28 35 42
 *
 * Position is stored as
 * - a bitboard "mask" with 1 on any color stones
 * - a bitboard "current_player" with 1 on stones of current player
 *
 * "current_player" bitboard can be transformed into a compact and non ambiguous key
 * by adding an extra bit on top of the last non empty cell of each column.
 * This allow to identify all the empty cells whithout needing "mask" bitboard
 *
 * current_player "x" = 1, opponent "o" = 0
 * board     position  mask      key       bottom
 *           0000000   0000000   0000000   0000000
 * .......   0000000   0000000   0001000   0000000
 * ...o...   0000000   0001000   0010000   0000000
 * ..xx...   0011000   0011000   0011000   0000000
 * ..ox...   0001000   0011000   0001100   0000000
 * ..oox..   0000100   0011100   0000110   0000000
 * ..oxxo.   0001100   0011110   1101101   1111111
 *
 * current_player "o" = 1, opponent "x" = 0
 * board     position  mask      key       bottom
 *           0000000   0000000   0001000   0000000
 * ...x...   0000000   0001000   0000000   0000000
 * ...o...   0001000   0001000   0011000   0000000
 * ..xx...   0000000   0011000   0000000   0000000
 * ..ox...   0010000   0011000   0010100   0000000
 * ..oox..   0011000   0011100   0011010   0000000
 * ..oxxo.   0010010   0011110   1110011   1111111
 *
 * key is an unique representation of a board key = position + mask + bottom
 * in practice, as bottom is constant, key = position + mask is also a
 * non-ambigous representation of the position.
 */
type Position struct {
    mask uint64
    currentPosition uint64
    moves uint
}

func (p Position) Play(col uint) Position {
    p.currentPosition ^= p.mask
    p.mask |= p.mask + bottomMaskCol(col)
    p.moves++
    return p
}

func computeWinningPosition(position uint64, mask uint64) uint64 {
    //vertical
    r := (position << 1) & (position << 2) & (position << 3)

    //horizontal
    p := (position << (HEIGHT+1)) & (position << (2*(HEIGHT+1)))
    r |= p & (position << (3*(HEIGHT+1)))
    r |= p & (position >> (HEIGHT+1))
    p = (position >> (HEIGHT+1)) & (position >> (2*(HEIGHT+1)))
    r |= p & (position << (HEIGHT+1))
    r |= p & (position >> (3*(HEIGHT+1)))

    //diagonal 1
    p = (position << HEIGHT) & (position << (2*HEIGHT))
    r |= p & (position << (3*HEIGHT))
    r |= p & (position >> HEIGHT)
    p = (position >> HEIGHT) & (position >> (2*HEIGHT))
    r |= p & (position << HEIGHT)
    r |= p & (position >> (3*HEIGHT))

    //diagonal 2
    p = (position << (HEIGHT+2)) & (position << (2*(HEIGHT+2)))
    r |= p & (position << (3*(HEIGHT+2)))
    r |= p & (position >> (HEIGHT+2))
    p = (position >> (HEIGHT+2)) & (position >> (2*(HEIGHT+2)))
    r |= p & (position << (HEIGHT+2))
    r |= p & (position >> (3*(HEIGHT+2)))

    return r & (BOARD_MASK ^ mask)
}

func (p Position) WinningPosition() uint64 {
    return computeWinningPosition(p.currentPosition, p.mask)
}

func (p Position) OpponentWinningPosition() uint64 {
    return computeWinningPosition(p.currentPosition ^ p.mask, p.mask)
}

func (p Position) Possible() uint64 {
    return (p.mask + BOTTOM_MASK) & BOARD_MASK
}

func (p Position) CanWinNext() bool {
    return p.WinningPosition() & p.Possible() != 0
}

func (p Position) IsWinningMove(col uint) bool {
    return p.WinningPosition() & p.Possible() & columnMask(col) != 0
}

func (p Position) CanPlay(col uint) bool {
    return (p.mask & topMaskCol(col)) == 0
}

func (p Position) MoveCount() uint {
    return p.moves
}

/*
 * Return a bitmap of all the possible next moves the do not lose in one turn.
 * A losing move is a move leaving the possibility for the opponent to win directly.
 *
 * Warning this function is intended to test position where you cannot win in one turn
 * If you have a winning move, this function can miss it and prefer to prevent the opponent
 * to make an alignment.
 */
func (p Position) NonLosingMoves() uint64 {
    possible_mask := p.Possible()
    opponent_win := p.OpponentWinningPosition()
    forced_moves := possible_mask & opponent_win
    if forced_moves != 0 {
        if forced_moves & (forced_moves - 1) != 0 {
            // Check if there is more than one forced move
            return 0
        } else {
            possible_mask = forced_moves
        }
    }
    // Avoid playing below an opponent's winning spot
    return possible_mask & ^(opponent_win >> 1)
}

func (p Position) Key() uint64 {
    k := p.currentPosition + p.mask
    k2 := ((k & 0x7f) << 42) | ((k & 0x3f80) << 28) | ((k & 0x1fc000) << 14) | (k & 0xfe00000) | ((k & 0x7f0000000) >> 14) | ((k & 0x3f800000000) >> 28) | ((k & 0x1fc0000000000) >> 42)
    if k2 > k {
        return k
    } else {
        return k2
    }
}

func (p Position) String() string {
    ret := ""
    for y := HEIGHT; y > 0; y-- {
        for x := uint(0); x < WIDTH; x++ {
            pos := uint64(1) << (x * (HEIGHT + 1) + y - 1)
            if pos & p.mask != 0 {
                if pos & p.currentPosition == 0 {
                    ret += "o"
                } else {
                    ret += "x"
                }
            } else {
                ret += "."
            }
        }
        ret += "\n"
    }
    return ret
}

// Evaluates the value of this position from the point of view of the last player
// to make a move (*not* the current player). The returned score is the number of
// winning positions available to the player.
func (p Position) Score() int {
    return popcount(computeWinningPosition(p.currentPosition ^ p.mask, p.mask))
}

func popcount(m uint64) int {
    c := 0
    for ; m != 0; c++ {
        m &= m - 1
    }
    return c
}

type PositionListEntry struct {
    p Position
    s int
}
type PositionList []PositionListEntry

func (l PositionList) Len() int { return len(l) }
func (l PositionList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l PositionList) Less(i, j int) bool { return l[j].s < l[i].s }

// TranspositionTable encodes a transposition table with up to 64 byte keys and 8 byte values.
// To prevent collisions, the length of the table must be coprime with 2^32, and the key must
// be no larger than len(t) * 2^32.
type TranspositionTable struct {
    keys []uint32
    values []uint8
}

var KeyOutOfRange error = errors.New("Transposition table key out of range")

func (t TranspositionTable) Put(key uint64, val uint8) error {
    if key >= uint64(len(t.keys)) * 0x100000000 {
        return KeyOutOfRange
    }
    // key is possibly trucated, but still unique as long as len(t) is coprime with 2^32.
    idx := key % uint64(len(t.keys))
    t.keys[idx] = uint32(key)
    t.values[idx] = val
    return nil
}

func (t TranspositionTable) Get(key uint64) (uint8, error) {
    if key >= uint64(len(t.keys)) * 0x100000000 {
        return 0, KeyOutOfRange
    }
    idx := key % uint64(len(t.keys))
    if t.keys[idx] != uint32(key) {
        return 0, nil
    }
    return t.values[idx], nil
}

func NewTranspositionTable(size uint) *TranspositionTable {
    return &TranspositionTable{
        keys: make([]uint32, size),
        values: make([]uint8, size),
    }
}

func negamax(transTable *TranspositionTable, p Position, alpha, beta int) (int, int) {
    next := p.NonLosingMoves()
    if next == 0 {
        return -int(WIDTH*HEIGHT - p.MoveCount()) / 2, 1
    }

    if p.MoveCount() >= WIDTH * HEIGHT - 2 {
        return 0, 1;
    }

    min := -int(WIDTH * HEIGHT - 2 - p.MoveCount()) / 2 // lower bound of score as opponent cannot win next move
    if alpha < min {
        alpha = min // there is no need to keep beta above our max possible score.
        if alpha >= beta {
            return alpha, 1 // prune the exploration if the [alpha;beta] window is empty.
        }
    }

    max := int(WIDTH * HEIGHT - 1 - p.MoveCount()) / 2
    if val, err := transTable.Get(p.Key()); err == nil && val != 0 {
        max = int(val) + MIN_SCORE - 1
    }
    if beta > max {
        beta = max
        if(alpha >= beta) {
            return beta, 1
        }
    }

    moves := make(PositionList, WIDTH)
    count := 0
    for x := uint(0); x < WIDTH; x++ {
        if next & columnMask(COLUMN_ORDER[x]) != 0 {
            p2 := p.Play(COLUMN_ORDER[x])
            move := PositionListEntry{p2, p2.Score()}
            idx := count
            for ; idx > 0 && moves[idx-1].s < move.s; idx-- {
                moves[idx] = moves[idx-1]
            }
            moves[idx] = move
            count++
        }
    }

    totalStates := 1
    for x := 0; x < count; x++ {
        p2 := moves[x]
        score, states := negamax(transTable, p2.p, -beta, -alpha)
        score = -score
        totalStates += states

        // no need to have good precision for score better than beta (opponent's score worse than -beta)
        if score >= beta {
            return score, totalStates
        }
        // no need to check for score worse than alpha (opponent's score worse better than -alpha)
        if score > alpha {
            alpha = score
        }
    }

    transTable.Put(p.Key(), uint8(alpha - MIN_SCORE + 1))
    return alpha, totalStates;
}

func solve(transTable *TranspositionTable, p Position) (int, int) {
    // check if win in one move as the Negamax function does not support this case.
    if p.CanWinNext() {
        return int(WIDTH*HEIGHT+1 - p.MoveCount()) / 2, 1
    }

    min := -int(WIDTH * HEIGHT - p.MoveCount()) / 2
    max := int(WIDTH * HEIGHT + 1 - p.MoveCount()) / 2
    totalStates := 0
    for min < max {
        med := min + (max - min) / 2
        if med <= 0 && min / 2 < med {
            med = min / 2
        } else if med >= 0 && max / 2 > med {
            med = max / 2
        }
        r, states := negamax(transTable, p, med, med + 1)
        totalStates += states
        if r <= med {
            max = r
        } else {
            min = r
        }
    }
    return min, totalStates
}

func play(moves string) {
    var p Position
    for i := 0; i < len(moves); i++ {
        col := uint(moves[i] - '1')
        if col >= WIDTH {
            fmt.Fprintf(os.Stderr, "Move %d: Position %c out of range\n", i, moves[i])
            return
        } else if !p.CanPlay(col) {
            fmt.Fprintf(os.Stderr, "Move %d: Cannot play in column %c:\n%s\n", i, moves[i], p)
            return
        } else if p.IsWinningMove(col) {
            fmt.Fprintf(os.Stderr, "Move %d: Cannot play in column %c, as it would end the game:\n%s\n", i, moves[i], p)
            return
        }
        p = p.Play(col)
    }

    transTable := NewTranspositionTable(16777259)

    start := time.Now()
    score, states := solve(transTable, p)
    elapsed := time.Since(start)
    fmt.Printf("%s %d %d %d\n", moves, score, states, elapsed.Nanoseconds() / 1000)
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        line := strings.Trim(scanner.Text(), "\n")
        play(line)
    }

    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "error:", err)
        os.Exit(1)
    }
}
