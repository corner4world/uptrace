// Code generated by "stringer -type=TokenID"; DO NOT EDIT.

package uql

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EOF_TOKEN-0]
	_ = x[BYTE_TOKEN-1]
	_ = x[IDENT_TOKEN-2]
	_ = x[NUMBER_TOKEN-3]
	_ = x[DURATION_TOKEN-4]
	_ = x[BYTES_TOKEN-5]
	_ = x[VALUE_TOKEN-6]
}

const _TokenID_name = "EOF_TOKENBYTE_TOKENIDENT_TOKENNUMBER_TOKENDURATION_TOKENBYTES_TOKENVALUE_TOKEN"

var _TokenID_index = [...]uint8{0, 9, 19, 30, 42, 56, 67, 78}

func (i TokenID) String() string {
	if i < 0 || i >= TokenID(len(_TokenID_index)-1) {
		return "TokenID(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenID_name[_TokenID_index[i]:_TokenID_index[i+1]]
}
