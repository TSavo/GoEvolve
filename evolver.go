package goevolve

import (
	"bufio"
	"github.com/TSavo/GoVirtual"
	"os"
	"runtime"
	"sort"
	"time"
	"os/exec"
	"log"
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

func NewIslandEvolver() *IslandEvolver {
	i := IslandEvolver{0, make(chan *PopulationReport, 100), make(InfluxBreeder, 100), 0}
	go i.Interbreed()
	return &i
}

func (self *IslandEvolver) AddPopulation(heap *govirtual.Memory, floatHeap *govirtual.FloatMemory, registerSize int, is *govirtual.InstructionSet, term govirtual.TerminationCondition, breeder Breeder, eval Evaluator, selector Selector) {
	breeders := Breeders(breeder, self.InfluxBreeder)
	population := NewPopulation(self.lastId, heap, floatHeap, registerSize, is, term, breeders, eval, selector)
	population.PopulationReportChan = self.PopulationReportChan
	go population.Run()
	self.lastId++
	self.ChampionSize++
}

func (self *IslandEvolver) Interbreed() {
	for {
		if self.ChampionSize < 2 {
			time.Sleep(1 * time.Second)
			continue
		}
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
			writeFile("bestProgram.vm", champs[0].Programs[0])
		}(best)
	}
}

func writeFile(name, data string) {
	f, _ := os.Create(name)
	w := bufio.NewWriter(f)
	w.WriteString(data)
	w.Flush()
	f.Close()
	cmd := exec.Command("git", "add .")
	out, _ := cmd.Output()
	log.Printf("Git: %v", string(out))
	cmd = exec.Command("git", "commit -m \"Automated pushing best program so far\"")
	out, _ = cmd.Output()
	log.Printf("Git: %v", string(out))
	cmd = exec.Command("git", "push")
	out, _ = cmd.Output()
	log.Printf("Git: %v", string(out))
}
