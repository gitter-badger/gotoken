package gotoken

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDepth(t *testing.T) {
	const leftX = 20
	const leftY = 100
	const rightX = 120
	const rightY = 10

	ga := assert.New(t)
	st := NewPolicyDepth(leftX, leftY, rightX, rightY)

	for i := 1; i <= leftX; i++ {
		ga.Equal(st.GetDepth(i), leftY, "depth policy left part")
	}

	memory := leftY
	for i := leftX + 1; i <= rightX; i++ {
		ga.True(st.GetDepth(i) <= memory, "depth policy middle part")
		memory = st.GetDepth(i)
	}

	for i := rightX; i <= rightX+20; i++ {
		ga.Equal(st.GetDepth(i), rightY, "depth policy right part")
	}
}
