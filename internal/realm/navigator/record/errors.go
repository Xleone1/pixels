package record

import "errors"

var (
	// ErrFavoriteUnavailable reports an invalid, duplicate, or over-limit favorite.
	ErrFavoriteUnavailable = errors.New("navigator favorite unavailable")
	// ErrLiftedRoomConflict reports an optimistic media version mismatch.
	ErrLiftedRoomConflict = errors.New("navigator lifted room version conflict")
	// ErrLiftedRoomNotFound reports missing active Navigator room media.
	ErrLiftedRoomNotFound = errors.New("navigator lifted room not found")
)
