package gotoken

type TokenizationPolicy interface {
	GetDepth(length int) int
}
