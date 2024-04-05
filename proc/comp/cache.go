package comp

type LRUCache struct {
	numberOfLines int
	lineLength    int
	cacheLength   int
	lines         []line
}

type line struct {
	boundary [2]int32
	data     []int8
}

func (l line) get(addr int32) (int8, bool) {
	if addr >= l.boundary[0] && addr < l.boundary[1] {
		return l.data[addr-l.boundary[0]], true
	}
	return 0, false
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
			c.lines = append(append([]line{l}, c.lines[:i]...), c.lines[i+1:]...)
			return v, exists
		}
	}
	return 0, false
}

func (c *LRUCache) Push(addr int32, data []int8) {
	newLine := line{
		boundary: [2]int32{addr, addr + int32(c.lineLength)},
		data:     data,
	}

	c.lines = append([]line{newLine}, c.lines...)
	if len(c.lines) > c.numberOfLines {
		c.lines = c.lines[:c.numberOfLines]
	}
}
