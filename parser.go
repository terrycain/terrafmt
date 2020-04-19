package main

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// LexConfig Stolen from unexported function in Terraform FMT call
func LexConfig(src []byte) hclwrite.Tokens {
	mainTokens, _ := hclsyntax.LexConfig(src, "", hcl.Pos{Byte: 0, Line: 1, Column: 1})

	tokBuf := make([]hclwrite.Token, len(mainTokens))
	var lastByteOffset int
	for i, mainToken := range mainTokens {
		// Create a copy of the bytes so that we can mutate without
		// corrupting the original token stream.
		bytes := make([]byte, len(mainToken.Bytes))
		copy(bytes, mainToken.Bytes)

		tokBuf[i] = hclwrite.Token{
			Type:  mainToken.Type,
			Bytes: bytes,

			// We assume here that spaces are always ASCII spaces, since
			// that's what the scanner also assumes, and thus the number
			// of bytes skipped is also the number of space characters.
			SpacesBefore: mainToken.Range.Start.Byte - lastByteOffset,
		}

		lastByteOffset = mainToken.Range.End.Byte
	}

	// Now make a slice of pointers into the previous slice.
	ret := make(hclwrite.Tokens, len(tokBuf))
	for i := range ret {
		ret[i] = &tokBuf[i]
	}

	return ret
}

// Format Stolen from function in hclwrite
func Format(tokens hclwrite.Tokens, indentLength int, lineassignment, linecomment bool) {
	lines := linesForFormat(tokens)
	formatIndent(lines, indentLength)
	formatSpaces(lines)
	formatCells(lines, lineassignment, linecomment)
}

// FormatFile Wraps Format with file reading stuff
func FormatFile(filepath string, indentLength int, lineassignment, linecomment bool) (string, string) {
	file, _ := os.Open(filepath)
	defer file.Close()

	src, _ := ioutil.ReadAll(file)

	// result := hclwrite.Format(src)

	tokens := LexConfig(src)
	Format(tokens, indentLength, lineassignment, linecomment)
	buf := &bytes.Buffer{}
	tokens.WriteTo(buf)
	return string(src), string(buf.Bytes())
}

func formatIndent(lines []formatLine, indentLength int) {
	// Our methodology for indents is to take the input one line at a time
	// and count the bracketing delimiters on each line. If a line has a net
	// increase in open brackets, we increase the indent level by one and
	// remember how many new openers we had. If the line has a net _decrease_,
	// we'll compare it to the most recent number of openers and decrease the
	// dedent level by one each time we pass an indent level remembered
	// earlier.
	// The "indent stack" used here allows for us to recognize degenerate
	// input where brackets are not symmetrical within lines and avoid
	// pushing things too far left or right, creating confusion.

	// We'll start our indent stack at a reasonable capacity to minimize the
	// chance of us needing to grow it; 10 here means 10 levels of indent,
	// which should be more than enough for reasonable HCL uses.
	indents := make([]int, 0, 10)

	for i := range lines {
		line := &lines[i]
		if len(line.lead) == 0 {
			continue
		}

		if line.lead[0].Type == hclsyntax.TokenNewline {
			// Never place spaces before a newline
			line.lead[0].SpacesBefore = 0
			continue
		}

		netBrackets := 0
		for _, token := range line.lead {
			netBrackets += tokenBracketChange(token)
			if token.Type == hclsyntax.TokenOHeredoc {
				break
			}
		}

		for _, token := range line.assign {
			netBrackets += tokenBracketChange(token)
		}

		switch {
		case netBrackets > 0:
			line.lead[0].SpacesBefore = indentLength * len(indents)
			indents = append(indents, netBrackets)
		case netBrackets < 0:
			closed := -netBrackets
			for closed > 0 && len(indents) > 0 {
				switch {

				case closed > indents[len(indents)-1]:
					closed -= indents[len(indents)-1]
					indents = indents[:len(indents)-1]

				case closed < indents[len(indents)-1]:
					indents[len(indents)-1] -= closed
					closed = 0

				default:
					indents = indents[:len(indents)-1]
					closed = 0
				}
			}
			line.lead[0].SpacesBefore = indentLength * len(indents)
		default:
			line.lead[0].SpacesBefore = indentLength * len(indents)
		}
	}
}

func formatCells(lines []formatLine, lineassignment, linecomment bool) {

	chainStart := -1
	maxColumns := 0

	if lineassignment {
		// We'll deal with the "assign" cell first, since moving that will
		// also impact the "comment" cell.
		closeAssignChain := func(i int) {
			for _, chainLine := range lines[chainStart:i] {
				columns := chainLine.lead.Columns()
				spaces := (maxColumns - columns) + 1
				chainLine.assign[0].SpacesBefore = spaces
			}
			chainStart = -1
			maxColumns = 0
		}
		for i, line := range lines {
			if line.assign == nil {
				if chainStart != -1 {
					closeAssignChain(i)
				}
			} else {
				if chainStart == -1 {
					chainStart = i
				}
				columns := line.lead.Columns()
				if columns > maxColumns {
					maxColumns = columns
				}
			}
		}
		if chainStart != -1 {
			closeAssignChain(len(lines))
		}
	}

	if linecomment {
		// Now we'll deal with the comments
		closeCommentChain := func(i int) {
			for _, chainLine := range lines[chainStart:i] {
				columns := chainLine.lead.Columns() + chainLine.assign.Columns()
				spaces := (maxColumns - columns) + 1
				chainLine.comment[0].SpacesBefore = spaces
			}
			chainStart = -1
			maxColumns = 0
		}
		for i, line := range lines {
			if line.comment == nil {
				if chainStart != -1 {
					closeCommentChain(i)
				}
			} else {
				if chainStart == -1 {
					chainStart = i
				}
				columns := line.lead.Columns() + line.assign.Columns()
				if columns > maxColumns {
					maxColumns = columns
				}
			}
		}
		if chainStart != -1 {
			closeCommentChain(len(lines))
		}
	}
}
