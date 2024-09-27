package marginblock

import (
	"gui/cstyle"
	"gui/element"
	"strings"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node) bool {
			return n.Style["margin-block"] != "" || n.Style["margin-block-start"] != "" || n.Style["margin-block-end"] != ""
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {

			writingMode := n.Style["writing-mode"]
			mb := parseMarginBlock(n.Style["margin-block"])

			if n.Style["margin-block-start"] != "" {
				mb[0] = n.Style["margin-block-start"]
			}
			if n.Style["margin-block-end"] != "" {
				mb[1] = n.Style["margin-block-end"]
			}

			if writingMode == "vertical-lr" {
				n.Style["margin-left"] = mb[0]
				n.Style["margin-right"] = mb[1]
			} else if writingMode == "vertical-rl" {
				// !ISSUE: This will not move everything over
				// + https://developer.mozilla.org/en-US/docs/Web/CSS/margin-block
				n.Style["margin-left"] = mb[1]
				n.Style["margin-right"] = mb[0]
			} else {
				n.Style["margin-top"] = mb[0]
				n.Style["margin-bottom"] = mb[1]
			}

			return n
		},
	}
}

func parseMarginBlock(s string) []string {
	split := strings.Split(s, " ")

	switch len(split) {
	case 2:
		return split
	case 1:
		return []string{split[0], split[0]}
	default:
		return []string{}
	}
}