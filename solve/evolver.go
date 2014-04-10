package solve

import (
	"github.com/TSavo/GoVirtual/vm"
	"runtime"
	"sort"
	"time"
)

type IslandEvolver struct {
	ChampionSize         int
	PopulationReportChan chan *PopulationReport
	InfluxBreeder
	lastId int
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
	i := IslandEvolver{champSize, make(chan *PopulationReport, 100), make(InfluxBreeder, 100), 0}
	go i.Interbreed()
	return &i
}

func (self *IslandEvolver) AddPopulation(heap *vm.Memory, registerSize int, is *vm.InstructionSet, term vm.TerminationCondition, breeder Breeder, eval Evaluator, selector Selector) {
	breeders := Breeders(breeder, self.InfluxBreeder)
	population := NewPopulation(self.lastId, heap, registerSize, is, term, breeders, eval, selector)
	population.PopulationReportChan = self.PopulationReportChan
	go population.Run()
	self.lastId++
}

func (self *IslandEvolver) Interbreed() {
	for {
		best := make(Champions, self.ChampionSize)
		for x := 0; x < self.ChampionSize; x++ {
			runtime.Gosched()
			populationReport := <-self.PopulationReportChan
			sort.Sort(populationReport)
			champ := Champion{populationReport.SolutionList[0].Reward, make([]string, len(populationReport.SolutionList))}
			for y := 0; y < len(populationReport.SolutionList); y++ {
				champ.Programs[y] = populationReport.SolutionList[y].Program
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
