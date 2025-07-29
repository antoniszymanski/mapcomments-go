// SPDX-FileCopyrightText: 2025 Antoni Szymański
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	_ "embed"
	"go/format"
	"os"
	"strconv"
	"text/template"

	"github.com/alecthomas/kong"
	"github.com/antoniszymanski/mapcomments-go"
)

type cli struct {
	Packages []string `arg:""`

	Package string `short:"P" default:"main"`
	Output  string `short:"W" type:"path" default:"commentmap_gen.go"`

	WithFullComment bool
}

//go:embed commentmap.tmpl
var tmplSource string

var tmpl = template.Must(
	template.New("commentmap").
		Funcs(template.FuncMap{"quote": strconv.Quote}).
		Parse(tmplSource),
)

func (c *cli) Run() error {
	commentMap := make(map[string]string)
	for _, path := range c.Packages {
		err := mapcomments.AddGoComments(commentMap, path, c.WithFullComment)
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
		CommentMap: commentMap,
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
	return err
}

func main() {
	var cli cli
	ctx := kong.Parse(&cli,
		kong.Name("mapcomments"),
		kong.UsageOnError(),
	)
	ctx.FatalIfErrorf(ctx.Run())
}
