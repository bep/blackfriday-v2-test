package main

import (
	"bytes"
	"testing"

	"github.com/russross/blackfriday"
)

func TestMain(t *testing.T) {
	var b bytes.Buffer
	run(md, &b, nil)

	if b.String() != `<h2>Heading 1</h2>

<p>Some <strong>text</strong>.</p>

<h2>Heading 2</h2>

<p>This is a footnote.<sup class="footnote-ref" id="fnref:fn1"><a href="#fn:fn1">1</a></sup></p>

<p>A fenced code block:</p>
<code class="highlight"><pre style="background-color: orange;">"type RenderModifier interface {\n}\n"</pre></code>
<h3>Heading 2-1</h3>

<p>Task list:</p>

<ul class="task-list">
<li><label><input type="checkbox" checked disabled class="task-list-item"> Finish my changes</label></li>
<li><label><input type="checkbox" disabled class="task-list-item"> Push my commits to GitHub</label></li>
<li><label><input type="checkbox" disabled class="task-list-item"> Open a pull request</label></li>
</ul>

<p>Some more text.<sup class="footnote-ref" id="fnref:fn2"><a href="#fn:fn2">2</a></sup></p>

<div class="footnotes">

<hr />

<ol>
<li id="fn:fn1">the footnote text.</li>

<li id="fn:fn2">the footnote 2 text.</li>
</ol>

</div>
` {
		t.Fatal(b.String())
	}
}

func BenchmarkGetModifiers(b *testing.B) {
	var nodes []*blackfriday.Node
	var buf bytes.Buffer
	run(md, &buf, func(node *blackfriday.Node) {
		nodes = append(nodes, node)
	})

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, node := range nodes {
			GetRenderMods(node)
		}
	}
}
