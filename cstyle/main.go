package cstyle

// package aui/goldie
// https://pkg.go.dev/automated.sh/goldie
// https://pkg.go.dev/automated.sh/aui

import (
	"fmt"
	"gui/color"
	"gui/font"
	"gui/parser"
	"gui/utils"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/go-shiori/dom"
	"golang.org/x/net/html"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type CSS struct {
	Width       float32
	Height      float32
	StyleSheets []map[string]map[string]string
}

type Mapped struct {
	Document *Node
	StyleMap map[string]map[string]string
	Render   []Node
}

type Node struct {
	Node     *html.Node
	Parent   *Node
	Children []Node
	Styles   map[string]string
	Id       string
	X        float32
	Y        float32
	Width    float32
	Height   float32
	Margin   Margin
	Padding  Padding
	Border   Border
	EM       float32
	Text     font.Text
	Colors   color.Colors
}

type Margin struct {
	Top    float32
	Right  float32
	Bottom float32
	Left   float32
}

type Padding struct {
	Top    float32
	Right  float32
	Bottom float32
	Left   float32
}

type Border struct {
	Top    BorderSide
	Right  BorderSide
	Bottom BorderSide
	Left   BorderSide
}

type BorderSide struct {
	Width  float32
	Style  string
	Color  string
	Radius float32
}

func (c *CSS) StyleSheet(path string) {
	// Parse the CSS file
	dat, err := os.ReadFile(path)
	check(err)
	styles := parser.ParseCSS(string(dat))

	c.StyleSheets = append(c.StyleSheets, styles)
}

func (c *CSS) StyleTag(css string) {
	styles := parser.ParseCSS(css)
	c.StyleSheets = append(c.StyleSheets, styles)
}

func (c *CSS) Map(doc *html.Node) Mapped {
	styleMap := make(map[string]map[string]string)
	for a := 0; a < len(c.StyleSheets); a++ {
		for key, styles := range c.StyleSheets[a] {
			matching := dom.QuerySelectorAll(doc, key)
			for _, v := range matching {
				if v.Type == html.ElementNode {
					id := dom.GetAttribute(v, "DOMNODEID")
					if len(id) == 0 {
						id = dom.TagName(v) + fmt.Sprint(rand.Int63())
						dom.SetAttribute(v, "DOMNODEID", id)
					}

					if styleMap[id] == nil {
						styleMap[id] = styles
					} else {
						styleMap[id] = utils.Merge(styleMap[id], styles)
					}
				}
			}
		}
	}

	// Inherit CSS styles from parent
	println("inherit")
	inherit(doc, styleMap)
	println(c.Width, c.Height)
	fId := dom.GetAttribute(doc.FirstChild, "DOMNODEID")
	node := Node{
		Node: doc.FirstChild,
		Parent: &Node{
			Id:     "ROOT",
			X:      0,
			Y:      0,
			Width:  c.Width,
			Height: c.Height,
			EM:     16,
			Styles: map[string]string{
				"width":  strconv.FormatFloat(float64(c.Width), 'f', -1, 32) + "px",
				"height": strconv.FormatFloat(float64(c.Height), 'f', -1, 32) + "px",
			},
		},
		Id:     fId,
		X:      0,
		Y:      0,
		Width:  c.Width,
		Height: c.Height,
		Styles: styleMap[fId],
	}
	fmt.Printf("%#v\n", node.Id)
	initNodes(&node, styleMap)

	node = ComputeNodeStyle(node)

	Print(&node, 0)

	renderLine := flatten(&node)

	d := Mapped{
		Document: &node,
		StyleMap: styleMap,
		Render:   renderLine,
	}
	return d
}

func ComputeNodeStyle(n Node) Node {
	// Need to make a function that builds a Node tree from a *html.Node
	// Kind of a chicken and the egg problem.. I need to have these styles to make the Node
	// Maybe make this function a function like inherit (circular i know lol) but just go ahead
	// and compute all styles so it can be passed directly to the renderer
	// but
	// its still the same issue. I need the Node tree to make compute the styles
	// unless the tree I create basic child nodes that have the parent node mapped and computed???
	// anything above should already be computed so it should work

	styleMap := n.Styles

	if styleMap["display"] == "none" {
		n.X = 0
		n.Y = 0
		n.Width = 0
		n.Height = 0
		return n
	}

	width, height := n.Width, n.Height
	x, y := n.Parent.X, n.Parent.Y

	var top, left, right, bottom bool = false, false, false, false

	if styleMap["position"] == "absolute" {
		println("ABSOLUTE")

		base := GetPositionOffsetNode(&n)

		if styleMap["top"] != "" {
			v, _ := utils.ConvertToPixels(styleMap["top"], float32(n.EM), n.Parent.Width)
			y = v + base.Y
			top = true
		}
		if styleMap["left"] != "" {
			v, _ := utils.ConvertToPixels(styleMap["left"], float32(n.EM), n.Parent.Width)
			x = v + base.X
			left = true
		}
		if styleMap["right"] != "" {
			v, _ := utils.ConvertToPixels(styleMap["right"], float32(n.EM), n.Parent.Width)
			x = (base.Width - width) - v
			right = true
		}
		if styleMap["bottom"] != "" {
			v, _ := utils.ConvertToPixels(styleMap["bottom"], float32(n.EM), n.Parent.Width)
			y = (base.Height - height) - v
			bottom = true
		}
	} else {
		for i, v := range n.Parent.Children {
			if v.Id == n.Id {
				if i-1 > 0 {
					sibling := n.Parent.Children[i-1]
					if styleMap["display"] == "inline" {
						if sibling.Styles["display"] == "inline" {
							y = sibling.Y
						} else {
							y = sibling.Y + sibling.Height
						}
					} else {
						y = sibling.Y + sibling.Height
					}
				}
				break
			}

		}

	}

	// Display modes need to be calculated here

	relPos := !top && !left && !right && !bottom

	if left || relPos {
		x += n.Margin.Left
	}
	if top || relPos {
		y += n.Margin.Top
	}
	if right {
		x -= n.Margin.Right
	}
	if bottom {
		y -= n.Margin.Bottom
	}

	// Display

	if styleMap["display"] == "block" {
		// If the element is display block and the width is unset then make it 100%
		if styleMap["width"] == "" {
			width, _ = utils.ConvertToPixels("100%", n.EM, n.Parent.Width)
			width -= n.Margin.Right + n.Margin.Left
		}
	}

	// The element is empty, need to calculate the height of the element
	// the only case where this would happen is a text node

	// NOTE: other elements can have text...
	if len(n.Children) == 0 {
		text := dom.InnerText(n.Node)
		// Confirm text exists
		if styleMap["width"] == "" {
			if len(text) > 0 {
				height = n.EM

				bold, italic := false, false

				if styleMap["font-weight"] == "bold" {
					bold = true
				}

				if styleMap["font-style"] == "italic" {
					italic = true
				}

				f, _ := font.LoadFont(styleMap["font-family"], int(n.EM), bold, italic)

				width = utils.Max(width, float32(font.MeasureText(f, text)))
				c, _ := color.Font(styleMap)
				n.Text = font.Text{
					Text:  text,
					Font:  f,
					Color: c,
				}
				n.Text.Render()
			}
		}
	}

	if styleMap["display"] == "inline" {
		copyOfX := x
		for _, v := range n.Parent.Children {
			if v.Id == n.Id {
				break
			} else if v.Styles["display"] == "inline" {
				x += v.Width
			} else {
				x = copyOfX
			}
		}
	}

	n.X = x
	n.Y = y
	n.Width = width
	n.Height = height

	// Call children here

	for i, v := range n.Children {
		v.Parent = &n
		n.Children[i] = ComputeNodeStyle(v)
		if styleMap["height"] == "" {
			if n.Children[i].Styles["position"] != "absolute" {
				n.Height += n.Children[i].Height
				n.Height += n.Children[i].Margin.Top
				n.Height += n.Children[i].Margin.Bottom
			}

		}
	}

	return n
}

var inheritedProps = []string{
	"color",
	"cursor",
	"font",
	"font-family",
	"font-size",
	"font-style",
	"font-weight",
	"letter-spacing",
	"line-height",
	"text-align",
	"text-indent",
	"text-justify",
	"text-shadow",
	"text-transform",
	"visibility",
	"word-spacing",
	"display",
}

func inherit(n *html.Node, styleMap map[string]map[string]string) {
	if n.Type == html.ElementNode {
		if dom.TagName(n) != "notaspan" || len(dom.InnerText(n)) != 0 {
			id := dom.GetAttribute(n, "DOMNODEID")
			if len(id) == 0 {
				id = dom.TagName(n) + fmt.Sprint(rand.Int63())
				dom.SetAttribute(n, "DOMNODEID", id)
			}
			pId := dom.GetAttribute(n.Parent, "DOMNODEID")
			if len(pId) > 0 {
				if styleMap[id] == nil {
					styleMap[id] = make(map[string]string)
				}
				if styleMap[pId] == nil {
					styleMap[pId] = make(map[string]string)
				}

				for _, v := range inheritedProps {
					if styleMap[id][v] == "" && styleMap[pId][v] != "" {
						styleMap[id][v] = styleMap[pId][v]
					}
				}
			}
			utils.SetMP(id, styleMap)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		inherit(c, styleMap)
	}
}

func initNodes(n *Node, styleMap map[string]map[string]string) {
	inline := parser.ParseStyleAttribute(dom.GetAttribute(n.Node, "style") + ";")
	styleMap[n.Id] = utils.Merge(styleMap[n.Id], inline)
	n.Styles = styleMap[n.Id]

	border, err := CompleteBorder(n.Styles)
	if err == nil {
		n.Border = border
	}

	fs, _ := utils.ConvertToPixels(n.Styles["font-size"], n.Parent.EM, n.Parent.Width)
	n.EM = fs

	mt, _ := utils.ConvertToPixels(n.Styles["margin-top"], n.EM, n.Parent.Width)
	mr, _ := utils.ConvertToPixels(n.Styles["margin-right"], n.EM, n.Parent.Width)
	mb, _ := utils.ConvertToPixels(n.Styles["margin-bottom"], n.EM, n.Parent.Width)
	ml, _ := utils.ConvertToPixels(n.Styles["margin-left"], n.EM, n.Parent.Width)
	n.Margin = Margin{
		Top:    mt,
		Right:  mr,
		Bottom: mb,
		Left:   ml,
	}

	pt, _ := utils.ConvertToPixels(n.Styles["padding-top"], n.EM, n.Parent.Width)
	pr, _ := utils.ConvertToPixels(n.Styles["padding-right"], n.EM, n.Parent.Width)
	pb, _ := utils.ConvertToPixels(n.Styles["padding-bottom"], n.EM, n.Parent.Width)
	pl, _ := utils.ConvertToPixels(n.Styles["padding-left"], n.EM, n.Parent.Width)
	n.Padding = Padding{
		Top:    pt,
		Right:  pr,
		Bottom: pb,
		Left:   pl,
	}

	width, _ := utils.ConvertToPixels(n.Styles["width"], n.EM, n.Parent.Width)
	height, _ := utils.ConvertToPixels(n.Styles["height"], n.EM, n.Parent.Height)

	n.Width = width
	n.Height = height

	n.Colors = color.Parse(n.Styles)

	for _, c := range dom.ChildNodes(n.Node) {
		if c.Type == html.ElementNode {
			id := dom.GetAttribute(c, "DOMNODEID")
			node := Node{
				Node:   c,
				Parent: n,
				Id:     id,
				Styles: styleMap[id],
			}
			initNodes(&node, styleMap)
			n.Children = append(n.Children, node)
		}
	}
}

func GetPositionOffsetNode(n *Node) *Node {
	pos := n.Styles["position"]

	if pos == "relative" {
		return n
	} else {
		if n.Parent.Node != nil {
			return GetPositionOffsetNode(n.Parent)
		} else {
			return nil
		}
	}
}

func parseBorderShorthand(borderShorthand string) (BorderSide, error) {
	// Split the shorthand into components
	borderComponents := strings.Fields(borderShorthand)

	// Ensure there are at least 1 component (width or style or color)
	if len(borderComponents) >= 1 {
		width := "0px" // Default width
		style := "solid"
		color := "#000000" // Default color

		// Extract numeric part for width
		if strings.ContainsAny(borderComponents[0], "0123456789") {
			width = strings.TrimRightFunc(borderComponents[0], func(r rune) bool {
				return !strings.ContainsRune("0123456789.", r)
			})
		}

		// Extract style and color if available
		if len(borderComponents) >= 2 {
			style = borderComponents[1]
		}
		if len(borderComponents) >= 3 {
			color = borderComponents[2]
		}

		// Parse width to float
		widthFloat, err := strconv.ParseFloat(width, 32)
		if err != nil {
			return BorderSide{}, fmt.Errorf("failed to parse border width: %v", err)
		}

		return BorderSide{
			Width:  float32(widthFloat),
			Style:  style,
			Color:  color,
			Radius: 0.0, // Default radius
		}, nil
	}

	return BorderSide{}, fmt.Errorf("invalid border shorthand format")
}

func CompleteBorder(cssProperties map[string]string) (Border, error) {
	borderShorthand, hasBorder := cssProperties["border"]

	var border Border

	if hasBorder {
		side, err := parseBorderShorthand(borderShorthand)
		if err != nil {
			return Border{}, err
		}

		border.Top = side
		border.Right = side
		border.Bottom = side
		border.Left = side

		// Remove the shorthand border property from the map
		delete(cssProperties, "border")
	}

	// Map individual border properties
	borderProperties := map[string]string{
		"top":    "border-top",
		"right":  "border-right",
		"bottom": "border-bottom",
		"left":   "border-left",
	}

	for side, property := range borderProperties {
		width := cssProperties[property+"-width"]
		style := cssProperties[property+"-style"]
		color := cssProperties[property+"-color"]

		if width == "" {
			width = "1px" // Default width
		}

		if style == "" {
			style = "solid" // Default style
		}

		if color == "" {
			color = "#000000" // Default color
		}

		radius := cssProperties[property+"-radius"]
		radiusValue, err := strconv.ParseFloat(radius, 32)
		if err != nil || radius == "" {
			radiusValue = 0.0 // Default radius
		}

		borderSide, err := parseBorderShorthand(fmt.Sprintf("%s %s %s", width, style, color))
		if err != nil {
			return Border{}, err
		}

		borderSide.Radius = float32(radiusValue)

		switch side {
		case "top":
			border.Top = borderSide
		case "right":
			border.Right = borderSide
		case "bottom":
			border.Bottom = borderSide
		case "left":
			border.Left = borderSide
		}
	}

	return border, nil
}

func Print(n *Node, indent int) {
	pre := strings.Repeat("\t", indent)
	fmt.Printf(pre+"%s\n", n.Id)
	fmt.Printf(pre+"-- Parent: %d\n", n.Parent.Id)
	fmt.Printf(pre+"\t-- Width: %f\n", n.Parent.Width)
	fmt.Printf(pre+"\t-- Height: %f\n", n.Parent.Height)
	fmt.Printf(pre + "-- Colors:\n")
	fmt.Printf(pre+"\t-- Font: %f\n", n.Colors.Font)
	fmt.Printf(pre+"\t-- Background: %f\n", n.Colors.Background)
	fmt.Printf(pre+"-- Children: %d\n", len(n.Children))
	fmt.Printf(pre+"-- EM: %f\n", n.EM)
	fmt.Printf(pre+"-- X: %f\n", n.X)
	fmt.Printf(pre+"-- Y: %f\n", n.Y)
	fmt.Printf(pre+"-- Width: %f\n", n.Width)
	fmt.Printf(pre+"-- Height: %f\n", n.Height)
	fmt.Printf(pre+"-- Styles: %#v\n", n.Styles)

	for _, v := range n.Children {
		Print(&v, indent+1)
	}
}

func InheritProp(n *Node, prop string) string {
	value := n.Styles[prop]

	if value != "" {
		return value
	} else {
		if n.Parent.Node != nil {
			v := InheritProp(n.Parent, prop)
			return v
		} else {
			return ""
		}
	}
}

func InheritPropWithNode(n *Node, prop string) (string, *Node) {
	value := n.Styles[prop]

	if value != "" {
		return value, n
	} else {
		if n.Parent != nil {
			v, p := InheritPropWithNode(n.Parent, prop)
			return v, p
		} else {
			return "", &Node{}
		}
	}
}

func flatten(n *Node) []Node {
	var nodes []Node
	nodes = append(nodes, *n)

	children := n.Children
	if len(children) > 0 {
		for _, ch := range children {
			chNodes := flatten(&ch)
			nodes = append(nodes, chNodes...)
		}
	}
	return nodes
}
