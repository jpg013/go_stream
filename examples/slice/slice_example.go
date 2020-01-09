package main

import (
	"stream/readable"
	"stream/writable"
)

var data = []string{
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
	"Donec quis urna condimentum, pretium quam elementum, tempus odio.",
	"Curabitur ullamcorper orci vel pharetra volutpat.",
	"Integer pellentesque lorem eget libero iaculis, eu pretium sapien bibendum.",
	"Nulla facilisi.",
}

func main() {
	r, err := readable.FromSlice(data)

	if err != nil {
		panic(err.Error())
	}

	w, err := writable.ToSTDOUT()

	if err != nil {
		panic(err.Error())
	}

	r.Pipe(w)
}
