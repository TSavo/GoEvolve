package solve

import (
	"time"
	"github.com/tsavo/golightly/vm"
)
type Evaluator interface {
	Evaluate(*vm.ProcessorCore) int64
}

type MultiEvaluator []*Evaluator

func NewMultiEvaluator(e... *Evaluator) *MultiEvaluator {
	m := &MultiEvaluator{}
	for _, x := range e {
		m.AddEvaluator(x)
	}
	return m
}

func (multi *MultiEvaluator) Evaluate(p *vm.ProcessorCore) int64 {
	e := int64(0)
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
	Evaluator
}

func Inverse(e *Evaluator) *InverseEvaluator {
	return &InverseEvaluator{*e}
}

func (inverse *InverseEvaluator) Evaluate(p *vm.ProcessorCore) int64 {
	return inverse.Evaluator.Evaluate(p) * -1
}

type TimeEvaluator struct {}

func NewTimeEvaluator() *TimeEvaluator {
	return &TimeEvaluator{}
}

func (t *TimeEvaluator) Evaluate(p *vm.ProcessorCore) int64 {
	return time.Now().UnixNano() - p.StartTime
}

type CostEvaluator struct {}

func NewCostEvaluator() *CostEvaluator {
	return &CostEvaluator{}
}

func (c *CostEvaluator) Evaluate(p *vm.ProcessorCore) int64 {
	return int64(p.Cost())
}