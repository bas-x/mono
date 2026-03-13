package app

import (
	"fmt"
	"image/color"

	"github.com/ebitenui/ebitenui"
	ebitenuiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"

	"github.com/bas-x/basex/services"
)

const sidebarWidth = 332

func buildSidebarUI(r *Runtime) (*ebitenui.UI, error) {
	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	tabStrip := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				Padding:            &widget.Insets{Top: 12, Left: 12, Right: sidebarWidth + 24},
			}),
		),
		widget.ContainerOpts.BackgroundImage(
			ebitenuiimage.NewNineSliceColor(color.NRGBA{R: 23, G: 28, B: 39, A: 220}),
		),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(6),
			widget.RowLayoutOpts.Padding(&widget.Insets{Top: 6, Left: 6, Right: 6, Bottom: 6}),
		)),
	)
	for _, tab := range r.tabs {
		tabID := tab.SimulationID
		selected := tabID == r.activeSimulationID
		tabStrip.AddChild(newTabButton(r, tab.Label, selected, func(id string) func() {
			return func() {
				r.state.Error = ""
				if err := r.setActiveTab(id); err != nil {
					r.state.SetError(err)
				}
			}
		}(tabID)))
	}

	toolbar := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				Padding:            &widget.Insets{Top: 56, Left: 12, Right: sidebarWidth + 24},
			}),
		),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(8),
		)),
	)
	toolbar.AddChild(newButton(r, "Create branch", func() {
		r.state.Error = ""
		if _, err := r.createBranchFromActiveTab(); err != nil {
			r.state.SetError(err)
		}
		r.uiDirty = true
	}))
	r.startButton = newButton(r, "Start simulation", func() {
		r.state.Error = ""
		var err error
		switch {
		case r.state.Paused:
			err = r.resumeSimulation()
		case r.state.Running:
			err = r.pauseSimulation()
		default:
			err = r.startSimulation()
		}
		if err != nil {
			r.state.SetError(err)
		}
		r.uiDirty = true
	})
	toolbar.AddChild(r.startButton)
	toolbar.AddChild(newButton(r, "Reset simulation", func() {
		r.state.Error = ""
		if err := r.resetSimulation(); err != nil {
			r.state.SetError(err)
		}
		r.uiDirty = true
	}))

	sidebar := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				StretchVertical:    true,
				Padding:            &widget.Insets{Top: 12, Right: 12, Bottom: 12},
			}),
			widget.WidgetOpts.MinSize(sidebarWidth, r.config.WindowHeight-24),
		),
		widget.ContainerOpts.BackgroundImage(
			ebitenuiimage.NewNineSliceColor(color.NRGBA{R: 32, G: 38, B: 52, A: 220}),
		),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(8),
			widget.RowLayoutOpts.Padding(&widget.Insets{Top: 12, Left: 12, Right: 12, Bottom: 12}),
		)),
	)

	addSectionTitle(sidebar, r, "Simulation")
	activeLabel := "Active simulation: none"
	for _, tab := range r.tabs {
		if tab.SimulationID == r.activeSimulationID {
			activeLabel = fmt.Sprintf("Active simulation: %s", tab.Label)
			break
		}
	}
	sidebar.AddChild(newLabel(r, activeLabel, textSecondary))
	r.statusLabel = newLabel(r, fmt.Sprintf("Status: %s | Tick: %d", r.state.Status, r.state.Tick), textSecondary)
	sidebar.AddChild(r.statusLabel)
	timeText := "Time: n/a"
	if !r.state.CurrentTime.IsZero() {
		timeText = fmt.Sprintf("Time: %s", r.state.CurrentTime.Format("2006-01-02 15:04:05"))
	}
	r.timeLabel = newLabel(r, timeText, textSecondary)
	sidebar.AddChild(r.timeLabel)

	addSectionTitle(sidebar, r, fmt.Sprintf("Airbases (%d)", len(r.state.Airbases)))
	r.airbaseRows = make([]*widget.Button, 6)
	for i := range r.airbaseRows {
		btn := newListRow(r, "", false, func(index int) func() {
			return func() {
				if index < len(r.state.Airbases) {
					r.state.SelectAirbaseIndex(index)
					r.viewport = r.viewport.FocusAirbase(r.state.Airbases[index])
					r.state.addEvent("selected airbase "+shortID(r.state.Airbases[index].ID), true)
					r.uiDirty = true
				}
			}
		}(i))
		r.airbaseRows[i] = btn
		sidebar.AddChild(btn)
	}

	addSectionTitle(sidebar, r, fmt.Sprintf("Aircraft (%d)", len(r.state.Aircraft)))
	r.aircraftRows = make([]*widget.Button, 8)
	for i := range r.aircraftRows {
		btn := newListRow(r, "", false, func(index int) func() {
			return func() {
				if index < len(r.state.Aircraft) {
					r.state.SelectAircraftIndex(index)
					r.state.addEvent("selected aircraft "+shortID(r.state.Aircraft[index].TailNumber), true)
					r.uiDirty = true
				}
			}
		}(i))
		r.aircraftRows[i] = btn
		sidebar.AddChild(btn)
	}

	root.AddChild(tabStrip)
	root.AddChild(toolbar)
	root.AddChild(sidebar)
	return &ebitenui.UI{Container: root}, nil
}

func addSectionTitle(parent *widget.Container, r *Runtime, text string) {
	parent.AddChild(newLabel(r, text, textPrimary))
}

func newLabel(r *Runtime, text string, col color.Color) *widget.Label {
	return widget.NewLabel(
		widget.LabelOpts.Text(text, r.uiFace, &widget.LabelColor{Idle: col}),
	)
}

func newButton(r *Runtime, text string, onClick func()) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    ebitenuiimage.NewBorderedNineSliceColor(color.NRGBA{R: 67, G: 97, B: 145, A: 255}, color.NRGBA{R: 95, G: 125, B: 180, A: 255}, 2),
			Hover:   ebitenuiimage.NewBorderedNineSliceColor(color.NRGBA{R: 84, G: 118, B: 171, A: 255}, color.NRGBA{R: 114, G: 151, B: 214, A: 255}, 2),
			Pressed: ebitenuiimage.NewBorderedNineSliceColor(color.NRGBA{R: 54, G: 78, B: 117, A: 255}, color.NRGBA{R: 95, G: 125, B: 180, A: 255}, 2),
		}),
		widget.ButtonOpts.Text(text, r.uiFace, &widget.ButtonTextColor{Idle: textPrimary}),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 10, Right: 10, Top: 6, Bottom: 6}),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			if onClick != nil {
				onClick()
			}
		}),
	)
}

func newListRow(r *Runtime, text string, selected bool, onClick func()) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.Image(listRowButtonImage(selected)),
		widget.ButtonOpts.Text(text, r.uiFace, &widget.ButtonTextColor{Idle: textPrimary}),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 8, Right: 8, Top: 4, Bottom: 4}),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			if onClick != nil {
				onClick()
			}
		}),
	)
}

func newTabButton(r *Runtime, text string, selected bool, onClick func()) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.Image(tabButtonImage(selected)),
		widget.ButtonOpts.Text(text, r.uiFace, &widget.ButtonTextColor{Idle: textPrimary}),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 14, Right: 14, Top: 8, Bottom: 8}),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			if onClick != nil {
				onClick()
			}
		}),
	)
}

func listRowButtonImage(selected bool) *widget.ButtonImage {
	idle := color.NRGBA{R: 48, G: 57, B: 77, A: 255}
	hover := color.NRGBA{R: 66, G: 78, B: 103, A: 255}
	pressed := color.NRGBA{R: 74, G: 88, B: 116, A: 255}
	if selected {
		idle = color.NRGBA{R: 84, G: 108, B: 153, A: 255}
		hover = color.NRGBA{R: 96, G: 121, B: 170, A: 255}
		pressed = color.NRGBA{R: 70, G: 90, B: 128, A: 255}
	}
	return &widget.ButtonImage{
		Idle:    ebitenuiimage.NewNineSliceColor(idle),
		Hover:   ebitenuiimage.NewNineSliceColor(hover),
		Pressed: ebitenuiimage.NewNineSliceColor(pressed),
	}
}

func tabButtonImage(selected bool) *widget.ButtonImage {
	idle := color.NRGBA{R: 39, G: 47, B: 64, A: 255}
	hover := color.NRGBA{R: 55, G: 66, B: 89, A: 255}
	pressed := color.NRGBA{R: 65, G: 78, B: 104, A: 255}
	border := color.NRGBA{R: 70, G: 83, B: 111, A: 255}
	if selected {
		idle = color.NRGBA{R: 84, G: 108, B: 153, A: 255}
		hover = color.NRGBA{R: 96, G: 121, B: 170, A: 255}
		pressed = color.NRGBA{R: 70, G: 90, B: 128, A: 255}
		border = color.NRGBA{R: 124, G: 151, B: 207, A: 255}
	}
	return &widget.ButtonImage{
		Idle:    ebitenuiimage.NewBorderedNineSliceColor(idle, border, 2),
		Hover:   ebitenuiimage.NewBorderedNineSliceColor(hover, border, 2),
		Pressed: ebitenuiimage.NewBorderedNineSliceColor(pressed, border, 2),
	}
}

func selectedAirbase(state *State) *services.Airbase {
	if state == nil {
		return nil
	}
	for i := range state.Airbases {
		if state.Airbases[i].ID == state.SelectedAirbaseID {
			return &state.Airbases[i]
		}
	}
	return nil
}

func selectedAircraft(state *State) *services.Aircraft {
	if state == nil {
		return nil
	}
	for i := range state.Aircraft {
		if state.Aircraft[i].TailNumber == state.SelectedAircraft {
			return &state.Aircraft[i]
		}
	}
	return nil
}
