package comp

import (
	"fmt"
	"strings"
)

type AlignedAddress int32

type LRUCache struct {
	numberOfLines int
	lineLength    int
	cacheLength   int
	lines         []Line
}

type Line struct {
	Boundary [2]AlignedAddress
	Data     []int8
}

func (l Line) String() string {
	return fmt.Sprintf("(%d-%d): %v", l.Boundary[0], l.Boundary[1], l.Data)
}

func (l Line) get(addr int32) (int8, bool) {
	if addr >= int32(l.Boundary[0]) && addr < int32(l.Boundary[1]) {
		return l.Data[addr-int32(l.Boundary[0])], true
	}
	return 0, false
}

func (l Line) set(addr int32, value int8) {
	l.Data[addr-int32(l.Boundary[0])] = value
}

func NewLRUCache(lineLength int, cacheLength int) *LRUCache {
	if cacheLength%lineLength != 0 {
		panic("cache length should be a multiple of the line length")
	}
	return &LRUCache{
		numberOfLines: cacheLength / lineLength,
		lineLength:    lineLength,
		cacheLength:   cacheLength,
	}
}

func (c *LRUCache) Get(addr int32) (int8, bool) {
	for i, l := range c.lines {
		if v, exists := l.get(addr); exists {
			c.lines = append(append([]Line{l}, c.lines[:i]...), c.lines[i+1:]...)
			return v, true
		}
	}
	return 0, false
}

func (c *LRUCache) GetCacheLine(addr AlignedAddress) ([]int8, bool) {
	for _, l := range c.lines {
		if _, exists := l.get(int32(addr)); exists {
			return l.Data, true
		}
	}
	return nil, false
}

func (c *LRUCache) EvictCacheLine(addr AlignedAddress) ([]int8, bool) {
	for i, l := range c.lines {
		if _, exists := l.get(int32(addr)); exists {
			c.lines = append(c.lines[:i], c.lines[i+1:]...)
			return l.Data, true
		}
	}
	return nil, false
}

func (c *LRUCache) Write(addr int32, data []int8) {
	for _, l := range c.lines {
		if _, exists := l.get(addr); exists {
			for i, v := range data {
				l.set(addr+int32(i), v)
			}
			fmt.Println(addr/4, l)
			return
		}
	}
	panic("cache line doesn't exist")
}

func (c *LRUCache) PushLine(addr AlignedAddress, data []int8) []int8 {
	newLine := Line{
		Boundary: [2]AlignedAddress{addr, addr + AlignedAddress(c.lineLength)},
		Data:     data,
	}

	c.lines = append([]Line{newLine}, c.lines...)
	if len(c.lines) > c.numberOfLines {
		c.lines = c.lines[:c.numberOfLines]
		// Return the evicted line
		return c.lines[len(c.lines)-1].Data
	}
	return nil
}

func (c *LRUCache) PushLineWithEvictionWarning(addr AlignedAddress, data []int8) *Line {
	newLine := Line{
		Boundary: [2]AlignedAddress{addr, addr + AlignedAddress(c.lineLength)},
		Data:     data,
	}

	c.lines = append([]Line{newLine}, c.lines...)
	if len(c.lines) > c.numberOfLines {
		line := c.lines[len(c.lines)-1]
		return &line
	}
	return nil
}

func (c *LRUCache) Lines() []Line {
	return c.lines
}

func (c *LRUCache) String() string {
	res := make([]string, 0, len(c.lines))
	for _, line := range c.lines {
		res = append(res, line.String())
	}
	return strings.Join(res, "\n")
}
