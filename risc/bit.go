package risc

func BytesFromLowBits(n int32) [4]int8 {
	var i1, i2, i3, i4 int8
	var index uint8

	for i := 0; i < 8; i++ {
		if getI32Bit(n, uint8(i)) {
			i1 = setI8Bit(i1, index)
		}
		index++
	}

	index = 0
	for i := 8; i < 16; i++ {
		if getI32Bit(n, uint8(i)) {
			i2 = setI8Bit(i2, index)
		}
		index++
	}

	index = 0
	for i := 16; i < 24; i++ {
		if getI32Bit(n, uint8(i)) {
			i3 = setI8Bit(i3, index)
		}
		index++
	}

	index = 0
	for i := 24; i < 32; i++ {
		if getI32Bit(n, uint8(i)) {
			i4 = setI8Bit(i4, index)
		}
		index++
	}

	return [4]int8{i1, i2, i3, i4}
}

func i32FromBytes(i1, i2, i3, i4 int8) int32 {
	var index uint8
	var result int32

	for i := 0; i < 8; i++ {
		if getI8Bit(i1, uint8(i)) {
			result = setI32Bit(result, index)
		}
		index++
	}

	for i := 0; i < 8; i++ {
		if getI8Bit(i2, uint8(i)) {
			result = setI32Bit(result, index)
		}
		index++
	}

	for i := 0; i < 8; i++ {
		if getI8Bit(i3, uint8(i)) {
			result = setI32Bit(result, index)
		}
		index++
	}

	for i := 0; i < 8; i++ {
		if getI8Bit(i4, uint8(i)) {
			result = setI32Bit(result, index)
		}
		index++
	}

	return result
}

func getI8Bit(input int8, n uint8) bool {
	return input&(1<<n) != 0
}

func getI32Bit(input int32, n uint8) bool {
	return input&(1<<n) != 0
}

func setI8Bit(n int8, i uint8) int8 {
	return n | (1 << int8(i))
}

func setI32Bit(n int32, i uint8) int32 {
	return n | (1 << int8(i))
}
