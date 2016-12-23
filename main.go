// +build ignore

package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tsavo/GoEvolve"
	"github.com/tsavo/GoVirtual"
	"log"
	"math/rand"
	"net/http"
	"os/user"
	"strconv"
	"time"
)

const (
	POPULATION_SIZE = 100
	CHAMPION_SIZE   = 3
	BEST_OF_BREED   = 10
	PROGRAM_LENGTH  = 12
	UNIVERSE_SIZE   = 9
	ROUND_LENGTH    = 10
)

func DefineInstructions(flapChan chan bool) (i *govirtual.InstructionSet) {
	i = govirtual.NewInstructionSet()
	i.Operator("noop", func(p *govirtual.Processor, m *govirtual.Memory) {
	})
	i.Movement("jump", func(p *govirtual.Processor, m *govirtual.Memory) {
		p.Jump(p.Registers.Get(1))
	})
	i.Movement("jumpIfZero", func(p *govirtual.Processor, m *govirtual.Memory) {
		if p.Registers.Get((*m).Get(0)) == 0 {
			p.Jump(p.Registers.Get(1))
		} else {
			p.InstructionPointer++
		}
	})
	i.Movement("jumpIfNotZero", func(p *govirtual.Processor, m *govirtual.Memory) {
		if p.Registers.Get((*m).Get(0)) != 0 {
			p.Jump(p.Registers[1])
		} else {
			p.InstructionPointer++
		}
	})
	i.Movement("jumpIfEquals", func(p *govirtual.Processor, m *govirtual.Memory) {
		if p.Registers.Get((*m).Get(0)) == p.Registers.Get((*m).Get(1)) {
			p.Jump(p.Registers[1])
		} else {
			p.InstructionPointer++
		}
	})
	i.Movement("jumpIfNotEquals", func(p *govirtual.Processor, m *govirtual.Memory) {
		if p.Registers.Get((*m).Get(0)) != p.Registers.Get((*m).Get(1)) {
			p.Jump(p.Registers[1])
		} else {
			p.InstructionPointer++
		}
	})
	i.Movement("jumpIfGreaterThan", func(p *govirtual.Processor, m *govirtual.Memory) {
		if p.Registers.Get((*m).Get(0)) > p.Registers.Get((*m).Get(1)) {
			p.Jump(p.Registers[1])
		} else {
			p.InstructionPointer++
		}
	})
	i.Movement("jumpIfLessThan", func(p *govirtual.Processor, m *govirtual.Memory) {
		if p.Registers.Get((*m).Get(0)) < p.Registers.Get((*m).Get(1)) {
			p.Jump(p.Registers[1])
		} else {
			p.InstructionPointer++
		}
	})
	i.Movement("call", func(p *govirtual.Processor, m *govirtual.Memory) {
		p.Call(p.Registers.Get((*m).Get(0)))
	})
	i.Movement("return", func(p *govirtual.Processor, m *govirtual.Memory) {
		p.Return()
	})
	i.Operator("set", func(p *govirtual.Processor, m *govirtual.Memory) {
		p.Registers.Set((*m).Get(0), (*m).Get(1))
	})
	i.Operator("store", func(p *govirtual.Processor, m *govirtual.Memory) {
		p.Heap.Set(p.Registers.Get(1), p.Registers.Get(0))
	})
	i.Operator("load", func(p *govirtual.Processor, m *govirtual.Memory) {
		p.Registers.Set(0, p.Heap.Get(p.Registers.Get(1)))
	})
	i.Operator("swap", func(p *govirtual.Processor, m *govirtual.Memory) {
		x := p.Registers.Get((*m).Get(0))
		p.Registers.Set((*m).Get(0), (*m).Get(1))
		p.Registers.Set((*m).Get(1), x)
	})
	i.Operator("push", func(p *govirtual.Processor, m *govirtual.Memory) {
		p.Stack.Push(p.Registers.Get((*m).Get(0)))
	})
	i.Operator("pop", func(p *govirtual.Processor, m *govirtual.Memory) {
		if x, err := p.Stack.Pop(); !err {
			p.Registers.Set((*m).Get(0), x)
		}
	})
	i.Operator("increment", func(p *govirtual.Processor, m *govirtual.Memory) {
		p.Registers.Increment((*m).Get(0))
	})
	i.Operator("decrement", func(p *govirtual.Processor, m *govirtual.Memory) {
		p.Registers.Decrement((*m).Get(0))
	})
	i.Operator("add", func(p *govirtual.Processor, m *govirtual.Memory) {
		p.Registers.Set((*m).Get(0), p.Registers.Get((*m).Get(0))+p.Registers.Get((*m).Get(1)))
	})
	i.Operator("subtract", func(p *govirtual.Processor, m *govirtual.Memory) {
		p.Registers.Set((*m).Get(0), p.Registers.Get((*m).Get(0))-p.Registers.Get((*m).Get(1)))
	})
	i.Operator("flap", func(p *govirtual.Processor, m *govirtual.Memory) {
		flapChan <- true
		time.Sleep(50 * time.Millisecond)
	})
	i.Operator("sleep", func(p *govirtual.Processor, m *govirtual.Memory) {
		time.Sleep(50 * time.Millisecond)
	})

	return
}

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

var h = hub{
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			delete(h.connections, c)
			close(c.send)
		}
	}
}

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *connection) reader(incoming chan string) {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		incoming <- string(message)
		//h.broadcast <- message
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

type FlappyEvaluator struct {
	reward int64
}

func (eval *FlappyEvaluator) Evaluate(p *govirtual.Processor) int64 {
	x := eval.reward - (p.Cost() / 10000)
	eval.reward = 0
	return x
}

type FlappyGenerator struct {
	InstructionSet *govirtual.InstructionSet
}

type FlappyBreeder struct {
	PopulationSize int
}

func (flap FlappyBreeder) Breed(seeds []string) []string {
	progs := make([]string, flap.PopulationSize)
	for i, _ := range progs {
		progs[i] = GenerateProgram()
	}
	return progs
}

func GenerateProgram() string {
	pr := `
set 4, 0
set 2, 3
set 1, 5
set 3, ` + strconv.Itoa(rand.Int()%2000) + `
load
subtract 3, 0
set 1, 0
jumpIfGreaterThan 3, 1
flap
sleep
jump`
	fmt.Println(pr)
	return pr
}

var populationInfluxChan goevolve.InfluxBreeder = make(goevolve.InfluxBreeder, 100)
var PopulationReportChan chan *goevolve.PopulationReport = make(chan *goevolve.PopulationReport, 100)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}
	h.register <- c
	defer func() { h.unregister <- c }()
	go c.writer()
	heap := make(govirtual.Memory, 8)
	outChan := make(chan bool, 1)
	is := DefineInstructions(outChan)
	deadChannel := govirtual.NewChannelTerminationCondition()
	terminationCondition := govirtual.OrTerminate(deadChannel, *govirtual.NewCostTerminationCondition(50000000))
	breeder := *goevolve.Breeders(goevolve.NewCopyBreeder(15), FlappyBreeder{10}, goevolve.NewRandomBreeder(25, 50, is), goevolve.NewMutationBreeder(25, 0.1, is), goevolve.NewCrossoverBreeder(25))
	flappyEval := new(FlappyEvaluator)
	selector := goevolve.AndSelect(goevolve.TopX(10), goevolve.Tournament(10))
	FlappyIsland.AddPopulation(&heap, 4, is, terminationCondition, breeder, flappyEval, selector)
	go func() {
		for {
			<-outChan
			go func() {
				c.send <- []byte("1")
			}()
		}
	}()
	inFlap := make(chan string, 1000)

	go func() {
		for {
			flap := <-inFlap
			if flap == "DEAD" {
				*deadChannel <- true
				continue
			}
			myX := 0
			y := 0
			center := 0
			fmt.Sscanf(flap, "%d,%d,%d", &myX, &y, &center)
			heap.Set(5, myX)
			heap.Set(6, y)
			heap.Set(7, center)
			rew := 1000 - myX
			if rew < 0 {
				rew *= -1
			}
			flappyEval.reward += int64(1000 - rew)
		}
	}()
	c.reader(inFlap)
}

func loadProgram(projectName string, id int) govirtual.Program {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(usr.HomeDir)
	return nil
}

var FlappyIsland *goevolve.IslandEvolver

func main() {
	FlappyIsland = goevolve.NewIslandEvolver(3)
	go h.run()
	go func() {
		http.HandleFunc("/ws", wsHandler)
		if err := http.ListenAndServe(":3000", nil); err != nil {
			fmt.Println("ListenAndServe:", err)
		}
	}()
	<-make(chan int)
}
