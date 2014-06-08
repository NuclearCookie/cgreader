package cgreader

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

// conf

var buffer int = 2048
var delay time.Duration
var timeout float64 = 1.0

func SetBuffer(size int) {
	buffer = size
}

func SetFrameRate(fps int) {
	if fps == 0 {
		SetDelay(0)
	} else {
		SetDelay(1000 / fps)
	}
}

func SetDelay(ms int) {
	t := fmt.Sprintf("%dms", ms)
	d, err := time.ParseDuration(t)
	if err == nil {
		delay = d
	}
}

func SetTimeout(f float64) {
	timeout = f
}

// output

type PrintfCallback func(format string, a ...interface{})

var Printf PrintfCallback = func(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func SetPrintfCallback(callback PrintfCallback) {
	Printf = callback
}

// debug

func Trace(msg string) {
	Printf("%s\n", msg)
}

func Tracef(format string, a ...interface{}) {
	Printf("%s\n", fmt.Sprintf(format, a...))
}

// src

func GetManualInput(in string) <-chan string {
	ch := make(chan string, buffer)
	file, err := ioutil.ReadFile(in)
	if err == nil {
		lines := strings.Split(string(file), "\n")
		go func() {
			for _, line := range lines {
				if line != "" {
					ch <- line
				}
			}
			close(ch)
		}()
	} else {
		close(ch)
	}
	return ch
}

func TestOutput(test string, output []string) bool {
	file, err := ioutil.ReadFile(test)
	if err == nil {
		test := strings.Split(string(file), "\n")

		for i, line := range output {
			if line != test[i] {
				return false
			}
		}

		return true
	}
	return false
}

type ProgramMain func(<-chan string, chan string)

func ReportResult(result bool, s float64) {
	if result {
		Printf("Your program finished in %fs and is correct! :)\n", s)
	} else {
		Printf("Your program finished in %fs and is incorrect. :(\n", s)
	}
}

func CheckProgramConditions(t time.Time) (s float64) {
	duration := time.Since(t)
	if s = duration.Seconds(); s > timeout {
		Printf("Your program timed out after %fs! :(\n", timeout)
	}
	return
}

type Function func()
type Execute func(chan string)
type Report func([]string, float64)

func RunFunction(function Function) (result bool) {
	ch := make(chan struct{})
	start := time.Now()
	go func() {
		function()
		close(ch)
	}()

	for {
		select {
		case <-ch:
			result = true
			return
		default:
			if CheckProgramConditions(start) > timeout {
				result = false
				return
			}
		}
	}
}

func RunProgram(execute Execute, report Report) bool {
	ch := make(chan float64)
	och := make(chan string, buffer)
	exit := make(chan struct{})
	error := make(chan struct{})

	output := make([]string, 0)

	start := time.Now()
	go func() {
		execute(och)
		ch <- time.Since(start).Seconds()
		close(ch)
	}()

	go func() {
		for {
			select {
			case <-exit:
				return
			default:
				if CheckProgramConditions(start) > timeout {
					close(error)
				}
				time.Sleep(500 * time.Second)
			}

		}
	}()

	for active := true; active; {
		select {
		case <-error:
			active = false
		case line, ok := <-och:
			if ok {
				output = append(output, line)
			} else {
				active = false
			}
		}
	}

	close(exit)

	report(output, <-ch)
	return true
}

func RunManualProgram(in string, main ProgramMain) {
	output := make(chan string, buffer)
	exit := make(chan struct{})

	go func() {
		main(GetManualInput(in), output)
		close(output)
		close(exit)
	}()

	for {
		select {
		case <-exit:
			return
		case line := <-output:
			Printf("%s\n", line)
		}
	}
}

func RunAndValidateManualProgram(in, test string, echo bool, main ProgramMain) {
	input := GetManualInput(in)
	RunProgram(func(output chan string) {
		main(input, output)
		close(output)
	}, func(output []string, time float64) {
		if echo {
			for _, line := range output {
				Printf("%s\n", line)
			}
		}

		result := TestOutput(test, output)
		ReportResult(result, time)
	})
}

type TargetProgram interface {
	ParseInitialData(<-chan string)
	GetInput() chan string
	Update(<-chan string, chan string)
	SetOutput([]string) string
	LoseConditionCheck() bool
	WinConditionCheck() bool
}

func RunTargetProgram(in string, trace bool, program TargetProgram) {
	ch := GetManualInput(in)

	if RunFunction(func() { program.ParseInitialData(ch) }) {
		for active := true; active; {
			input := program.GetInput()
			if RunProgram(func(output chan string) {
				program.Update(input, output)
				close(output)
			}, func(output []string, duration float64) {
				result := program.SetOutput(output)

				if trace {
					for _, line := range output {
						Printf("%s\n", line)
					}
					Printf("\n%s\n\n", result)
				}

				duration += duration

				if program.WinConditionCheck() {
					ReportResult(true, duration)
					active = false
				} else if program.LoseConditionCheck() {
					ReportResult(false, duration)
					active = false
				}

				time.Sleep(delay)
			}) == false {
				return
			}
		}
	}
}

type MapObject interface {
	GetMapCoordinates() string // returns string in format x;y
	GetMapIcon() string        // return 1 character string
}

func DrawMap(width, height int, background string, objects ...MapObject) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			position := fmt.Sprintf("%d;%d", x, y)
			c := background
			for _, object := range objects {
				pos := object.GetMapCoordinates()
				if pos == position {
					c = object.GetMapIcon()
					break
				}
			}
			Printf("%s ", c)
		}
		Printf("\n")
	}
	Printf("\n")
}
