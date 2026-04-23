package render

import "strings"

// emitBlock writes one parsed block in pretty form.
func emitBlock(out *strings.Builder, b block) {
	switch b.kind {
	case bkHeading:
		out.WriteString(b.text)
		out.WriteByte('\n')
	case bkSubtitle:
		out.WriteString(bodyIndent)
		out.WriteString(TokMutedOpen)
		out.WriteString(b.text)
		out.WriteString(TokMutedClose)
		out.WriteByte('\n')
	case bkParagraph:
		out.WriteString(bodyIndent)
		out.WriteString(HighlightQuotes(b.text))
		out.WriteByte('\n')
	case bkFence:
		for _, l := range b.lines {
			out.WriteString(bodyIndent)
			out.WriteString(l)
			out.WriteByte('\n')
		}
	case bkBlank:
		out.WriteByte('\n')
	}
}

// HighlightQuotes wraps every "double-quoted span" in cyan tokens
// (TokCyanOpen/TokCyanClose). Single quotes are left untouched (they're
// usually apostrophes). The output uses sentinel tokens — call
// HighlightQuotesANSI to get a string with real ANSI escape codes, or
// run the result through RenderANSI's swap layer.
//
// Exported so other CLI surfaces (e.g. the changelog pretty-printer) can
// share the exact same quote-rendering rule that the help-text renderer
// applies, keeping formatting consistent across commands.
func HighlightQuotes(s string) string {
	var b strings.Builder
	inQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '"' {
			b.WriteByte(c)

			continue
		}
		if !inQuote {
			b.WriteString(TokCyanOpen)
			b.WriteByte('"')
			inQuote = true

			continue
		}
		b.WriteByte('"')
		b.WriteString(TokCyanClose)
		inQuote = false
	}
	if inQuote { // unterminated quote: close the token defensively
		b.WriteString(TokCyanClose)
	}

	return b.String()
}
