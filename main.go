package main

import (
	"io"
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
	run(md, f)

	//fmt.Println("BlackFriday v2 Test:\n", string(output))
}

func run(input string, out io.Writer, opts ...blackfriday.Option) {
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
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		//fmt.Printf("%d:%d:%s:%v\t\t%s\n", node.Level, node.NoteID, node.Destination, node.Type, node.String())
		return r.RenderNode(out, node, entering)
	})
	r.RenderFooter(out, ast)
}

// Hugo custom
// BlockCode (code highlighting) => Node type Code
// ListItem, List (todo)
//

type RenderModifier interface {
}

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
