package branding

import (
	"encoding/json"
	"strconv"
)

// extraData contains the exact keys consumed by Nitro room branding logic.
type extraData struct {
	// State stores the visual state.
	State string `json:"state"`
	// ImageURL stores the public image URL.
	ImageURL string `json:"imageUrl"`
	// ClickURL stores the optional billboard URL.
	ClickURL string `json:"clickUrl"`
	// OffsetX stores the horizontal renderer offset.
	OffsetX string `json:"offsetX"`
	// OffsetY stores the vertical renderer offset.
	OffsetY string `json:"offsetY"`
	// OffsetZ stores the depth renderer offset.
	OffsetZ string `json:"offsetZ"`
}

// encodeExtraData creates canonical furniture state for Nitro projection.
func encodeExtraData(mutation Mutation, enabled bool) string {
	data := extraData{
		State:    strconv.Itoa(int(mutation.State)),
		ImageURL: mutation.ImageURL,
		ClickURL: mutation.ClickURL,
		OffsetX:  strconv.Itoa(mutation.OffsetX),
		OffsetY:  strconv.Itoa(mutation.OffsetY),
		OffsetZ:  strconv.Itoa(mutation.OffsetZ),
	}
	if !enabled {
		data = extraData{State: "0", OffsetX: "0", OffsetY: "0", OffsetZ: "0"}
	}
	encoded, _ := json.Marshal(data)
	return string(encoded)
}
