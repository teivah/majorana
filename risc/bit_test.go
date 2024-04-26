package risc

import (
	"fmt"
	"testing"
)

func TestI32FromBytes(t *testing.T) {
	fmt.Println(I32FromBytes(0, 1, 2, 3))
}
