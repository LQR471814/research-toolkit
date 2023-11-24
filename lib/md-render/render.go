package mdrender

import (
	"fmt"
	"strings"
)

type renderState struct {
	listDepth int
}

func Render(nodes []Node) string {
	text := ""
	for _, n := range nodes {
		text += render(n, renderState{}) + "\n\n"
	}
	return text
}

func render(node Node, state renderState) string {
	text := ""

	switch typedNode := node.(type) {
	case Header:
		if typedNode.Order > 6 {
			panic("header order must be <= 6")
		}
		for i := 0; i < typedNode.Order; i++ {
			text += "#"
		}
		text += " "
		text += render(typedNode.Content, state)
	case Paragraph:
		for i, e := range typedNode.Elements {
			if i > 0 {
				text += " "
			}
			text += render(e, state)
			if i >= len(typedNode.Elements) {
				text += " "
			}
		}
	case Link:
		inner := ""

		switch typedNode.Title.(type) {
		case Link:
			inner = render(typedNode.Title.(Link).Title, state)
		default:
			inner = render(typedNode.Title, state)
		}

		if typedNode.Image {
			text += "!"
		}
		text += fmt.Sprintf("[%s](%s)", inner, typedNode.URL)
	case LineBreak:
		text = "<br>"
	case HorizontalRule:
		text = "<hr>"
	case InlineCode:
		text = fmt.Sprintf("`%s`", strings.ReplaceAll(typedNode.Content, "`", "\\`"))
	case BlockCode:
		text = fmt.Sprintf(
			"```%s\n%s\n```", typedNode.Language,
			strings.ReplaceAll(
				strings.TrimRight(typedNode.Content, " \n\t"),
				"`", "\\`",
			),
		)
	case PlainText:
		text = strings.Trim(typedNode.Content, " \n\t")
	case DecoratedText:
		inside := render(typedNode.Content, state)
		switch typedNode.Type {
		case DECOR_BOLD:
			text = fmt.Sprintf("**%s**", strings.ReplaceAll(inside, "*", "_"))
		case DECOR_ITALIC:
			text = fmt.Sprintf("*%s*", strings.ReplaceAll(inside, "*", "_"))
		case DECOR_UNDERLINE:
			text = fmt.Sprintf("<ins>%s</ins>", inside)
		}
	case List:
		indent := ""
		for i := 0; i < state.listDepth; i++ {
			indent += "   "
		}

		prefix := "-"
		if typedNode.Type == LIST_ORDERED {
			prefix = "1."
		}

		for i, item := range typedNode.Items {
			switch item.(type) {
			case List:
				text += render(item, renderState{
					listDepth: state.listDepth + 1,
				})
			default:
				text += fmt.Sprintf(
					"%s %s %s",
					indent, prefix, render(item, renderState{
						listDepth: state.listDepth + 1,
					}),
				)
			}
			if i < len(typedNode.Items)-1 {
				text += "\n"
			}
		}
	}

	return text
}
