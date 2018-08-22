package main

import (
	"bytes"
	"log"
	"os"

	"github.com/russross/blackfriday"
)

func main() {

	f, err := os.Create("/Users/bep/sites/dump/bf.html")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var b bytes.Buffer
	run(md, &b)
	b.WriteTo(f)

	//fmt.Println("BlackFriday v2 Test:\n", string(output))
}

type mods struct {
	pos       int
	modifiers []RenderModifier
}

func run(input string, out *bytes.Buffer, opts ...blackfriday.Option) {
	r := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	})
	optList := []blackfriday.Option{
		blackfriday.WithRenderer(r),
		blackfriday.WithExtensions(blackfriday.CommonExtensions | blackfriday.Footnotes)}

	optList = append(optList, opts...)
	parser := blackfriday.New(optList...)
	ast := parser.Parse([]byte(input))
	r.RenderHeader(out, ast)

	modsMap := make(map[*blackfriday.Node]*mods)

	getMods := func(node *blackfriday.Node) *mods {
		if m, found := modsMap[node]; found {
			return m
		}
		rm := GetRenderMods(node.Type)
		if rm == nil {
			return nil
		}
		m := &mods{modifiers: rm, pos: out.Len()}
		modsMap[node] = m
		return m
	}

	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {

		mods := getMods(node)

		if entering && mods != nil {
			for _, mod := range mods.modifiers {
				status := mod(out, mods.pos, node, RenderStart)
				if status == WalkStatusDone {
					return blackfriday.SkipChildren
				}
			}
		}

		if status := r.RenderNode(out, node, entering); status != blackfriday.GoToNext {
			return status
		}

		if mods != nil {
			for _, mod := range mods.modifiers {
				state := RenderEntered
				if !entering {
					state = RenderDone
				}
				status := mod(out, mods.pos, node, state)
				if status == WalkStatusDone {
					return blackfriday.SkipChildren
				}
			}
		}

		return blackfriday.GoToNext
	})

	r.RenderFooter(out, ast)
}

// Hugo custom
// BlockCode (code highlighting) => Node type Code
// ListItem, List (todo)
//

const md = `

## Heading 1

Some **text**.

## Heading 2

This is a footnote.[^fn1]

A fenced code block:

` + "```" + `go
type RenderModifier interface {
}
` + "```" + `

### Heading 2-1

Task list:

- [x] Finish my changes
- [ ] Push my commits to GitHub
- [ ] Open a pull request

Some more text.[^fn2]

[^fn1]: the footnote text.
[^fn2]: the footnote 2 text.

`
