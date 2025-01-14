package textAlign

import (
	"gui/cstyle"
	"gui/element"
)

func Init() cstyle.Plugin {
	return cstyle.Plugin{
		Selector: func(n *element.Node) bool {
			styles := map[string]string{
				"text-align": "*",
			}
			matches := true
			for name, value := range styles {
				if (n.Style[name] != value || n.Style[name] == "") && !(value == "*") {
					matches = false
				}
			}
			return matches
		},
		Level: 3,
		Handler: func(n *element.Node, state *map[string]element.State) {
			s := *state
			self := s[n.Properties.Id]
			minX := float32(9e15)
			maxXW := float32(0)

			nChildren := []*element.Node{}

			for _, v := range n.Children {
				// This prevents using absolutely positionioned elements in the alignment of text
				// + Will need to add the other styles
				if v.Style["position"] != "absolute" {
					nChildren = append(nChildren, v)
				}
			}

			if n.Style["text-align"] == "center" {
				if len(nChildren) > 0 {
					minX = s[nChildren[0].Properties.Id].X
					baseY := s[nChildren[0].Properties.Id].Y + s[nChildren[0].Properties.Id].Height
					last := 0
					for i := 1; i < len(nChildren)-1; i++ {
						cState := s[nChildren[i].Properties.Id]
						next := s[nChildren[i+1].Properties.Id]

						if cState.X < minX {
							minX = cState.X
						}
						if (cState.X + cState.Width) > maxXW {
							maxXW = cState.X + cState.Width
						}

						if baseY != next.Y+next.Height {
							baseY = next.Y + next.Height
							for a := last; a < i+1; a++ {
								cState := s[nChildren[a].Properties.Id]
								cState.X += self.Padding.Left + ((((self.Width - (self.Padding.Left + self.Padding.Right)) + (self.Border.Left.Width + self.Border.Right.Width)) - (maxXW - minX)) / 2) - (minX - self.X)
								(*state)[nChildren[a].Properties.Id] = cState
							}
							minX = 9e15
							maxXW = 0
							last = i + 1
						}

					}
					minX = s[nChildren[last].Properties.Id].X
					maxXW = s[nChildren[len(nChildren)-1].Properties.Id].X + s[nChildren[len(nChildren)-1].Properties.Id].Width
					for a := last; a < len(nChildren); a++ {
						cState := s[nChildren[a].Properties.Id]
						cState.X += self.Padding.Left + ((((self.Width - (self.Padding.Left + self.Padding.Right)) + (self.Border.Left.Width + self.Border.Right.Width)) - (maxXW - minX)) / 2) - (minX - self.X)
						(*state)[nChildren[a].Properties.Id] = cState
					}
				}
			} else if n.Style["text-align"] == "right" {
				if len(nChildren) > 0 {
					minX = s[nChildren[0].Properties.Id].X
					baseY := s[nChildren[0].Properties.Id].Y + s[nChildren[0].Properties.Id].Height
					last := 0
					for i := 1; i < len(nChildren)-1; i++ {

						cState := s[nChildren[i].Properties.Id]
						next := s[nChildren[i+1].Properties.Id]

						if cState.X < minX {
							minX = cState.X
						}
						if (cState.X + cState.Width) > maxXW {
							maxXW = cState.X + cState.Width
						}

						if baseY != next.Y+next.Height {
							baseY = next.Y + next.Height
							for a := last; a < i+1; a++ {
								cState := s[nChildren[a].Properties.Id]
								cState.X += ((self.Width + (self.Border.Left.Width + self.Border.Right.Width)) - (maxXW - minX)) + ((self.X - minX) * 2)
								(*state)[nChildren[a].Properties.Id] = cState
							}
							minX = 9e15
							maxXW = 0
							last = i + 1
						}

					}
					minX = s[nChildren[last].Properties.Id].X
					maxXW = s[nChildren[len(nChildren)-1].Properties.Id].X + s[nChildren[len(nChildren)-1].Properties.Id].Width
					for a := last; a < len(nChildren); a++ {
						cState := s[nChildren[a].Properties.Id]
						cState.X += ((self.Width + (self.Border.Left.Width + self.Border.Right.Width)) - (maxXW - minX)) + ((self.X - minX) * 2)
						(*state)[nChildren[a].Properties.Id] = cState
					}

				}
			}

			(*state)[n.Properties.Id] = self
		},
	}
}
