package geosite

import (
	"strings"
)

type RecordType int

const (
	RECORD_KEYWORD RecordType = 0
	RECORD_REGEXP  RecordType = 1
	RECORD_DOMAIN  RecordType = 2
	RECORD_FULL    RecordType = 3
)

// An attribute has either bool or int64 value.
// But for now only bool(true) is present,
// so here nil means true, non-nil means actual integer.
type RecordAttr map[string]*int64

type Record struct {
	Type  RecordType
	Value string
	Attr  RecordAttr
}

type RecordList []*Record

// An entry is stripped-down to country code and a list of domains.
// Here they are used as key and value pairs.
type Collection map[string]RecordList

func ParseRecordType(buf []byte) (RecordType, []byte, int) {
	val, buf, n := ParseVarint(buf)
	if n == PARSE_ERR || val > 3 {
		return 0, buf, PARSE_ERR
	}
	return RecordType(val), buf, n
}

func ParseRecordAttr(buf []byte) (string, *int64, []byte, int) {
	var attr string
	var value *int64
	size, buf, ns := ParseVarint(buf)
	if ns == PARSE_ERR {
		return "", nil, buf, PARSE_ERR
	}
	for left := int(size); left > 0; {
		var field uint64
		var nk, nv int
		field, _, buf, nk = ParseKey(buf)
		if nk == PARSE_ERR {
			return "", nil, buf, PARSE_ERR
		}
		switch field {
		case 1:
			attr, buf, nv = ParseString(buf)
		case 2:
			_, buf, nv = ParseBool(buf)
		case 3:
			var intValue int64
			intValue, buf, nv = ParseSigned(buf)
			value = &intValue
		}
		if nv == PARSE_ERR || left < nk+nv {
			return "", nil, buf, PARSE_ERR
		}
		left -= nk + nv
	}
	return strings.ToLower(attr), value, buf, ns + int(size)
}

func ParseRecord(buf []byte) (*Record, []byte, int) {
	var record = &Record{}
	size, buf, ns := ParseVarint(buf)
	if ns == PARSE_ERR {
		return nil, buf, PARSE_ERR
	}
	for left := int(size); left > 0; {
		var field uint64
		var nk, nv int
		field, _, buf, nk = ParseKey(buf)
		if nk == PARSE_ERR {
			return nil, buf, PARSE_ERR
		}
		switch field {
		case 1:
			record.Type, buf, nv = ParseRecordType(buf)
		case 2:
			record.Value, buf, nv = ParseString(buf)
		case 3:
			if record.Attr == nil {
				record.Attr = make(map[string]*int64)
			}
			var name string
			var value *int64
			name, value, buf, nv = ParseRecordAttr(buf)
			record.Attr[name] = value
		}
		if nv == PARSE_ERR || left < nk+nv {
			return nil, buf, PARSE_ERR
		}
		left -= nk + nv
	}
	return record, buf, ns + int(size)
}

func ParseRecordList(buf []byte) (string, RecordList, []byte, int) {
	var name string
	var list = make(RecordList, 0)
	size, buf, ns := ParseVarint(buf)
	if ns == PARSE_ERR {
		return "", nil, buf, PARSE_ERR
	}
	for left := int(size); left > 0; {
		var field uint64
		var nk, nv int
		field, _, buf, nk = ParseKey(buf)
		if nk == PARSE_ERR {
			return "", nil, buf, PARSE_ERR
		}
		switch field {
		case 1:
			name, buf, nv = ParseString(buf)
		case 2:
			var record *Record
			record, buf, nv = ParseRecord(buf)
			list = append(list, record)
		}
		if nv == PARSE_ERR || left < nk+nv {
			return "", nil, buf, PARSE_ERR
		}
		left -= nk + nv
	}
	return strings.ToLower(name), list, buf, ns + int(size)
}

// Read all data from buf and parse into Collection.
// Returns nil on error.
func ParseCollection(buf []byte) Collection {
	var collection = make(Collection)
	for len(buf) > 0 {
		var nk, nv int
		_, _, buf, nk = ParseKey(buf)
		if nk == PARSE_ERR {
			return nil
		}
		var name string
		var list []*Record
		name, list, buf, nv = ParseRecordList(buf)
		if nv == PARSE_ERR {
			return nil
		}
		collection[name] = list
	}
	return collection
}
