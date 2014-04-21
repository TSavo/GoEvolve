package goevolve

import (
	"github.com/TSavo/GoVirtual"
)
type Evaluator interface {
	Evaluate(*govirtual.Processor) int
}

type MultiEvaluator []*Evaluator

func NewMultiEvaluator(e... *Evaluator) *MultiEvaluator {
	m := &MultiEvaluator{}
	for _, x := range e {
		m.AddEvaluator(x)
	}
	return m
}

func (multi *MultiEvaluator) Evaluate(p *govirtual.Processor) int {
	e := int(0)
	for _, x := range *multi {
		e += (*x).Evaluate(p)
	}
	return e
}

func (multi *MultiEvaluator) AddEvaluator(e *Evaluator) *MultiEvaluator {
	*multi = append(*multi, e)
	return multi 
}

type InverseEvaluator struct {
	Evaluator *Evaluator
}

func Inverse(e Evaluator) *InverseEvaluator {
	return &InverseEvaluator{&e}
}

func (inverse InverseEvaluator) Evaluate(p *govirtual.Processor) int {
	return (*inverse.Evaluator).Evaluate(p) * -1
}

type TimeEvaluator struct {}

func NewTimeEvaluator() *TimeEvaluator {
	return &TimeEvaluator{}
}

func (t *TimeEvaluator) Evaluate(p *govirtual.Processor) int {
	return Now() - p.StartTime
}

type CostEvaluator struct {}

func NewCostEvaluator() *CostEvaluator {
	return &CostEvaluator{}
}

func (c *CostEvaluator) Evaluate(p *govirtual.Processor) int {
	return p.Cost()
}