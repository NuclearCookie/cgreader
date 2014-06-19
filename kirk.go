package cgreader

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

const (
	KIRK_M0 = 0x000f
	KIRK_M1 = 0x00f0
	KIRK_M2 = 0x0f00
	KIRK_MX = 0xf000

	KIRK_S0 = 0
	KIRK_S1 = 4
	KIRK_S2 = 8
	KIRK_SX = 12

	KIRK_N = 8
)

type Kirk struct {
	UserInitialize UserInitializeFunction
	UserUpdate     UserUpdateFunction
	trace          bool
	maxHeight      uint8
	mountains      []uint32
	player         Vector
	direction      int8
	canFire        bool
}

func (kirk *Kirk) ParseInitialData(ch <-chan string) {
	fmt.Sscanf(<-ch, "%u", &kirk.maxHeight)
	kirk.player = Vector{0, int(kirk.maxHeight), "S"}

	for i := 0; i < KIRK_N; i++ {
		heights := strings.Split(<-ch, " ")
		for u, h := range heights {
			var height uint32
			fmt.Sscanf(h, "%u", &height)
			kirk.mountains[i] += height << uint32(u) * KIRK_S1
		}
		if kirk.mountains[i]&KIRK_M0 != 0 {
			kirk.mountains[i] += uint32(len(heights)) << KIRK_SX
		}
	}
}

func GetKirkMountainHeight(m uint32) (height uint32) {
	if id := (m & KIRK_MX) >> KIRK_SX; id > 0 {
		for id--; id > 0; id-- {
			s := KIRK_S1 * id
			height += (m & (KIRK_M0 << s)) >> (KIRK_S1 * s)
		}
	}
	return
}

func (kirk *Kirk) GetInput() (ch chan string) {
	ch = make(chan string)
	go func() {
		ch <- fmt.Sprintf("%d %d", kirk.player.x, kirk.player.y)
		for _, mountain := range kirk.mountains {
			ch <- fmt.Sprintf("%u", GetKirkMountainHeight(mountain))
		}
	}()
	return
}

func (kirk *Kirk) Update(input <-chan string, output chan string) {
	kirk.UserUpdate(input, output)
}

func (kirk *Kirk) SetOutput(output []string) string {
	playerFired, damage := kirk.canFire && output[0] == "FIRE", uint32(0)
	if playerFired {
		m := kirk.mountains[kirk.player.x]
		x := m >> KIRK_SX
		if x > 0 {
			id := x - 1
			s := KIRK_S1 * id
			height := int32((m >> s) & KIRK_M0)
			damage = uint32(rand.Int31n(height))
			height -= int32(damage)
			if height == 0 {
				x--
			}
			m &= math.MaxUint32 - KIRK_MX - (KIRK_M0 << id * 4)
			kirk.mountains[kirk.player.x] = m | (uint32(height) << s) | (x << KIRK_SX)
		}
	}

	kirk.player.x += int(kirk.direction)
	if kirk.player.x < 0 || kirk.player.x > KIRK_N-1 {
		kirk.player.y, kirk.direction = kirk.player.y-1, -kirk.direction
		kirk.player.x += int(kirk.direction * 2)
	}

	if kirk.trace {
		icons := make([]MapObject, KIRK_N+1)
		icons[KIRK_N] = MapObject(kirk.player)

		for i, mountain := range kirk.mountains {
			icons[i] = MapObject(Vector{i, int(GetKirkMountainHeight(mountain)), "^"})
		}

		DrawMap(KIRK_N, int(kirk.maxHeight), ".", icons...)

		shipInfo := fmt.Sprintf("Ship = (%d,%d)\n", kirk.player.x, kirk.player.y)
		if playerFired {
			return shipInfo + fmt.Sprintf("Ship fired and did %u damage.", damage)
		} else {
			return shipInfo + "Ship hold fire."
		}
	}

	return ""
}

func (kirk *Kirk) LoseConditionCheck() bool {
	return kirk.player.y > 0 &&
		kirk.player.y <= int(GetKirkMountainHeight(kirk.mountains[kirk.player.x]))
}

func (kirk *Kirk) WinConditionCheck() bool {
	for _, mountain := range kirk.mountains {
		if mountain>>KIRK_SX > 0 {
			return false
		}
	}
	return true
}

func RunKirkProgram(input string, trace bool, initialize UserInitializeFunction, update UserUpdateFunction) {
	kirk := Kirk{}

	SetTimeout(0.1)

	kirk.UserInitialize, kirk.UserUpdate, kirk.trace = initialize, update, trace
	kirk.mountains, kirk.direction, kirk.canFire = make([]uint32, KIRK_N), 1, true

	RunTargetProgram(input, trace, &kirk)
}

func RunKirkPrograms(input []string, trace bool, initialize UserInitializeFunction, update UserUpdateFunction) {
	for i := range input {
		RunKirkProgram(input[i], trace, initialize, update)
		Printf("\n")
	}
}
