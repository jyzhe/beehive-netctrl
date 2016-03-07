package kandoo

import (
	bh "github.com/jyzhe/beehive"
)

// Implements the map function for nonLocal handlers.
type NonLocal struct{}

func (h NonLocal) Map(msg bh.Msg, ctx bh.MapContext) bh.MappedCells {
	return bh.MappedCells{{"__D__", "__0__"}}
}
