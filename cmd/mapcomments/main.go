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

	"github.com/alexflint/go-arg"
	"github.com/antoniszymanski/mapcomments-go"
)

var args struct {
	Packages        []string `arg:"required,positional"`
	Package         string   `arg:"-P" default:"main"`
	Output          string   `arg:"-O" default:"commentmap_gen.go"`
	WithFullComment bool     `arg:"--with-full-comment"`
}

//go:embed commentmap.tmpl
var tmplSource string

var tmpl = template.Must(
	template.New("commentmap").
		Funcs(template.FuncMap{"quote": strconv.Quote}).
		Parse(tmplSource),
)

func main() {
	cfg := arg.Config{
		Program: "mapcomments",
		Out:     os.Stderr,
	}
	p, err := arg.NewParser(cfg, &args)
	if err != nil {
		panic(err)
	}
	var flags []string
	if len(os.Args) > 0 {
		flags = os.Args[1:]
	}
	switch err = p.Parse(flags); {
	case err == arg.ErrHelp:
		p.WriteHelp(cfg.Out)
		os.Exit(0) //nolint:gocritic // exitAfterDefer
	case err != nil:
		p.WriteHelp(cfg.Out)
		printErr(err)
		os.Exit(2)
	}
	if err = run(); err != nil {
		printErr(err)
	}
}

func run() error {
	commentMap := make(map[string]string)
	for _, path := range args.Packages {
		err := mapcomments.AddGoComments(commentMap, path, args.WithFullComment)
		if err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct {
		Package    string
		CommentMap map[string]string
	}{
		Package:    args.Package,
		CommentMap: commentMap,
	}); err != nil {
		return err
	}

	data, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	if args.Output != "-" {
		err = os.WriteFile(args.Output, data, 0600)
	} else {
		_, err = os.Stdout.Write(data)
	}
	return err
}

//nolint:errcheck
func printErr(err error) {
	os.Stderr.WriteString("error: ")
	os.Stderr.WriteString(err.Error())
	os.Stderr.WriteString("\n")
}
