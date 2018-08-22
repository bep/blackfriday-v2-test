package main

import (
	"bytes"
	"fmt"
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

	var mod RenderModifier

	fmt.Println(">>", mod)

	var b bytes.Buffer
	run(md, &b)
	b.WriteTo(f)

	//fmt.Println("BlackFriday v2 Test:\n", string(output))
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

	modsMap := make(map[*blackfriday.Node][]RenderModifier)

	// The hooks may be stateful.
	var getRenderMods = func(node *blackfriday.Node, entering bool) []RenderModifier {
		var mods []RenderModifier
		if entering {
			mods = GetRenderMods(node.Type)
			if mods != nil {
				modsMap[node] = mods
			}
		} else {
			mods = modsMap[node]
		}
		return mods
	}

	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {

		mods := getRenderMods(node, entering)

		if entering {
			for _, mod := range mods {
				status := mod(out, node, RenderStart)
				if status == WalkStatusDone {
					return blackfriday.SkipChildren
				}
			}
		}

		if status := r.RenderNode(out, node, entering); status != blackfriday.GoToNext {
			return status
		}

		for _, mod := range mods {
			state := RenderEntered
			if !entering {
				state = RenderDone
			}
			status := mod(out, node, state)
			if status == WalkStatusDone {
				return blackfriday.SkipChildren
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
