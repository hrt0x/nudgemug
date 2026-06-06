package main

import "encoding/binary"

func iconICO(active bool) []byte {
	const (
		width      = 16
		height     = 16
		pixelBytes = width * height * 4
		maskBytes  = width / 8 * height * 2
		dibBytes   = 40 + pixelBytes + maskBytes
		totalBytes = 6 + 16 + dibBytes
	)

	ico := make([]byte, totalBytes)
	binary.LittleEndian.PutUint16(ico[2:], 1)
	binary.LittleEndian.PutUint16(ico[4:], 1)

	entry := ico[6:]
	entry[0] = width
	entry[1] = height
	binary.LittleEndian.PutUint16(entry[4:], 1)
	binary.LittleEndian.PutUint16(entry[6:], 32)
	binary.LittleEndian.PutUint32(entry[8:], dibBytes)
	binary.LittleEndian.PutUint32(entry[12:], 22)

	dib := ico[22:]
	binary.LittleEndian.PutUint32(dib[0:], 40)
	binary.LittleEndian.PutUint32(dib[4:], width)
	binary.LittleEndian.PutUint32(dib[8:], height*2)
	binary.LittleEndian.PutUint16(dib[12:], 1)
	binary.LittleEndian.PutUint16(dib[14:], 32)
	binary.LittleEndian.PutUint32(dib[20:], pixelBytes+maskBytes)

	pixels := dib[40:]
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			topY := height - 1 - y
			offset := (y*width + x) * 4
			r, g, b, a := iconPixel(x, topY, active)
			pixels[offset+0] = b
			pixels[offset+1] = g
			pixels[offset+2] = r
			pixels[offset+3] = a
		}
	}

	return ico
}

func iconPixel(x, y int, active bool) (byte, byte, byte, byte) {
	bg := func() (byte, byte, byte, byte) { return 0, 0, 0, 0 }
	cup := func() (byte, byte, byte, byte) {
		if active {
			return 245, 238, 224, 255
		}
		return 164, 170, 178, 255
	}
	edge := func() (byte, byte, byte, byte) {
		if active {
			return 128, 84, 46, 255
		}
		return 84, 91, 101, 255
	}
	steam := func() (byte, byte, byte, byte) {
		if active {
			return 217, 132, 58, 255
		}
		return 118, 124, 133, 210
	}

	if active && ((x == 5 && y >= 1 && y <= 3) || (x == 8 && y >= 0 && y <= 2) || (x == 11 && y >= 1 && y <= 3)) {
		return steam()
	}
	if y == 13 && x >= 3 && x <= 12 {
		return edge()
	}
	if y == 12 && x >= 4 && x <= 11 {
		return cup()
	}
	if x >= 4 && x <= 10 && y >= 6 && y <= 11 {
		if x == 4 || x == 10 || y == 6 || y == 11 {
			return edge()
		}
		return cup()
	}
	if ((x == 11 || x == 13) && y >= 7 && y <= 10) || ((y == 7 || y == 10) && x >= 11 && x <= 13) {
		return edge()
	}
	return bg()
}
