package converter

import (
	"bytes"

	katex "github.com/FurqanSoftware/goldmark-katex"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// kindMathBlock identifies a display-math ($$...$$) node.
var kindMathBlock = ast.NewNodeKind("MathBlock")

// mathBlock is a block-level node holding the raw LaTeX of a display equation.
type mathBlock struct {
	ast.BaseBlock
}

func (n *mathBlock) Kind() ast.NodeKind { return kindMathBlock }

// IsRaw reports that the body is raw LaTeX, so goldmark does not run inline
// Markdown parsing over it (which would turn "_" into emphasis, drop "\,", etc.).
func (n *mathBlock) IsRaw() bool { return true }

func (n *mathBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// mathBlockParser parses display math ($$...$$) as a block-level element.
//
// goldmark-katex only ships an *inline* parser for "$$". That means a
// multi-line equation is first chopped up by block parsing: a line such as
// "}_{A \in \mathbb{R}^{r \times d}}" followed by "=" becomes a Setext <h1>,
// "\\" row separators collide with hard-wrap <br/>, and the inline scan also
// gives up after 20 lines. Parsing the block at the block level instead keeps
// its contents completely opaque to Markdown, so any size of matrix renders.
type mathBlockParser struct{}

func (b *mathBlockParser) Trigger() []byte { return []byte{'$'} }

func (b *mathBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := util.TrimLeftSpaceLength(line)

	// The line must open with "$$".
	if pos+1 >= len(line) || line[pos] != '$' || line[pos+1] != '$' {
		return nil, parser.NoChildren
	}

	node := &mathBlock{}
	rest := line[pos+2:]

	// Closing "$$" on the same line means a single-line display block.
	if idx := bytes.Index(rest, []byte("$$")); idx >= 0 {
		start := segment.Start + pos + 2
		node.Lines().Append(text.NewSegment(start, start+idx))
		reader.AdvanceToEOL()
		return node, parser.Close
	}

	reader.AdvanceToEOL()
	return node, parser.NoChildren
}

func (b *mathBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	pos := util.TrimLeftSpaceLength(line)

	// A line that opens with "$$" closes the block.
	if pos+1 < len(line) && line[pos] == '$' && line[pos+1] == '$' {
		reader.AdvanceToEOL()
		return parser.Close
	}

	node.Lines().Append(segment)
	reader.AdvanceToEOL()
	return parser.Continue | parser.NoChildren
}

func (b *mathBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (b *mathBlockParser) CanInterruptParagraph() bool { return true }

func (b *mathBlockParser) CanAcceptIndentedLine() bool { return false }

// mathBlockRenderer renders mathBlock nodes to HTML via KaTeX, reusing the
// library's exported Render so display math matches the inline math styling.
type mathBlockRenderer struct{}

func (r *mathBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindMathBlock, r.render)
}

func (r *mathBlockRenderer) render(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	var eq bytes.Buffer
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		eq.Write(seg.Value(source))
	}

	w.WriteString("<div>")
	if err := katex.Render(w, eq.Bytes(), true, false); err != nil {
		return ast.WalkStop, err
	}
	w.WriteString("</div>")
	return ast.WalkContinue, nil
}
