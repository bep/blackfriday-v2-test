package main

import (
	"bytes"
	"fmt"

	"github.com/russross/blackfriday"
)

func main() {
	const md = `

## Heading 1

Some **text**

## Heading 2

This is a footnote.[^fn1]

### Heading 2-1

Some more text.[^fn2]

[^fn1]: the footnote text.
[^fn2]: the footnote 2 text.

`
	run(md)

	//fmt.Println("BlackFriday v2 Test:\n", string(output))
}

func run(input string, opts ...blackfriday.Option) []byte {
	r := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	})
	optList := []blackfriday.Option{
		blackfriday.WithRenderer(r),
		blackfriday.WithExtensions(blackfriday.CommonExtensions | blackfriday.Footnotes)}

	optList = append(optList, opts...)
	parser := blackfriday.New(optList...)
	ast := parser.Parse([]byte(input))
	var buf bytes.Buffer
	r.RenderHeader(&buf, ast)
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		fmt.Printf("%d:%d:%s:%v\t\t%s\n", node.Level, node.NoteID, node.Destination, node.Type, node.String())
		return r.RenderNode(&buf, node, entering)
	})
	r.RenderFooter(&buf, ast)
	return buf.Bytes()
}
