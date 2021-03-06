package main

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/russross/blackfriday"
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

type RenderModifier func(b *bytes.Buffer, pos int, node *blackfriday.Node, step RenderStep) WalkStatus

func newSingle(nt blackfriday.NodeType) [3]blackfriday.NodeType {
	return [3]blackfriday.NodeType{-1, -1, nt}
}

var renderMods = map[[3]blackfriday.NodeType][]RenderModifier{
	newSingle(blackfriday.CodeBlock): []RenderModifier{codeBlock},
	newSingle(blackfriday.List):      []RenderModifier{taskList},
	[3]blackfriday.NodeType{blackfriday.Item, blackfriday.Paragraph, blackfriday.Text}: []RenderModifier{taskListItem},
}

var (
	cache   = make(map[[3]blackfriday.NodeType][]RenderModifier)
	cacheMu sync.RWMutex
)

func GetRenderMods(node *blackfriday.Node) []RenderModifier {
	k := newSingle(node.Type)

	if node.Parent != nil {
		k[1] = node.Parent.Type
		if node.Parent.Parent != nil {
			k[0] = node.Parent.Parent.Type
		}
	}

	cacheMu.RLock()
	mods, found := cache[k]
	cacheMu.RUnlock()
	if found {
		return mods
	}

	var kk [3]blackfriday.NodeType
	for i := 0; i < len(k); i++ {
		kk[i] = k[i]
	}

	for i := 0; i < len(kk); i++ {
		if kk[i] == -1 {
			continue
		}
		if m, found := renderMods[kk]; found {
			mods = append(mods, m...)
		}
		kk[i] = -1
	}

	cacheMu.Lock()
	cache[k] = mods
	cacheMu.Unlock()

	return mods
}

func codeBlock(b *bytes.Buffer, pos int, node *blackfriday.Node, step RenderStep) WalkStatus {
	fmt.Fprintf(b, `<code class="highlight"><pre style="background-color: orange;">%q</pre></code>`, node.Literal)
	return WalkStatusDone
}

func taskList(b *bytes.Buffer, pos int, node *blackfriday.Node, step RenderStep) WalkStatus {

	if step != RenderDone {
		// We need the list to be rendered before we can modify it.
		return WalkStatusContinue
	}

	if b.Len() > pos {
		list := b.Bytes()[pos:]

		if bytes.Contains(list, []byte("task-list-item")) {
			// Find the index of the first >, it might be 3 or 4 depending on whether
			// there is a new line at the start, but this is safer than just hardcoding it.
			closingBracketIndex := bytes.Index(list, []byte(">"))
			// Rewrite the buffer from the marker
			b.Truncate(pos)
			// Safely assuming closingBracketIndex won't be -1 since there is a list
			// May be either dl, ul or ol
			list := append(list[:closingBracketIndex], append([]byte(` class="task-list"`), list[closingBracketIndex:]...)...)
			b.Write(list)
		}
	}

	return WalkStatusContinue

}

func taskListItem(b *bytes.Buffer, pos int, node *blackfriday.Node, step RenderStep) WalkStatus {

	if step != RenderStart {
		return WalkStatusContinue
	}

	var newItem []byte

	switch {
	case bytes.HasPrefix(node.Literal, []byte("[ ] ")):
		newItem = append([]byte(`<label><input type="checkbox" disabled class="task-list-item">`), node.Literal[3:]...)
		newItem = append(newItem, []byte(`</label>`)...)
	case bytes.HasPrefix(node.Literal, []byte("[x] ")) || bytes.HasPrefix(node.Literal, []byte("[X] ")):
		newItem = append([]byte(`<label><input type="checkbox" checked disabled class="task-list-item">`), node.Literal[3:]...)
		newItem = append(newItem, []byte(`</label>`)...)
	}

	if newItem != nil {
		b.Write(newItem)
		return WalkStatusDone
	}

	return WalkStatusContinue

}
