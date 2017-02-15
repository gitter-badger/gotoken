package gotoken

type PolicyDepth struct {
	maxLength int
	maxDepth  int
	minLength int
	minDepth  int
}

func NewPolicyDepth(maxLength int, maxDepth int, minLength int, minDepth int) SmartTokenPolicy {
	return &PolicyDepth{
		maxLength: maxLength,
		maxDepth:  maxDepth,
		minLength: minLength,
		minDepth:  minDepth,
	}
}

func (p *PolicyDepth) GetDepth(length int) int {
	// TODO: It seems min & max conditoins are reversed???
	if length <= p.maxLength {
		return p.maxDepth
	} else if length >= p.minLength {
		return p.minDepth
	}

	// Linear interpolation
	return int(float64(p.minDepth) + float64(p.maxDepth-p.minDepth)*float64(p.minLength-length)/float64(p.minLength-p.maxLength))
}
