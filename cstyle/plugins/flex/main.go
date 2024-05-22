package flex

import (
	"fmt"
	"gui/cstyle"
	"gui/element"
	"gui/utils"
	"slices"
	"strings"
)

func Init() cstyle.Plugin {
	return cstyle.Plugin{
		Selector: func(n *element.Node) bool {
			styles := map[string]string{
				"display": "flex",
				// "justify-content": "*",
				// "align-content":   "*",
				// "align-items":     "*",
				// "flex-wrap":       "*",
				// "flex-direction":  "*",
			}
			matches := true
			for name, value := range styles {
				if n.Style[name] != value && !(value == "*") && n.Style[name] != "" {
					matches = false
				}
			}
			return matches
		},
		Level: 2,
		Handler: func(n *element.Node, state *map[string]element.State) {
			// !ISSUE: align-items is not impleamented
			// + issues with the width when notaspans are included
			s := *state
			self := s[n.Properties.Id]
			// Brief: justify does not align the bottom row correctly
			//        y axis also needs to be done
			verbs := strings.Split(n.Style["flex-direction"], "-")

			orderedNode := order(*n, state, n.Children, verbs[0], len(verbs) > 1, n.Style["flex-wrap"] == "wrap")
			// fmt.Println(orderedNode)
			fmt.Println("######")
			for y := range orderedNode[0] {
				fmt.Print(y, " ")
			}
			fmt.Print("\n")

			for x, row := range orderedNode {
				fmt.Print(x, " ")
				for _, col := range row {
					fmt.Print(col.Properties.Id, " ")
				}
				fmt.Print("\n")
			}

			// c := s[n.Children[0].Properties.Id]
			// c.Width = 100
			// (*state)[n.Children[0].Properties.Id] = c

			(*state)[n.Properties.Id] = self
		},
	}
}

func order(p element.Node, state *map[string]element.State, elements []element.Node, direction string, reversed, wrap bool) [][]element.Node {
	// Get the state of the parent node
	s := *state
	self := s[p.Properties.Id]

	// Variables for handling direction and margins
	var dir, marginStart, marginEnd string
	if direction == "column" {
		dir = "Height"
		marginStart = "Top"
		marginEnd = "Bottom"
	} else {
		dir = "Width"
		marginStart = "Left"
		marginEnd = "Right"
	}

	// Get the maximum size in the specified direction
	max, _ := utils.GetStructField(&self, dir)

	// Container for the ordered nodes
	nodes := [][]element.Node{}

	if wrap {
		// If wrapping is enabled
		counter := 0
		if direction == "column" {
			// Collect nodes for column direction
			collector := []element.Node{}
			for _, v := range elements {
				vState := s[v.Properties.Id]
				elMax := vState.Height
				elMS, _ := utils.GetStructField(&vState.Margin, marginStart)
				elME, _ := utils.GetStructField(&vState.Margin, marginEnd)
				tMax := elMax + elMS.(float32) + elME.(float32)

				// Check if the current element can fit in the current row/column
				if counter+int(tMax) < int(max.(float32)) {
					collector = append(collector, v)
				} else {
					if reversed {
						slices.Reverse(collector)
					}
					nodes = append(nodes, collector)
					collector = []element.Node{}
					collector = append(collector, v)
					counter = 0
				}
				counter += int(tMax)
			}
			if len(collector) > 0 {
				nodes = append(nodes, collector)
			}
		} else {
			// Collect nodes for row direction
			var mod int
			for _, v := range elements {
				vState := s[v.Properties.Id]
				elMax := vState.Width
				elMS, _ := utils.GetStructField(&vState.Margin, marginStart)
				elME, _ := utils.GetStructField(&vState.Margin, marginEnd)
				tMax := elMax + elMS.(float32) + elME.(float32)

				// Check if the current element can fit in the current row/column
				if counter+int(tMax) < int(max.(float32)) {
					if len(nodes)-1 < mod {
						nodes = append(nodes, []element.Node{v})
					} else {
						nodes[mod] = append(nodes[mod], v)
					}
				} else {
					mod = 0
					counter = 0
					if len(nodes)-1 < mod {
						nodes = append(nodes, []element.Node{v})
					} else {
						nodes[mod] = append(nodes[mod], v)
					}
				}
				counter += int(tMax)
				mod++
			}
			if reversed {
				slices.Reverse(nodes)
			}
		}
	} else {
		// If wrapping is not enabled
		var tMax float32
		for _, v := range elements {
			vState := s[v.Properties.Id]
			elMax, _ := utils.GetStructField(&vState, dir)
			elMS, _ := utils.GetStructField(&vState.Margin, marginStart)
			elME, _ := utils.GetStructField(&vState.Margin, marginEnd)
			tMax += elMax.(float32) + elMS.(float32) + elME.(float32)
		}

		pMax, _ := utils.GetStructField(&self, dir)

		// Resize nodes to fit
		var newSize float32
		if tMax > pMax.(float32) {
			newSize = pMax.(float32) / float32(len(elements))
		}
		if dir == "Width" {
			for _, v := range elements {
				vState := s[v.Properties.Id]
				if newSize != 0 {
					vState.Width = newSize
				}
				nodes = append(nodes, []element.Node{v})
			}
			if reversed {
				slices.Reverse(nodes)
			}
		} else {
			nodes = append(nodes, []element.Node{})
			for _, v := range elements {
				vState := s[v.Properties.Id]
				if newSize != 0 {
					vState.Height = newSize
				}
				nodes[0] = append(nodes[0], v)
			}
			if reversed {
				slices.Reverse(nodes[0])
			}
		}
	}

	// Update the state of the parent node
	(*state)[p.Properties.Id] = self

	return nodes
}
