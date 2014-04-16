package goevolve

import (
	"fmt"
	"github.com/TSavo/GoVirtual"
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
	FloatHeap			 *govirtual.FloatMemory
}

type Solution struct {
	Reward  int64
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

func NewPopulation(id int, sharedMemory *govirtual.Memory, floatMemory *govirtual.FloatMemory, rl int, is *govirtual.InstructionSet, term govirtual.TerminationCondition, gen Breeder, eval Evaluator, selector Selector) *Population {
	return &Population{id, rl, is, &gen, &eval, &selector, &term, make(chan bool, 1), make(chan *PopulationReport, 1), sharedMemory, floatMemory}
}

func (s *Population) Run() {
	programs := (*s.Breeder).Breed(nil)
	processors := make([]*govirtual.Processor, 0)
	for {
		solutions := make(SolutionList, len(programs))
		for len(processors) < len(solutions) {
			c := govirtual.NewProcessor(s.RegisterLength, s.InstructionSet, s.Heap, s.FloatHeap, s.TerminationCondition)
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
			fmt.Printf("#%d: %d\n", s.Id, x)
			pro.Reset()
			pro.CompileAndLoad(programs[x])
			pro.Run()
			solutions[x] = &Solution{(*s.Evaluator).Evaluate(pro), pro.Program.Decompile()}
		}
		select {
		case s.PopulationReportChan <- &PopulationReport{s.Id, solutions}:
		default:
		}
		programs = (*s.Breeder).Breed((*s.Selector).Select(&solutions).GetPrograms())
	}
}
