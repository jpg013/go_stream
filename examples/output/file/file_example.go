package main

import (
	"github.com/jpg013/go_stream/output"
	"github.com/jpg013/go_stream/stream/readable"
	"github.com/jpg013/go_stream/stream/writable"
	"path"
)

var data = []string{
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
	"Donec quis urna condimentum, pretium quam elementum, tempus odio.",
	"Curabitur ullamcorper orci vel pharetra volutpat.",
	"Integer pellentesque lorem eget libero iaculis, eu pretium sapien bibendum.",
	"Nulla facilisi.",
}

func main() {
	path := path.Join("examples", "output", "file", "file_example")
	rs, err := readable.FromSlice(data)

	if err != nil {
		panic(err.Error())
	}

	conf := output.NewConfig()
	conf.File.Path = path
	conf.File.Delim = ":)\n"

	ws, err := writable.ToFile(conf)

	if err != nil {
		panic(err.Error())
	}

	rs.Pipe(ws)

	// wait for writeble stream to finish
	<-ws.Done()
}
