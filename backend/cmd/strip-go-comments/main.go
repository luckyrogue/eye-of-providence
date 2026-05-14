package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := flag.String("root", ".", "repository root to walk")
	dry := flag.Bool("dry", false, "print paths only, do not write")
	flag.Parse()
	abs, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var n int
	err = filepath.WalkDir(abs, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "node_modules" || base == "vendor" || base == ".cursor" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		if *dry {
			fmt.Println(path)
			return nil
		}
		if err := stripFile(path); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		n++
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !*dry {
		fmt.Printf("updated %d Go files\n", n)
	}
}

func keepCommentGroup(g *ast.CommentGroup) bool {
	for _, c := range g.List {
		t := strings.TrimSpace(c.Text)
		switch {
		case strings.HasPrefix(t, "//go:"):
			return true
			return true
		case strings.HasPrefix(t, "//line "):
			return true
		case strings.HasPrefix(t, "// #cgo"), strings.HasPrefix(t, "//#cgo"):
			return true
		case strings.HasPrefix(t, "//export "):
			return true
		case strings.HasPrefix(t, "//extern "):
			return true
		}
	}
	return false
}

func stripDocs(n ast.Node) {
	ast.Inspect(n, func(node ast.Node) bool {
		if node == nil {
			return false
		}
		switch x := node.(type) {
		case *ast.Field:
			x.Doc = nil
			x.Comment = nil
		case *ast.ImportSpec:
			x.Doc = nil
			x.Comment = nil
		case *ast.ValueSpec:
			x.Doc = nil
			x.Comment = nil
		case *ast.TypeSpec:
			x.Doc = nil
			x.Comment = nil
		case *ast.FuncDecl:
			x.Doc = nil
		case *ast.GenDecl:
			x.Doc = nil
		case *ast.File:
			x.Doc = nil
		}
		return true
	})
}

func stripFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return err
	}
	var kept []*ast.CommentGroup
	for _, g := range f.Comments {
		if keepCommentGroup(g) {
			kept = append(kept, g)
		}
	}
	f.Comments = kept
	stripDocs(f)

	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	if err := cfg.Fprint(&buf, fset, f); err != nil {
		return err
	}
	out, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format: %w", err)
	}
	return os.WriteFile(path, out, 0o644)
}
