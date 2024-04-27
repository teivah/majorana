package mvp6_4

import (
	"github.com/teivah/majorana/risc"
)

// TODO Why is 100 blocking?
const maxCacheControllers = 1000

type msiState = int32

const (
	invalid msiState = iota // Default has to be invalid
	shared
	modified
)

type cacheController struct {
	ctx *risc.Context
	id  int
	mmu *memoryManagementUnit
}

func newCacheController(id int, ctx *risc.Context, mmu *memoryManagementUnit) *cacheController {
	cc := &cacheController{
		ctx: ctx,
		id:  id,
		mmu: mmu,
	}
	return cc
}

func (cc *cacheController) flush() int {
	// TODO
	return 0
}
