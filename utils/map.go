package utils

import (
	"hugobot/types"

	"github.com/fatih/structs"
)

func StructToJsonMap(in interface{}) types.JsonMap {
	out := make(types.JsonMap)

	s := structs.New(in)
	for _, f := range s.Fields() {
		out[f.Tag("json")] = f.Value()
	}

	return out
}
