package solve

import (
	"fmt"

	"github.com/tsavo/golightly/vm"
)

type Solver struct {
	Id, RegisterLength   int
	InstructionSet       *vm.InstructionSet
	Breeder              *Breeder
	Evaluator            *Evaluator
	Selector             *Selector
	TerminationCondition *vm.TerminationCondition
	ControlChan          chan bool
	SolverReportChan     chan *SolverReport
	Heap                 *vm.Memory
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

type SolverReport struct {
	Id int
	SolutionList
}

func (s SolutionList) Len() int           { return len(s) }
func (s SolutionList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SolutionList) Less(i, j int) bool { return s[i].Reward > s[j].Reward }

func NewSolver(id int, sharedMemory *vm.Memory, rl int, is *vm.InstructionSet, term vm.TerminationCondition, gen Breeder, eval Evaluator, selector Selector) *Solver {
	return &Solver{id, rl, is, &gen, &eval, &selector, &term, make(chan bool, 1), make(chan *SolverReport, 1), sharedMemory}
}

func (s *Solver) SolveOneAtATime() {
	programs := (*s.Breeder).Breed(nil)
	processors := make([]*vm.ProcessorCore, 0)
	for {
		solutions := make(SolutionList, len(programs))
		for len(processors) < len(solutions) {
			c := vm.NewProcessorCore(s.RegisterLength, s.InstructionSet, s.Heap, s.TerminationCondition)
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
		case s.SolverReportChan <- &SolverReport{s.Id, solutions}:
		default:
		}
		programs = (*s.Breeder).Breed((*s.Selector).Select(&solutions).GetPrograms())
	}
}
