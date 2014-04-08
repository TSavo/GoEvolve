package solve

import (
	"github.com/tsavo/golightly/intutil"
	"github.com/tsavo/golightly/vm"
	"strings"
)

type Breeder interface {
	Breed(seeds []string) []string
}

type MultiBreeder []Breeder

func Breeders(breeders... Breeder) *MultiBreeder {
	out := MultiBreeder(breeders)
	return &out
}

func (multi MultiBreeder) Breed(seeds []string) []string {
	ret := make([]string, 0)
	for _, x := range multi {
		ret = append(ret, (x).Breed(seeds)...)
	}
	return ret
}

type CopyBreeder struct {
	PopulationSize int
}

func NewCopyBreeder(size int) CopyBreeder {
	return CopyBreeder{size}
}

func (cp CopyBreeder) Breed(initialPop []string) []string {
	if len(initialPop) == 0 {
		return nil
	}
	pop := make([]string, cp.PopulationSize)
	for x, y := 0, 0; x < cp.PopulationSize; x++ {
		y = y % len(initialPop)
		pop[x] = initialPop[y]
		y++
	}
	return pop
}

type RandomBreeder struct {
	PopulationSize, ProgramLength int
	*vm.InstructionSet
}

func NewRandomBreeder(popSize int, programLen int, is *vm.InstructionSet) *RandomBreeder {
	return &RandomBreeder{popSize, programLen, is}
}

func (breeder RandomBreeder) Breed([]string) []string {
	progs := make([]string, breeder.PopulationSize)
	for x := 0; x < breeder.PopulationSize; x++ {
		p := ""
		for y := 0; y < breeder.ProgramLength; y++ {
			p += breeder.RandomOperation().String() + "\n"
		}
		progs[x] = p
	}
	return progs
}

type CrossoverBreeder struct {
	PopulationSize int
}

func NewCrossoverBreeder(popSize int) CrossoverBreeder {
	return CrossoverBreeder{popSize}
}

func (breeder CrossoverBreeder) Breed(seeds []string) []string {
	if len(seeds) == 0 {
		return nil
	}
	outProg := make([]string, breeder.PopulationSize)
	for i := 0; i < breeder.PopulationSize; i++ {
		prog1 := strings.Split(seeds[rng.Int()%len(seeds)], "\n")
		prog2 := strings.Split(seeds[rng.Int()%len(seeds)], "\n")

		l1 := len(prog1)
		l2 := len(prog2)
		prog := make([]string, intutil.Max(l1, l2))
		split := rng.Int() % intutil.Min(l1, l2)
		endSplit := (rng.Int()%intutil.Min(l1, l2) - split) + split
		for x := 0; x < len(prog); x++ {
			if x > len(prog1)-1 || (x < endSplit && x >= split && x < len(prog2)) {
				prog[x] = prog2[x]
			} else {
				prog[x] = prog1[x]
			}
		}
		outProg[i] = strings.Join(prog, "\n")
	}
	return outProg
}

type MutationBreeder struct {
	PopulationSize int
	MutationChance float64
	*vm.InstructionSet
}

func NewMutationBreeder(popSize int, mutationChance float64, is *vm.InstructionSet) MutationBreeder {
	return MutationBreeder{popSize, mutationChance, is}
}

func (breeder MutationBreeder) Breed(seeds []string) []string {
	if len(seeds) == 0 {
		return nil
	}
	out := make([]string, breeder.PopulationSize)
	y := 0
	for x := 0; x < breeder.PopulationSize; x++ {
		y = y % len(seeds)
		startProg := seeds[y]
		prog := breeder.CompileProgram(startProg)
		outProg := make(vm.Program, 0)
		for _, op := range *prog {
			if rng.Float64() < breeder.MutationChance {
				if rng.Float64() < 0.1 {
					for r := rng.Int() % 10; r < 10; r++ {
						outProg = append(outProg, breeder.RandomOperation())
					}
				}
				if rng.Float64() < 0.1 && len(outProg) > 0 {
					continue
				}
				decode := op.Decode()
				if rng.Float64() < 0.5 {
					decode.Set(0, rng.Int())
				}
				if rng.Float64() < 0.5 {
					decode.Set(1, rng.Int())
				}
				if rng.Float64() < 0.5 {
					decode.Set(2, rng.Int())
				}
				outProg = append(outProg, breeder.Encode(decode))
			} else {
				outProg = append(outProg, op)
			}
		}
		out[x] = outProg.Decompile()
		y++
	}
	return out
}

type InfluxBreeder chan []string

func (breeder InfluxBreeder) Breed([]string) []string {
	select {
	case x := <-breeder:
		return x
	default:
		return nil
	}
}
