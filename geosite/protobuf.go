package geosite

const PARSE_ERR = 0

func Read(buf []byte, n int) ([]byte, []byte, int) {
	if n > len(buf) || n < 1 {
		return nil, buf, PARSE_ERR
	}
	return buf[:n], buf[n:], n
}

func ParseVarint(buf []byte) (uint64, []byte, int) {
	one, buf, n := Read(buf, 1)
	if n == PARSE_ERR {
		return 0, buf, PARSE_ERR
	}
	val := uint64(one[0] & 0b1111111)
	size := n
	for one[0]>>7 == 1 {
		one, buf, n = Read(buf, 1)
		if n == PARSE_ERR || size > 9 {
			return 0, buf, PARSE_ERR
		}
		val += uint64(one[0]&0b1111111) << (size * 7)
		size += n
	}
	return val, buf, size
}

func ParseKey(buf []byte) (uint64, byte, []byte, int) {
	val, buf, n := ParseVarint(buf)
	if n == PARSE_ERR {
		return 0, 0, buf, PARSE_ERR
	}
	return val >> 3, byte(val & 0b111), buf, n
}
func ParseBool(buf []byte) (bool, []byte, int) {
	val, buf, n := ParseVarint(buf)
	if n == PARSE_ERR {
		return false, buf, PARSE_ERR
	}
	return val != 0, buf, n
}

func ParseSigned(buf []byte) (int64, []byte, int) {
	val, buf, n := ParseVarint(buf)
	half := int64(val / 2)
	if uint64(half)*2 < val {
		half = -half - 1
	}
	return half, buf, n
}

func ParseString(buf []byte) (string, []byte, int) {
	strlen, buf, n0 := ParseVarint(buf)
	if n0 == PARSE_ERR {
		return "", buf, PARSE_ERR
	}
	binary, buf, n1 := Read(buf, int(strlen))
	if n1 == PARSE_ERR {
		return "", buf, PARSE_ERR
	}
	return string(binary), buf, n0 + n1
}
