package goevolve

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/TSavo/GoVirtual"
	"io"
	"log"
	"os"
	"time"
)

type Population struct {
	Id, RegisterLength   int
	InstructionSet       *govirtual.InstructionSet
	Breeder              *Breeder
	Evaluator            *Evaluator
	Selector             *Selector
	TerminationCondition *govirtual.TerminationCondition
	ControlChan          chan bool
	PopulationReportChan chan *PopulationReport
	Heap                 *govirtual.Memory
}

var SolutionCache map[string]*Solution

func init() {
	SolutionCache = make(map[string]*Solution)
	go func() {
		for {
			time.Sleep(60 * time.Second)
			WriteSolutionCache(EncodeSolutionCache())
		}
	}()

	defer recover()
	cache := ReadSolutionCache()
	if(cache != nil) {
		SolutionCache = *DecodeSolutionCache(cache)
	}
}

func EncodeSolutionCache() (b *bytes.Buffer) {
	b = new(bytes.Buffer)
	e := gob.NewEncoder(b)

	// Encoding the map
	err := e.Encode(&SolutionCache)
	if err != nil {
		panic(err)
	}
	return
}

func DecodeSolutionCache(b *bytes.Buffer) *map[string]*Solution {
	s := make(map[string]*Solution)
	d := gob.NewDecoder(b)

	// Decoding the serialized data
	err := d.Decode(&s)
	if err != nil {
		return nil
	}
	return &s
}

func WriteSolutionCache(b *bytes.Buffer) {
	f, er := os.OpenFile("SolutionCache.gob", os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0777  )
	f.Write(b.Bytes()) // Error handling elided for brevity.
	f.Close()
}

func ReadSolutionCache() *bytes.Buffer {
	buf := bytes.NewBuffer(nil)
	f, _ := os.Open("SolutionCache.gob") // Error handling elided for brevity.
	written, err := io.Copy(buf, f)      // Error handling elided for brevity.
	f.Close()
	if err == nil && written > 0 {
		return buf
	}
	return nil
}

type Solution struct {
	Reward  int
	Program string
}

type SolutionList []*Solution

func (sol *SolutionList) GetPrograms() []string {
	x := make([]string, len(*sol))
	for i, solution := range *sol {
		x[i] = solution.Program
	}
	return x
}

type PopulationReport struct {
	Id int
	SolutionList
}

func (s SolutionList) Len() int           { return len(s) }
func (s SolutionList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SolutionList) Less(i, j int) bool { return s[i].Reward > s[j].Reward }

func NewPopulation(id int, sharedMemory *govirtual.Memory, rl int, is *govirtual.InstructionSet, term govirtual.TerminationCondition, gen Breeder, eval Evaluator, selector Selector) *Population {
	return &Population{id, rl, is, &gen, &eval, &selector, &term, make(chan bool, 1), make(chan *PopulationReport, 1), sharedMemory}
}

func (s *Population) Run() {
	programs := (*s.Breeder).Breed((*s.Breeder).Breed(nil))
	processors := make([]*govirtual.Processor, 0)
	for {
		solutions := make(SolutionList, len(programs))
		for len(processors) < len(solutions) {
			c := govirtual.NewProcessor(s.Id, s.RegisterLength, s.InstructionSet, s.Heap, s.TerminationCondition)
			processors = append(processors, c)
		}
		if len(processors) > len(solutions) {
			processors = processors[:len(solutions)]
		}

		for x, pro := range processors {
			select {
			case <-s.ControlChan:
				return
			default:
			}
			log.Printf("#%d: %d\n", s.Id, x)
			sha := fmt.Sprintf("%x", sha256.Sum256([]byte(programs[x])))
			sol, notNeeded := SolutionCache[sha]
			if notNeeded {
				solutions[x] = sol
			} else {
				pro.Reset()
				pro.CompileAndLoad(programs[x])
				pro.Run()
				solutions[x] = &Solution{(*s.Evaluator).Evaluate(pro), programs[x]}
				potential, present := SolutionCache[sha]
				if !present || potential.Reward > solutions[x].Reward {
					SolutionCache[sha] = solutions[x]
				}
			}
		}
		select {
		case s.PopulationReportChan <- &PopulationReport{s.Id, solutions}:
		default:
		}
		programs = (*s.Breeder).Breed((*s.Selector).Select(&solutions).GetPrograms())
	}
}
