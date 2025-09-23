// SPDX-FileCopyrightText: 2025 Antoni Szymański
// SPDX-License-Identifier: MPL-2.0

package mapcomments

import (
	"errors"
	"go/ast"
	"go/doc"
	"strings"

	"github.com/antoniszymanski/loadpackage-go"
	"golang.org/x/tools/go/packages"
)

func AddGoComments(commentMap map[string]string, path string, withFullComment bool) error {
	if commentMap == nil {
		return errors.New("commentMap is nil")
	}

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax,
	}
	pkg, err := loadpackage.Load("pattern="+path, cfg)
	if err != nil {
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
