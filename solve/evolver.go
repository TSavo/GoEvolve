package solve

import (
	"github.com/tsavo/golightly/vm"
	"runtime"
	"sort"
	"time"
)

type IslandEvolver struct {
	ChampionSize     int
	SolverReportChan chan *SolverReport
	InfluxBreeder
}

type Champion struct {
	Reward   int64
	Programs []string
}

type Champions []Champion

func (s Champions) Len() int      { return len(s) }
func (s Champions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s Champions) Less(i, j int) bool {
	return s[i].Reward > s[j].Reward
}

func NewIslandEvolver(champSize int) *IslandEvolver {
	i := IslandEvolver{champSize, make(chan *SolverReport, 100), make(InfluxBreeder, 100)}
	go i.CollectBest()
	return &i
}

func (self *IslandEvolver) AddPopulation(id int, heap *vm.Memory, registerSize int, is *vm.InstructionSet, term vm.TerminationCondition, breeder Breeder, eval Evaluator, selector Selector) {
	breeders := Breeders(breeder, self.InfluxBreeder)
	solver := NewSolver(id, heap, registerSize, is, term, breeders, eval, selector)
	solver.SolverReportChan = self.SolverReportChan
}

func (self *IslandEvolver) CollectBest() {
	for {
		best := make(Champions, self.ChampionSize)
		for x := 0; x < self.ChampionSize; x++ {
			runtime.Gosched()
			solverReport := <- self.SolverReportChan
			sort.Sort(solverReport)
			champ := Champion{solverReport.SolutionList[0].Reward, make([]string, len(solverReport.SolutionList))}
			for y := 0; y < len(solverReport.SolutionList); y++ {
				champ.Programs[y] = solverReport.SolutionList[y].Program
			}
			best[x] = champ
		}
		go func(champs Champions) {
			sort.Sort(champs)
			time.Sleep(time.Second)
			self.InfluxBreeder <- champs[0].Programs
		}(best)
	}
}
