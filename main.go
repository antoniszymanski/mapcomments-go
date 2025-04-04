package main

import (
	"bytes"
	_ "embed"
	"go/format"
	"os"
	"strconv"
	"text/template"

	"github.com/alecthomas/kong"
	"github.com/invopop/jsonschema"
)

type cli struct {
	Entries []entry `arg:""`

	Package string `short:"P" default:"main"`
	Output  string `short:"W" type:"path" default:"commentmap_gen.go"`

	WithFullComment bool
}

type entry struct {
	Base, Path string
}

func (e *entry) Decode(ctx *kong.DecodeContext) error {
	err := ctx.Scan.PopValueInto("base", &e.Base)
	if err != nil {
		return err
	}

	return ctx.Scan.PopValueInto("path", &e.Path)
}

//go:embed mapcomments.go.tmpl
var tmplSource string

var tmpl = template.Must(
	template.New("mapcomments").
		Funcs(template.FuncMap{"quote": strconv.Quote}).
		Parse(tmplSource),
)

func (c *cli) Run() error {
	var opts []jsonschema.CommentOption
	if c.WithFullComment {
		opts = append(opts, jsonschema.WithFullComment())
	}

	var r jsonschema.Reflector
	for _, entry := range c.Entries {
		err := r.AddGoComments(entry.Base, entry.Path, opts...)
		if err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct {
		Package    string
		CommentMap map[string]string
	}{
		Package:    c.Package,
		CommentMap: r.CommentMap,
	}); err != nil {
		return err
	}

	data, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	if c.Output != "-" {
		err = os.WriteFile(c.Output, data, 0600)
	} else {
		_, err = os.Stdout.Write(data)
	}
	if err != nil {
		return err
	}

	return nil
}

func main() {
	var cli cli
	ctx := kong.Parse(&cli,
		kong.Name("mapcomments"),
		kong.UsageOnError(),
	)
	ctx.FatalIfErrorf(ctx.Run())
}
