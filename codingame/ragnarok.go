package codingame

import (
	"fmt"
	"strings"
)

type Ragnarok struct {
	thor, target, dimensions Vector
	energy                   int
	trail                    []Vector
	UserInitialize           UserInitializeFunction
	UserUpdate               UserUpdateFunction
	trace                    bool
}

func GetDirection(x, y int) <-chan int {
	ch := make(chan int)
	go func() {
		difference := x - y
		switch {
		case difference < 0:
			ch <- -1
		case difference > 0:
			ch <- 1
		default:
			ch <- 0
		}
		close(ch)
	}()
	return ch
}

func GetDirectionLetter(a, b string, v int) string {
	switch v {
	default:
		return ""
	case -1:
		return a
	case 1:
		return b
	}
	return ""
}

func (ragnarok *Ragnarok) ParseInitialData(ch <-chan string) {
	fmt.Sscanf(
		<-ch,
		"%d %d %d %d %d %d %d",
		&ragnarok.dimensions.x,
		&ragnarok.dimensions.y,
		&ragnarok.thor.x,
		&ragnarok.thor.y,
		&ragnarok.energy,
		&ragnarok.target.x,
		&ragnarok.target.y)

	output := make(chan string)
	go func() {
		output <- fmt.Sprintf(
			"%d %d %d %d\n",
			ragnarok.target.x,
			ragnarok.target.y,
			ragnarok.thor.x,
			ragnarok.thor.y)
	}()
	ragnarok.UserInitialize(output)

	ragnarok.thor.icon, ragnarok.target.icon = "H", "T"
}

func (ragnarok *Ragnarok) GetInput() (ch chan string) {
	ch = make(chan string)
	go func() {
		ch <- fmt.Sprintf("%d\n", ragnarok.energy)
	}()
	return
}

func (ragnarok *Ragnarok) Update(input <-chan string, output chan string) {
	ragnarok.UserUpdate(input, output)
}

func (ragnarok *Ragnarok) SetOutput(output []string) string {
	ragnarok.trail = append(ragnarok.trail, Vector{ragnarok.thor.x, ragnarok.thor.y, "+"})

	if strings.Contains(output[0], "N") {
		ragnarok.thor.y -= 1
	} else if strings.Contains(output[0], "S") {
		ragnarok.thor.y += 1
	}

	if strings.Contains(output[0], "E") {
		ragnarok.thor.x += 1
	} else if strings.Contains(output[0], "W") {
		ragnarok.thor.x -= 1
	}

	ragnarok.energy -= 1

	if ragnarok.trace {
		trail := append(ragnarok.trail, ragnarok.thor, ragnarok.target)

		map_info := make([]MapObject, len(trail))
		for i, v := range trail {
			map_info[i] = MapObject(v)
		}

		DrawMap(
			ragnarok.dimensions.x,
			ragnarok.dimensions.y,
			".",
			map_info...)

		return fmt.Sprintf(
			"Target = (%d,%d)\nThor = (%d,%d)\nEnergy = %d",
			ragnarok.target.x,
			ragnarok.target.y,
			ragnarok.thor.x,
			ragnarok.thor.y,
			ragnarok.energy)
	}

	return ""
}

func (ragnarok *Ragnarok) LoseConditionCheck() bool {
	if ragnarok.energy <= 0 {
		return true
	}

	x, y := ragnarok.thor.x, ragnarok.thor.y
	dx, dy := ragnarok.dimensions.x, ragnarok.dimensions.y

	if x < 0 || x >= dx || y < 0 || y >= dy {
		return true
	}

	return false
}

func (ragnarok *Ragnarok) WinConditionCheck() bool {
	return ragnarok.target.x == ragnarok.thor.x &&
		ragnarok.target.y == ragnarok.thor.y
}

func RunRagnarokProgram(input string, trace bool, initialize UserInitializeFunction, update UserUpdateFunction) bool {
	ragnarok := Ragnarok{}
	ragnarok.UserInitialize = initialize
	ragnarok.UserUpdate = update
	ragnarok.trace = trace

	return RunTargetProgram(input, trace, &ragnarok)
}

func RunRagnarokPrograms(input []string, trace bool, initialize UserInitializeFunction, update UserUpdateFunction) {
	var counter int
	for i := range input {
		if RunRagnarokProgram(input[i], trace, initialize, update) {
			counter++
		}
		Println("")
	}
	ReportTotalResult(counter, len(input))
}
