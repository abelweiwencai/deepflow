package datatype

import (
	"fmt"
)

type TaggedFlow struct {
	Flow
	Tag
}

func (f *TaggedFlow) String() string {
	return fmt.Sprintf("Flow: %+v, Tag: %+v", f.Flow, f.Tag)
}
