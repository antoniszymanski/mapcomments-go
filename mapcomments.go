/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

package mapcomments

import (
	"errors"
	"go/ast"
	"go/doc"
	"strings"

	"golang.org/x/tools/go/packages"
)

func AddGoComments(commentMap map[string]string, path string, withFullComment bool) error {
	// https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns
	// https://pkg.go.dev/golang.org/x/tools/go/packages#pkg-overview
	switch path {
	case "main", "all", "std", "cmd", "tool":
		return errors.New("path cannot be a reserved name")
	}
	if strings.Contains(path, "...") {
		return errors.New("path cannot contain wildcards")
	}

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax,
	}
	pkgs, err := packages.Load(cfg, "pattern="+path)
	if err != nil {
		return err
	}
	pkg := pkgs[0]
	if err = newPackageError(pkg); err != nil {
		return err
	}

	var pkgDoc *doc.Package
	if !withFullComment {
		pkgDoc, err = doc.NewFromFiles(pkg.Fset, pkg.Syntax, pkg.PkgPath, doc.PreserveAST)
		if err != nil {
			return err
		}
	}

	for _, file := range pkg.Syntax {
		var typeName, typeText string
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.TypeSpec:
				if !node.Name.IsExported() {
					typeName = ""
					break
				}
				typeName = node.Name.Name
				text := node.Doc.Text()
				if text == "" && typeText != "" {
					text = typeText
					typeText = ""
				}
				if !withFullComment {
					text = pkgDoc.Synopsis(text)
					// Synopsis removes leading and trailing white spaces
				} else {
					text = strings.TrimSpace(text)
				}
				if text != "" {
					commentMap[pkg.PkgPath+"."+typeName] = text
				}
			case *ast.Field:
				text := node.Doc.Text()
				if text == "" {
					text = node.Comment.Text()
				}
				text = strings.TrimSpace(text)
				if typeName == "" || text == "" {
					break
				}
				for _, name := range node.Names {
					if name.IsExported() {
						fieldName := name.Name
						commentMap[pkg.PkgPath+"."+typeName+"."+fieldName] = text
					}
				}
			case *ast.GenDecl:
				typeText = node.Doc.Text() // remember for the next type
			}
			return true
		})
	}
	return nil
}

func newPackageError(pkg *packages.Package) error {
	if len(pkg.Errors) == 0 && (pkg.Module == nil || pkg.Module.Error == nil) {
		return nil
	}

	var err PackageError
	err.Errors = pkg.Errors
	if pkg.Module != nil && pkg.Module.Error != nil {
		err.ModuleError = pkg.Module.Error
	}
	return err
}

type PackageError struct {
	Errors      []packages.Error
	ModuleError *packages.ModuleError
}

// Based on [packages.PrintErrors]
func (e PackageError) Error() string {
	var sb strings.Builder

	for _, pkgErr := range e.Errors {
		sb.WriteString(pkgErr.Error())
		sb.WriteByte('\n')
	}
	if e.ModuleError != nil {
		sb.WriteString(e.ModuleError.Err)
		sb.WriteByte('\n')
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

func (e PackageError) Unwrap() []error {
	pkgErrs := make([]error, 0, len(e.Errors))
	for _, pkgErr := range e.Errors {
		pkgErrs = append(pkgErrs, pkgErr)
	}
	return pkgErrs
}
