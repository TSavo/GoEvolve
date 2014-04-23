package goevolve

import (
	//"fmt"
	"github.com/TSavo/GoVirtual"
	"strconv"
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
	out := make([]string, 0)
	for _, x := range ret {
		if(len(strings.TrimSpace(x)) > 0){
			out = append(out, x)
		}
	}
	return out
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
	for x, y := 0, 0; x < cp.PopulationSize && x < len(initialPop); x++ {
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
		p := ":start\n"

		for y := 0; y < breeder.ProgramLength; y++ {
			i := (*breeder.InstructionSet)[rng.Int()%len(*breeder.InstructionSet)]
			for i.Infix || strings.HasPrefix(i.Name, ":") {
				i = (*breeder.InstructionSet)[rng.Int()%len(*breeder.InstructionSet)]
			}
			p += i.Name + " " + ArgsForInstruction(i, []string{":start"}) + "\n"
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

func ArgsForInstruction(op *govirtual.Instruction, labels []string) string {
	args := make([]string, len(op.Arguments))
	for x, arg := range op.Arguments {
		switch arg.Type {
		case "ref":
			args[x] = "#" + strconv.Itoa(rng.SmallInt())
		case "string":
			args[x] = labels[rng.Int()%len(labels)]
		case "int":
			if rng.Float64() < 0.5 {
				args[x] = "#" + strconv.Itoa(rng.SmallInt())
			} else {
				args[x] = strconv.Itoa(rng.SmallInt())
			}
		default:
			args[x] = "0"
		}
	}
	return strings.Join(args, ",")
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
		labels := breeder.CompileProgram(startProg, nil).LabelNames()
		prog := strings.Split(startProg, "\n")
		outProg := ""
		for _, op := range prog {
			op = strings.TrimSpace(op)
			if len(op) < 1 {
				continue
			}
			if rng.Float64() < breeder.MutationChance {
				if rng.Float64() < breeder.MutationChance {
					for r := rng.Int() % 10; r < 10; r++ {
						if rng.Float64() < 0.1 {
							if rng.Float64() < 0.5 && len(labels) > 0 {
								outProg += labels[rng.Int()%len(labels)] + "\n"
							} else {
								outProg += ":" + USDict.RandomWord() + "\n"
							}
						} else {
							i := (*breeder.InstructionSet)[rng.Int()%len(*breeder.InstructionSet)]
							for i.Infix || strings.HasPrefix(i.Name, ":") {
								i = (*breeder.InstructionSet)[rng.Int()%len(*breeder.InstructionSet)]
							}
							args := ArgsForInstruction(i, labels)
							outProg += i.Name + " " + args + "\n"
						}
					}
				}
				if rng.Float64() < 0.1 && len(outProg) > 0 {
					continue
				}
				if rng.Float64() < 0.1 {
					if rng.Float64() < 0.5 && len(labels) > 0 {
						outProg += labels[rng.Int()%len(labels)] + "\n"
					} else {
						outProg += ":" + USDict.RandomWord() + "\n"
					}
					continue
				}
				i := (*breeder.InstructionSet)[rng.Int()%len(*breeder.InstructionSet)]
				for i.Infix || strings.HasPrefix(i.Name, ":") {
					i = (*breeder.InstructionSet)[rng.Int()%len(*breeder.InstructionSet)]
				}
				parts := strings.Split(op, " ")
				if rng.Float64() > 0.5 && strings.HasPrefix(parts[0], ":") {
					if rng.Float64() > 0.5 && len(labels) > 0 {
						outProg += labels[rng.Int()%len(labels)] + "\n"
						continue
					} else if rng.Float64() > 0.5 {
						outProg += ":" + USDict.RandomWord() + "\n"
						continue
					}
				} else if rng.Float64() > 0.5 && !strings.HasPrefix(parts[0], ":") {
					i = (*breeder.InstructionSet).Compile(parts[0]).Instruction
				}
				if strings.HasPrefix(parts[0], ":") {
					outProg += parts[0]
				} else {
					outProg += parts[0] + " " + ArgsForInstruction(i, labels) + "\n"
				}
			} else {
				outProg += op + "\n"
			}
		}
		y++
		out = append(out, outProg)
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
