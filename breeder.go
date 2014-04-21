package goevolve

import (
	"github.com/TSavo/GoVirtual"
	"strings"
)

type Breeder interface {
	Breed(seeds []string) []string
}

type MultiBreeder []Breeder

func Breeders(breeders ...Breeder) *MultiBreeder {
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
	*govirtual.InstructionSet
}

func NewRandomBreeder(popSize int, programLen int, is *govirtual.InstructionSet) *RandomBreeder {
	return &RandomBreeder{popSize, programLen, is}
}

func (breeder RandomBreeder) Breed([]string) []string {
	progs := make([]string, breeder.PopulationSize)
	for x := 0; x < breeder.PopulationSize; x++ {
		p := ""

		for y := 0; y < breeder.ProgramLength; y++ {
			p += breeder.Encode(&govirtual.Memory{rng.Int(), rng.SmallInt(), rng.SmallInt(), rng.SmallInt()}).String() + "\n"
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
		prog := make([]string, Max(l1, l2))
		split := rng.Int() % Min(l1, l2)
		endSplit := (rng.Int() % (Min(l1, l2) - split)) + split
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
	*govirtual.InstructionSet
}

func NewMutationBreeder(popSize int, mutationChance float64, is *govirtual.InstructionSet) MutationBreeder {
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
		outProg := govirtual.NewProgram(0)
		for _, op := range prog.Operations {
			if rng.Float64() < breeder.MutationChance {
				if rng.Float64() < breeder.MutationChance {
					for r := rng.Int() % 10; r < 10; r++ {
						if rng.Float64() < 0.1 {
							labels := prog.LabelNames()
							if rng.Float64() < 0.5 && len(labels) > 0 {
								outProg.Append(breeder.CompileLabel(labels[rng.Int()%len(labels)]))
							} else {
								outProg.Append(breeder.CompileLabel(":" + USDict.RandomWord()))
							}
						} else {
							outProg = outProg.Append(breeder.Encode(&govirtual.Memory{rng.Int(), rng.SmallInt(), rng.SmallInt(), rng.SmallInt()}))
						}
					}
				}
				if rng.Float64() < 0.1 && outProg.Len() > 0 {
					continue
				}
				if rng.Float64() < 0.1 {
					labels := prog.LabelNames()
					if rng.Float64() < 0.5 && len(labels) > 0 {
						outProg.Append(breeder.CompileLabel(labels[rng.Int()%len(labels)]))
					} else {
						outProg.Append(breeder.CompileLabel(":" + USDict.RandomWord()))
					}
					continue
				}
				decode := op.Decode()
				if rng.Float64() < 0.5 {
					decode.Set(0, rng.Int())
				}
				if rng.Float64() < 0.5 {
					decode.Set(1, rng.SmallInt())
				} else if rng.Float64() < 0.5 {
					decode.Set(1, decode.GetCardinal(1)+1)
				} else if rng.Float64() < 0.5 {
					decode.Set(1, decode.GetCardinal(1)-1)
				}
				if rng.Float64() < 0.5 {
					decode.Set(2, rng.SmallInt())
				} else if rng.Float64() < 0.5 {
					decode.Set(2, decode.GetCardinal(2)+1)
				} else if rng.Float64() < 0.5 {
					decode.Set(2, decode.GetCardinal(2)-1)
				}
				if rng.Float64() < 0.5 {
					decode.Set(3, rng.SmallInt())
				} else if rng.Float64() < 0.5 {
					decode.Set(3, decode.GetCardinal(3)+1)
				} else if rng.Float64() < 0.5 {
					decode.Set(3, decode.GetCardinal(3)-1)
				}
				outProg = outProg.Append(breeder.Encode(decode))
			} else {
				outProg = outProg.Append(op)
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
