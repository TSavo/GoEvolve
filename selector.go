package goevolve

import (
	"sort"
)

type Selector interface {
	Select(*SolutionList) *SolutionList
}

type AndSelector []Selector

func AndSelect(selectors ...Selector) *AndSelector {
	x := AndSelector(selectors)
	return &x
}

func (multi AndSelector) Select(s *SolutionList) *SolutionList {
	solutions := make(SolutionList, 0)
	for _, x := range multi {
		solutions = append(solutions, *(x).Select(s)...)
	}
	return &solutions
}

func (multi *AndSelector) AddSelector(s Selector) {
	(*multi) = append(*multi, s)
}

type OrSelector []Selector

func OrSelect(selectors ...Selector) *OrSelector {
	x := OrSelector(selectors)
	return &x
}

func (multi OrSelector) Select(s *SolutionList) *SolutionList {
	for _, x := range multi {
		solution := (x).Select(s)
		if len(*solution) > 0 {
			return solution
		}
	}
	return nil

}

type TopXSelector struct {
	Keep int
}

func TopX(keep int) *TopXSelector {
	return &TopXSelector{keep}
}

func (topx TopXSelector) Select(s *SolutionList) *SolutionList {
	sort.Sort(s)
	x := (*s)[:topx.Keep%len(*s)]
	return &x
}

type StochasticUniversalSelector struct {
	Keep int
}

func NewStochasticUniversalSelector(keep int) *StochasticUniversalSelector {
	return &StochasticUniversalSelector{keep}
}

func (sel StochasticUniversalSelector) Select(s *SolutionList) *SolutionList {
	sort.Sort(sort.Reverse(*s))
	f := (*s)[len(*s)-1].Reward
	n := sel.Keep
	p := f / n
	start := rng.Int()%p + 1
	pointers := make([]int, n)
	for i := 0; i < n; i++ {
		pointers[i] = start + i*p
	}
	ret := RWS(s, pointers)
	return ret
}

func RWS(solutions *SolutionList, pointers []int) *SolutionList {
	keep := make(SolutionList, len(pointers))
	i := 0
	for _, p := range pointers {
		for int((*solutions)[i].Reward) < p {
			i++
		}
		keep = append(keep, (*solutions)[i])
	}
	return &keep
}

type TournamentSelector struct {
	Keep int
}

func Tournament(keep int) TournamentSelector {
	return TournamentSelector{keep}
}

func (t TournamentSelector) Select(solutions *SolutionList) *SolutionList {
	keepers := make(SolutionList, 0)
	for x := 0; x < t.Keep; x++ {
		keepers = append(keepers, FightInTournament((*solutions)[rng.Int()%len(*solutions)], (*solutions)[rng.Int()%len(*solutions)]))
	}
	return &keepers
}

func FightInTournament(warrior1 *Solution, warrior2 *Solution) *Solution {
	var highest, lowest *Solution
	if warrior1.Reward >= warrior2.Reward {
		highest, lowest = warrior1, warrior2
	} else {
		highest, lowest = warrior2, warrior1
	}
	if(highest.Reward <= 0) {
		return highest
	}
	if rng.Int()%highest.Reward > lowest.Reward/2 {
		return highest
	} else {
		return lowest
	}
}
