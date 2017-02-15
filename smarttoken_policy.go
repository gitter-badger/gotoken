package gotoken

type SmartTokenPolicy interface {
	GetDepth(length int) int
}
