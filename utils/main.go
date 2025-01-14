package utils

import (
	"bytes"
	"fmt"
	"gui/element"
	ic "image/color"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func GetXY(n *element.Node, state *map[string]element.State) (float32, float32) {
	s := *state
	// self := s[n.Properties.Id]

	offsetX := float32(0)
	offsetY := float32(0)

	if n.Parent != nil {
		parent := s[n.Parent.Properties.Id]
		// x, y := GetXY(n.Parent, state)
		offsetX += parent.Border.Left.Width + parent.Padding.Left
		offsetY += parent.Border.Top.Width + parent.Padding.Top
	}

	return offsetX, offsetY
}

type WidthHeight struct {
	Width  float32
	Height float32
}

func GetWH(n element.Node, state *map[string]element.State) WidthHeight {
	s := *state
	self := s[n.Properties.Id]
	var parent element.State

	if n.Style == nil {
		n.Style = make(map[string]string)
	}

	fs := self.EM

	var pwh WidthHeight
	if n.Parent != nil {
		parent = s[n.Parent.Properties.Id]
		// pwh = GetWH(*n.Parent, state)
		pwh = WidthHeight{
			Width:  parent.Width,
			Height: parent.Height,
		}
	} else {
		pwh = WidthHeight{}
		if width, exists := n.Style["width"]; exists {
			if f, err := strconv.ParseFloat(strings.TrimSuffix(width, "px"), 32); err == nil {
				pwh.Width = float32(f)
			}
		}
		if height, exists := n.Style["height"]; exists {
			if f, err := strconv.ParseFloat(strings.TrimSuffix(height, "px"), 32); err == nil {
				pwh.Height = float32(f)
			}
		}
	}

	wStyle := n.Style["width"]

	if wStyle == "" && n.Style["display"] != "inline" {
		wStyle = "100%"
	}

	width := ConvertToPixels(wStyle, fs, pwh.Width)
	height := ConvertToPixels(n.Style["height"], fs, pwh.Height)

	if minWidth, exists := n.Style["min-width"]; exists {
		width = Max(width, ConvertToPixels(minWidth, fs, pwh.Width))
	}
	if maxWidth, exists := n.Style["max-width"]; exists {
		width = Min(width, ConvertToPixels(maxWidth, fs, pwh.Width))
	}
	if minHeight, exists := n.Style["min-height"]; exists {
		height = Max(height, ConvertToPixels(minHeight, fs, pwh.Height))
	}
	if maxHeight, exists := n.Style["max-height"]; exists {
		height = Min(height, ConvertToPixels(maxHeight, fs, pwh.Height))
	}

	wh := WidthHeight{
		Width:  width,
		Height: height,
	}

	if n.Parent != nil {
		wh.Width += self.Padding.Left + self.Padding.Right
		wh.Height += self.Padding.Top + self.Padding.Bottom
	}

	if wStyle == "100%" {
		wh.Width -= (self.Margin.Right + self.Margin.Left + self.Border.Left.Width + self.Border.Right.Width + parent.Padding.Left + parent.Padding.Right + self.Padding.Left + self.Padding.Right)
	}

	if n.Style["height"] == "100%" {
		if n.Style["position"] == "absolute" {
			wh.Height -= (self.Margin.Top + self.Margin.Bottom)
		} else {
			wh.Height -= (self.Margin.Top + self.Margin.Bottom + parent.Padding.Top + parent.Padding.Bottom)
		}
	}

	return wh
}

func GetMP(n element.Node, wh WidthHeight, state *map[string]element.State, t string) element.MarginPadding {
	s := *state
	self := s[n.Properties.Id]
	fs := self.EM
	m := element.MarginPadding{}

	// Cache style properties
	style := n.Style
	leftKey, rightKey, topKey, bottomKey := t+"-left", t+"-right", t+"-top", t+"-bottom"

	leftStyle := style[leftKey]
	rightStyle := style[rightKey]
	topStyle := style[topKey]
	bottomStyle := style[bottomKey]

	if style[t] != "" {
		left, right, top, bottom := convertMarginToIndividualProperties(style[t])
		if leftStyle == "" {
			leftStyle = left
		}
		if rightStyle == "" {
			rightStyle = right
		}
		if topStyle == "" {
			topStyle = top
		}
		if bottomStyle == "" {
			bottomStyle = bottom
		}
	}

	// Convert left and right properties
	if leftStyle != "" || rightStyle != "" {
		m.Left = ConvertToPixels(leftStyle, fs, wh.Width)
		m.Right = ConvertToPixels(rightStyle, fs, wh.Width)
	}

	// Convert top and bottom properties
	if topStyle != "" || bottomStyle != "" {
		m.Top = ConvertToPixels(topStyle, fs, wh.Height)
		m.Bottom = ConvertToPixels(bottomStyle, fs, wh.Height)
	}

	if t == "margin" {
		siblingMargin := float32(0)
		firstChild := false
		// Margin Collapse
		if n.Parent != nil && ParentStyleProp(n.Parent, "display", func(prop string) bool {
			return prop == "flex"
		}) {
			sibIndex := -1
			for i, v := range n.Parent.Children {
				if v.Properties.Id == n.Properties.Id {
					sibIndex = i - 1

					break
				}
			}
			if sibIndex > -1 {
				sib := s[n.Parent.Children[sibIndex].Properties.Id]
				siblingMargin = sib.Margin.Bottom
			}
		}

		// Handle top margin collapse
		for i, v := range n.Parent.Children {
			if v.Properties.Id == n.Properties.Id {
				if i == 0 {
					firstChild = true
				}
				break
			}
		}
		if firstChild {
			parent := s[n.Parent.Properties.Id]
			if parent.Margin.Top < m.Top {
				parent.Margin.Top = m.Top
				(*state)[n.Parent.Properties.Id] = parent
			}
			m.Top = 0
		} else {
			if m.Top != 0 {
				if m.Top < 0 {
					m.Top += siblingMargin
				} else {
					m.Top = Max(m.Top-siblingMargin, 0)
				}
			}
		}

		// Handle auto margins
		if style["margin"] == "auto" && leftStyle == "" && rightStyle == "" {
			// pwh := GetWH(*n.Parent, state)
			parent := s[n.Parent.Properties.Id]
			pwh := WidthHeight{
				Width: parent.Width,
			}
			m.Left = Max((pwh.Width-wh.Width)/2, 0)
			m.Right = m.Left
		}
	}

	return m
}

func convertMarginToIndividualProperties(margin string) (string, string, string, string) {
	parts := strings.Fields(margin)
	switch len(parts) {
	case 1:
		return parts[0], parts[0], parts[0], parts[0]
	case 2:
		return parts[0], parts[1], parts[0], parts[1]
	case 3:
		return parts[0], parts[1], parts[2], parts[1]
	case 4:
		return parts[0], parts[1], parts[2], parts[3]
	}
	return "0px", "0px", "0px", "0px"
}

var unitFactors = map[string]float64{
	"px": 1,
	"em": -1, // special handling
	"pt": 1.33,
	"pc": 16.89,
	"%":  -1, // special handling
	"vw": -1, // special handling
	"vh": -1, // special handling
	"cm": 37.79527559,
	"in": 96,
}

// ConvertToPixels converts a CSS measurement to pixels.
func ConvertToPixels(value string, em, max float32) float32 {
	// Quick check for predefined units
	switch value {
	case "thick":
		return 5
	case "medium":
		return 3
	case "thin":
		return 1
	}

	// Handle calculation expression
	if len(value) > 5 && value[:5] == "calc(" {
		return evaluateCalcExpression(value[5:len(value)-1], em, max)
	}

	for unit, factor := range unitFactors {
		if strings.HasSuffix(value, unit) {
			cutStr := strings.TrimSuffix(value, unit)
			numericValue, err := strconv.ParseFloat(cutStr, 64)
			if err != nil {
				return 0
			}

			// Handle special units like "em", "%" etc.
			if factor == -1 {
				switch unit {
				case "em":
					return float32(numericValue) * em
				case "%", "vw", "vh":
					return float32(numericValue) * (max / 100)
				}
			}
			return float32(numericValue) * float32(factor)
		}
	}

	// Default return if no match
	return 0
}

// evaluateCalcExpression recursively evaluates 'calc()' expressions
func evaluateCalcExpression(expression string, em, max float32) float32 {
	terms := strings.FieldsFunc(expression, func(c rune) bool {
		return c == '+' || c == '-' || c == '*' || c == '/'
	})

	operators := strings.FieldsFunc(expression, func(c rune) bool {
		return c != '+' && c != '-' && c != '*' && c != '/'
	})

	var result float32

	for i, term := range terms {
		value := ConvertToPixels(strings.TrimSpace(term), em, max)

		if i > 0 {
			switch operators[i-1] {
			case "+":
				result += value
			case "-":
				result -= value
			case "*":
				result *= value
			case "/":
				if value != 0 {
					result /= value
				} else {
					return 0
				}
			}
		} else {
			result = value
		}
	}

	return result
}

func Merge(m1, m2 map[string]string) map[string]string {
	// Create a new map and copy m1 into it
	result := make(map[string]string)
	for k, v := range m1 {
		result[k] = v
	}

	// Merge m2 into the new map
	for k, v := range m2 {
		result[k] = v
	}

	return result
}

func Max(a, b float32) float32 {
	if a > b {
		return a
	} else {
		return b
	}
}

func Min(a, b float32) float32 {
	if a < b {
		return a
	} else {
		return b
	}
}

// getStructField uses reflection to get the value of a struct field by name
func GetStructField(data interface{}, fieldName string) interface{} {
	val := reflect.ValueOf(data)

	// Make sure we have a pointer to a struct
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return nil
	}

	// Get the struct field by name
	field := val.Elem().FieldByName(fieldName)

	// Check if the field exists
	if !field.IsValid() {
		return nil
	}

	return field.Interface()
}

func SetStructFieldValue(data interface{}, fieldName string, newValue interface{}) {
	val := reflect.ValueOf(data)

	// Make sure we have a pointer to a struct
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return
	}

	// Get the struct field by name
	field := val.Elem().FieldByName(fieldName)

	// Check if the field exists
	if !field.IsValid() {
		return
	}

	// Check if the new value type is assignable to the field type
	if !reflect.ValueOf(newValue).Type().AssignableTo(field.Type()) {
		return
	}

	// Set the new value
	field.Set(reflect.ValueOf(newValue))

}

func GetInnerText(n *html.Node) string {
	var result strings.Builder

	var getText func(*html.Node)
	getText = func(n *html.Node) {
		// Skip processing if the node is a head tag
		if n.Type == html.ElementNode && n.Data == "head" {
			return
		}

		// If it's a text node, append its content
		if n.Type == html.TextNode {
			result.WriteString(n.Data)
		}

		// Traverse child nodes recursively
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			getText(c)
		}
	}

	getText(n)

	return result.String()
}

func GetPositionOffsetNode(n *element.Node) *element.Node {
	pos := n.Style["position"]

	if pos == "relative" || pos == "absolute" {
		return n
	} else {
		if n.Parent.TagName != "ROOT" {
			if n.Parent.Style != nil {
				return GetPositionOffsetNode(n.Parent)
			} else {
				return nil
			}
		} else {
			return n.Parent
		}
	}
}

func IsParent(n element.Node, name string) bool {
	if n.Parent == nil {
		return false
	}
	if n.Parent.TagName != "ROOT" {
		if n.Parent.TagName == name {
			return true
		} else {
			return IsParent(*n.Parent, name)
		}
	} else {
		return false
	}
}

func ChildrenHaveText(n *element.Node) bool {
	for _, child := range n.Children {
		if len(strings.TrimSpace(child.InnerText)) != 0 {
			return true
		}
		// Recursively check if any child nodes have text
		if ChildrenHaveText(child) {
			return true
		}
	}
	return false
}

func NodeToHTML(node *element.Node) (string, string) {
	// if node.TagName == "notaspan" {
	// 	return node.InnerText + " ", ""
	// }

	var buffer bytes.Buffer
	buffer.WriteString("<" + node.TagName)

	if node.ContentEditable {
		buffer.WriteString(" contentEditable=\"true\"")
	}

	// Add ID if present
	if node.Id != "" {
		buffer.WriteString(" id=\"" + node.Id + "\"")
	}

	// Add ID if present
	if node.Title != "" {
		buffer.WriteString(" title=\"" + node.Title + "\"")
	}

	// Add ID if present
	if node.Src != "" {
		buffer.WriteString(" src=\"" + node.Src + "\"")
	}

	// Add ID if present
	if node.Href != "" {
		buffer.WriteString(" href=\"" + node.Href + "\"")
	}

	// Add class list if present
	if len(node.ClassList.Classes) > 0 || node.ClassList.Value != "" {
		classes := ""
		for _, v := range node.ClassList.Classes {
			if len(v) > 0 {
				if string(v[0]) != ":" {
					classes += v + " "
				}
			}
		}
		classes = strings.TrimSpace(classes)
		if len(classes) > 0 {
			buffer.WriteString(" class=\"" + classes + "\"")
		}
	}

	// Add style if present
	if len(node.Style) > 0 {

		style := ""
		for key, value := range node.Style {
			if key != "inlineText" {
				style += key + ":" + value + ";"
			}
		}
		style = strings.TrimSpace(style)

		if len(style) > 0 {
			buffer.WriteString(" style=\"" + style + "\"")
		}
	}

	// Add other attributes if present
	for key, value := range node.Attribute {
		if strings.TrimSpace(value) != "" {
			buffer.WriteString(" " + key + "=\"" + value + "\"")
		}
	}

	buffer.WriteString(">")

	// Add inner text if present
	if node.InnerText != "" && !ChildrenHaveText(node) {
		buffer.WriteString(node.InnerText)
	}
	return buffer.String(), "</" + node.TagName + ">"
}

func OuterHTML(node *element.Node) string {
	var buffer bytes.Buffer

	tag, closing := NodeToHTML(node)

	buffer.WriteString(tag)

	// Recursively add children
	for _, child := range node.Children {
		buffer.WriteString(OuterHTML(child))
	}

	buffer.WriteString(closing)

	return buffer.String()
}

func InnerHTML(node *element.Node) string {
	var buffer bytes.Buffer
	// Recursively add children
	for _, child := range node.Children {
		buffer.WriteString(OuterHTML(child))
	}
	return buffer.String()
}

func ReverseSlice[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func ParentStyleProp(n *element.Node, prop string, selector func(string) bool) bool {
	if n.Parent != nil {
		if selector(n.Parent.Style[prop]) {
			return true
		} else {
			return ParentStyleProp(n.Parent, prop, selector)
		}
	}
	return false
}
func RGBAtoString(c ic.RGBA) string {
	return fmt.Sprintf("R%d%d%d%d", c.R, c.G, c.B, c.A)
}
