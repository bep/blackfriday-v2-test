package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/russross/blackfriday"
	"github.com/sanity-io/litter"
)

type RenderStep int
type WalkStatus int

const (
	RenderStart RenderStep = iota + 1
	RenderEntered
	RenderDone
)

const (
	WalkStatusContinue WalkStatus = iota + 1
	WalkStatusDone
)

type RenderModifier func(b *bytes.Buffer, node *blackfriday.Node, step RenderStep) WalkStatus
type CreateRenderModifier func() RenderModifier

var renderMods = map[blackfriday.NodeType][]CreateRenderModifier{
	blackfriday.CodeBlock: []CreateRenderModifier{staticRenderModifier(codeBlock)},
	blackfriday.List:      []CreateRenderModifier{taskList},
	blackfriday.Text:      []CreateRenderModifier{taskListItem},
}

func GetRenderMods(nt blackfriday.NodeType) []RenderModifier {
	rmcs, found := renderMods[nt]
	if !found {
		return nil
	}

	rms := make([]RenderModifier, len(rmcs))
	for i, rmc := range rmcs {
		rms[i] = rmc()
	}

	return rms
}

func codeBlock(b *bytes.Buffer, node *blackfriday.Node, step RenderStep) WalkStatus {
	fmt.Fprintf(b, `<code class="highlight"><pre style="background-color: orange;">%q</pre></code>`, node.Literal)
	return WalkStatusDone
}

func staticRenderModifier(mod RenderModifier) CreateRenderModifier {
	return func() RenderModifier {
		return mod
	}
}

var taskListListItemReplacer = strings.NewReplacer(
	"<li>[x]", "asdf",
)

func taskList() RenderModifier {
	var marker int

	return func(b *bytes.Buffer, node *blackfriday.Node, step RenderStep) WalkStatus {
		if step == RenderStart {
			// Capture the start of this list so we can modify it later.
			marker = b.Len()
		}

		if step != RenderDone {
			// We need the list to be rendered before we can modify it.
			return WalkStatusContinue
		}

		if b.Len() > marker {
			list := b.Bytes()[marker:]

			//	fmt.Println(">>LIST:", string(list))
			if bytes.Contains(list, []byte("task-list-item")) {
				// Find the index of the first >, it might be 3 or 4 depending on whether
				// there is a new line at the start, but this is safer than just hardcoding it.
				closingBracketIndex := bytes.Index(list, []byte(">"))
				// Rewrite the buffer from the marker
				b.Truncate(marker)
				// Safely assuming closingBracketIndex won't be -1 since there is a list
				// May be either dl, ul or ol
				list := append(list[:closingBracketIndex], append([]byte(` class="task-list"`), list[closingBracketIndex:]...)...)
				b.Write(list)
			}
		}

		return WalkStatusContinue
	}
}

var _ = litter.Config

func taskListItem() RenderModifier {

	return func(b *bytes.Buffer, node *blackfriday.Node, step RenderStep) WalkStatus {
		if step != RenderStart || node.Parent == nil || node.Parent.Parent == nil || node.Parent.Parent.Type != blackfriday.Item {
			return WalkStatusContinue
		}

		//fmt.Printf(">> %v %s %v\n", node.FirstChild.Type, "FOO", string(node.FirstChild.Literal))

		item := node.Literal
		var newItem []byte

		switch {
		case bytes.HasPrefix(item, []byte("[ ] ")):
			newItem = append([]byte(`<label><input type="checkbox" disabled class="task-list-item">`), item[3:]...)
			newItem = append(newItem, []byte(`</label>`)...)
		case bytes.HasPrefix(item, []byte("[x] ")) || bytes.HasPrefix(item, []byte("[X] ")):
			newItem = append([]byte(`<label><input type="checkbox" checked disabled class="task-list-item">`), item[3:]...)
			newItem = append(newItem, []byte(`</label>`)...)
		}

		if newItem != nil {
			b.Write(newItem)
			return WalkStatusDone
		}

		return WalkStatusContinue

	}

}
