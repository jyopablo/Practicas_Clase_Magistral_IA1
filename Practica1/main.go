package main

import (
	"container/heap"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// --- Estado del puzzle ---
type State struct {
	board   [9]int
	zeroPos int
	g, h, f int
	parent  *State
	move    string
}

type PriorityQueue []*State

func (pq PriorityQueue) Len() int            { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool  { return pq[i].f < pq[j].f }
func (pq PriorityQueue) Swap(i, j int)       { pq[i], pq[j] = pq[j], pq[i] }
func (pq *PriorityQueue) Push(x interface{}) { *pq = append(*pq, x.(*State)) }
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

// --- Puzzle Solver ---
type PuzzleSolver struct {
	goalState [9]int
}

func NewPuzzleSolver() *PuzzleSolver {
	return &PuzzleSolver{goalState: [9]int{1, 2, 3, 4, 5, 6, 7, 8, 0}}
}

func (ps *PuzzleSolver) manhattanDistance(board [9]int) int {
	dist := 0
	for i, val := range board {
		if val != 0 {
			rowGoal, colGoal := (val-1)/3, (val-1)%3
			row, col := i/3, i%3
			dist += int(math.Abs(float64(row-rowGoal)) + math.Abs(float64(col-colGoal)))
		}
	}
	return dist
}

func (ps *PuzzleSolver) getSuccessors(state *State) []*State {
	moves := []struct {
		dr, dc int
		name   string
	}{
		{-1, 0, "UP"}, {1, 0, "DOWN"}, {0, -1, "LEFT"}, {0, 1, "RIGHT"},
	}
	row, col := state.zeroPos/3, state.zeroPos%3
	var successors []*State

	for _, m := range moves {
		newRow, newCol := row+m.dr, col+m.dc
		if newRow >= 0 && newRow < 3 && newCol >= 0 && newCol < 3 {
			newPos := newRow*3 + newCol
			newBoard := state.board
			newBoard[state.zeroPos], newBoard[newPos] = newBoard[newPos], newBoard[state.zeroPos]
			newState := &State{board: newBoard, zeroPos: newPos, g: state.g + 1, parent: state, move: m.name}
			newState.h = ps.manhattanDistance(newBoard)
			successors = append(successors, newState)
		}
	}
	return successors
}

func (ps *PuzzleSolver) isGoal(board [9]int) bool {
	return board == ps.goalState
}

// A* solver genérico
func (ps *PuzzleSolver) Solve(initial [9]int) []*State {
	start := &State{board: initial, zeroPos: indexOf(initial, 0)}
	start.h = ps.manhattanDistance(initial)
	start.f = start.g + start.h

	openList := &PriorityQueue{start}
	heap.Init(openList)
	visited := make(map[[9]int]bool)

	for openList.Len() > 0 {
		current := heap.Pop(openList).(*State)
		if ps.isGoal(current.board) {
			return reconstructPath(current)
		}
		if visited[current.board] {
			continue
		}
		visited[current.board] = true

		for _, succ := range ps.getSuccessors(current) {
			if !visited[succ.board] {
				succ.f = succ.g + succ.h
				heap.Push(openList, succ)
			}
		}
	}
	return nil
}

func reconstructPath(state *State) []*State {
	var path []*State
	for state != nil {
		path = append([]*State{state}, path...)
		state = state.parent
	}
	return path
}

func indexOf(arr [9]int, val int) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}
	return -1
}

// --- GUI ---
type PuzzleGUI struct {
	window       fyne.Window
	solver       *PuzzleSolver
	buttons      [9]*widget.Button
	currentState [9]int
	solution     []*State
	stepIndex    int
	infoLabel    *widget.Label
	stepLabel    *widget.Label
}

func NewPuzzleGUI() *PuzzleGUI {
	gui := &PuzzleGUI{solver: NewPuzzleSolver(), currentState: [9]int{1, 2, 3, 4, 5, 6, 7, 8, 0}}
	myApp := app.New()
	gui.window = myApp.NewWindow("8-Puzzle Solver")
	gui.window.Resize(fyne.NewSize(400, 500))
	gui.setupUI()
	return gui
}

func (gui *PuzzleGUI) setupUI() {
	grid := container.NewGridWithColumns(3)
	for i := 0; i < 9; i++ {
		index := i
		btn := widget.NewButton("", func() { gui.moveTile(index) })
		gui.buttons[i] = btn
		grid.Add(btn)
	}
	gui.updateButtons()

	// Botones de control
	initBtn := widget.NewButton("Iniciar", gui.reset)
	shuffleBtn := widget.NewButton("Desordenar", gui.shufflePuzzle)
	solveBtn := widget.NewButton("Resolver", gui.solveGUI)
	stepBtn := widget.NewButton("Paso a Paso", gui.solveStepByStep)
	nextBtn := widget.NewButton("Siguiente Paso", gui.nextStep)

	gui.infoLabel = widget.NewLabel("Haz clic en 'Iniciar'")
	gui.stepLabel = widget.NewLabel("")

	content := container.NewVBox(
		grid,
		container.NewVBox(initBtn, shuffleBtn, container.NewHBox(solveBtn, stepBtn, nextBtn)),
		gui.infoLabel, gui.stepLabel,
	)
	gui.window.SetContent(content)
}

func (gui *PuzzleGUI) updateButtons() {
	for i, val := range gui.currentState {
		if val == 0 {
			gui.buttons[i].SetText("")
		} else {
			gui.buttons[i].SetText(strconv.Itoa(val))
		}
	}
}

func (gui *PuzzleGUI) moveTile(pos int) {
	zeroPos := indexOf(gui.currentState, 0)
	row1, col1 := pos/3, pos%3
	row2, col2 := zeroPos/3, zeroPos%3
	if (math.Abs(float64(row1-row2)) == 1 && col1 == col2) || (math.Abs(float64(col1-col2)) == 1 && row1 == row2) {
		gui.currentState[pos], gui.currentState[zeroPos] = gui.currentState[zeroPos], gui.currentState[pos]
		gui.updateButtons()
		if gui.solver.isGoal(gui.currentState) {
			gui.infoLabel.SetText("¡Felicidades! Puzzle resuelto manualmente")
		}
	}
}

func (gui *PuzzleGUI) shufflePuzzle() {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 50; i++ { // n pasos aleatorios
		zero := indexOf(gui.currentState, 0)
		moves := []int{}
		row, col := zero/3, zero%3
		if row > 0 {
			moves = append(moves, zero-3)
		}
		if row < 2 {
			moves = append(moves, zero+3)
		}
		if col > 0 {
			moves = append(moves, zero-1)
		}
		if col < 2 {
			moves = append(moves, zero+1)
		}
		if len(moves) > 0 {
			move := moves[rand.Intn(len(moves))]
			gui.currentState[zero], gui.currentState[move] = gui.currentState[move], gui.currentState[zero]
		}
	}
	gui.updateButtons()
	gui.infoLabel.SetText("Puzzle desordenado")
	gui.stepLabel.SetText("")
	gui.solution = nil
	gui.stepIndex = 0
}

func (gui *PuzzleGUI) solveGUI() {
	gui.solution = gui.solver.Solve(gui.currentState)
	if gui.solution == nil {
		gui.infoLabel.SetText("No se encontró solución")
		return
	}
	go func() {
		for i := 1; i < len(gui.solution); i++ {
			time.Sleep(400 * time.Millisecond)
			gui.currentState = gui.solution[i].board
			gui.updateButtons()
			gui.stepLabel.SetText(fmt.Sprintf("Paso %d/%d: %s", i, len(gui.solution)-1, gui.solution[i].move))
		}
		gui.stepLabel.SetText("¡Puzzle resuelto!")
	}()
}

func (gui *PuzzleGUI) solveStepByStep() {
	gui.solution = gui.solver.Solve(gui.currentState)
	gui.stepIndex = 0
	if gui.solution == nil {
		gui.infoLabel.SetText("No se encontró solución")
		return
	}
	gui.infoLabel.SetText(fmt.Sprintf("Solución lista. Pasos: %d", len(gui.solution)-1))
	gui.stepLabel.SetText("Presiona 'Siguiente Paso'")
}

func (gui *PuzzleGUI) nextStep() {
	if gui.solution == nil || gui.stepIndex >= len(gui.solution)-1 {
		return
	}
	gui.stepIndex++
	gui.currentState = gui.solution[gui.stepIndex].board
	gui.updateButtons()
	if gui.stepIndex < len(gui.solution)-1 {
		gui.stepLabel.SetText(fmt.Sprintf("Paso %d/%d: %s", gui.stepIndex, len(gui.solution)-1, gui.solution[gui.stepIndex].move))
	} else {
		gui.stepLabel.SetText("¡Puzzle resuelto completamente!")
	}
}

func (gui *PuzzleGUI) reset() {
	gui.currentState = [9]int{1, 2, 3, 4, 5, 6, 7, 8, 0}
	gui.updateButtons()
	gui.infoLabel.SetText("Puzzle reiniciado")
	gui.stepLabel.SetText("")
	gui.solution = nil
	gui.stepIndex = 0
}

func (gui *PuzzleGUI) Run() { gui.window.ShowAndRun() }

func main() { NewPuzzleGUI().Run() }
