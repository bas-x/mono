package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/ebitenui/ebitenui"
	ebitenuiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	textv2 "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/bas-x/basex/services"
)

const (
	shellInset          = 12
	toolbarSpacing      = 8
	toolbarButtonW      = 112
	toolbarButtonH      = 36
	toolbarInputMinW    = 280
	toolbarControlH     = 36
	toolbarFontSize     = 16
	toolbarPaddingX     = 10
	toolbarPaddingY     = 8
	textInputPaddingX   = 10
	textInputPaddingY   = 9
	tabStripSpacing     = 6
	tabButtonMinW       = 120
	addTabButtonW       = 40
	tabButtonH          = 32
	tabViewportPaddingX = 10
	tabViewportPaddingY = 10
)

var (
	shellFrameColor          = color.NRGBA{R: 18, G: 22, B: 30, A: 255}
	toolbarColor             = color.NRGBA{R: 28, G: 34, B: 46, A: 255}
	viewportFrameColor       = color.NRGBA{R: 26, G: 31, B: 42, A: 255}
	tabStripColor            = color.NRGBA{R: 31, G: 37, B: 49, A: 255}
	tabBodyColor             = color.NRGBA{R: 34, G: 40, B: 54, A: 255}
	tabIdleColor             = color.NRGBA{R: 50, G: 59, B: 77, A: 255}
	tabHoverColor            = color.NRGBA{R: 68, G: 80, B: 101, A: 255}
	tabPressedColor          = color.NRGBA{R: 83, G: 97, B: 120, A: 255}
	tabActiveColor           = color.NRGBA{R: 102, G: 118, B: 144, A: 255}
	inputIdleColor           = color.NRGBA{R: 22, G: 27, B: 37, A: 255}
	inputDisabledColor       = color.NRGBA{R: 41, G: 46, B: 58, A: 255}
	buttonIdleColor          = color.NRGBA{R: 54, G: 65, B: 83, A: 255}
	buttonHoverColor         = color.NRGBA{R: 72, G: 84, B: 106, A: 255}
	buttonPressedColor       = color.NRGBA{R: 89, G: 102, B: 125, A: 255}
	buttonDisabledColor      = color.NRGBA{R: 52, G: 57, B: 68, A: 255}
	toolbarTextColor         = color.NRGBA{R: 234, G: 239, B: 245, A: 255}
	toolbarDisabledTextColor = color.NRGBA{R: 151, G: 160, B: 173, A: 255}
)

type textShell struct {
	options          launchOptions
	service          *services.SimulationService
	workspace        shellWorkspace
	ui               *ebitenui.UI
	seedInput        *widget.TextInput
	infoLabel        *widget.Text
	pendingSelectID  string
	pendingBranch    bool
	pendingReset     bool
	pendingPlayPause bool
	qaFrameArmed     bool
	qaFrameCaptured  bool
	qaErr            error
	closed           bool
}

func newTextShell(options launchOptions, service *services.SimulationService) (*textShell, error) {
	if service == nil {
		return nil, fmt.Errorf("simulation service is required")
	}
	if options.windowWidth <= 0 {
		return nil, fmt.Errorf("window width must be positive")
	}
	if options.windowHeight <= 0 {
		return nil, fmt.Errorf("window height must be positive")
	}
	if options.windowTitle == "" {
		options.windowTitle = defaultWindowTitle
	}

	workspace, err := bootstrapWorkspace(service)
	if err != nil {
		return nil, err
	}

	shell := &textShell{
		options:   options,
		service:   service,
		workspace: workspace,
	}
	if err := shell.rebuildUI(); err != nil {
		return nil, err
	}

	return shell, nil
}

func (s *textShell) Update() error {
	if s == nil || s.closed {
		return nil
	}
	if s.qaErr != nil {
		return s.qaErr
	}
	if s.ui != nil {
		s.ui.Update()
	}
	s.refreshInfoLabel()
	if s.pendingBranch {
		s.pendingBranch = false
		workspace, err := branchBaseTab(s.service)
		if err != nil {
			return err
		}
		s.workspace = workspace
		if err := s.rebuildUI(); err != nil {
			return err
		}
	}
	if s.pendingReset {
		s.pendingReset = false
		seed := seedFromInput(s.seedInput)
		s.service.Reset()
		workspace, err := bootstrapWorkspaceWithSeed(s.service, seed)
		if err != nil {
			return err
		}
		s.workspace = workspace
		if err := s.rebuildUI(); err != nil {
			return err
		}
	}
	if s.pendingPlayPause {
		s.pendingPlayPause = false
		if err := s.togglePlayPause(); err != nil {
			return err
		}
	}
	if s.pendingSelectID != "" && s.pendingSelectID != s.workspace.activeTabID {
		workspace, err := workspaceForActiveTab(s.service, s.pendingSelectID)
		if err != nil {
			return err
		}
		s.pendingSelectID = ""
		s.workspace = workspace
		if err := s.rebuildUI(); err != nil {
			return err
		}
	} else {
		s.pendingSelectID = ""
	}
	if !s.options.exitAfterQA {
		return nil
	}
	if !s.qaFrameArmed {
		s.qaFrameArmed = true
		return nil
	}
	if s.qaFrameCaptured {
		return ebiten.Termination
	}
	return nil
}

func (s *textShell) Draw(screen *ebiten.Image) {
	if s == nil || screen == nil {
		return
	}
	if s.ui != nil {
		s.ui.Draw(screen)
	}
	if !s.options.exitAfterQA || !s.qaFrameArmed || s.qaFrameCaptured || s.qaErr != nil {
		return
	}
	if err := captureShellFrame(screen, s.options.screenshotOut); err != nil {
		s.qaErr = err
		return
	}
	s.qaFrameCaptured = true
}

func (s *textShell) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth <= 0 {
		outsideWidth = s.options.windowWidth
	}
	if outsideHeight <= 0 {
		outsideHeight = s.options.windowHeight
	}
	return outsideWidth, outsideHeight
}

func (s *textShell) Close() {
	if s == nil || s.closed {
		return
	}
	s.closed = true
	if s.service != nil {
		s.service.Reset()
	}
}

func (s *textShell) rebuildUI() error {
	ui, seedInput, infoLabel, err := buildTextShellUI(s.workspace, func(tabID string) {
		s.pendingSelectID = tabID
	}, func() {
		s.pendingBranch = true
	}, func() {
		s.pendingReset = true
	}, func() {
		s.pendingPlayPause = true
	})
	if err != nil {
		return err
	}
	s.ui = ui
	s.seedInput = seedInput
	s.infoLabel = infoLabel
	return nil
}

func buildTextShellUI(workspace shellWorkspace, onSelect func(string), onBranch func(), onReset func(), onPlayPause func()) (*ebitenui.UI, *widget.TextInput, *widget.Text, error) {
	face, err := loadToolbarFace(toolbarFontSize)
	if err != nil {
		return nil, nil, nil, err
	}

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ebitenuiimage.NewNineSliceColor(shellFrameColor)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(shellInset),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Left:   shellInset,
				Right:  shellInset,
				Top:    shellInset,
				Bottom: shellInset,
			}),
		)),
	)

	toolbar := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ebitenuiimage.NewNineSliceColor(toolbarColor)),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Stretch([]bool{true, false, false}, []bool{true}),
			widget.GridLayoutOpts.Spacing(toolbarSpacing, 0),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:   toolbarPaddingX,
				Right:  toolbarPaddingX,
				Top:    toolbarPaddingY,
				Bottom: toolbarPaddingY,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)

	seedInput := widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(toolbarInputMinW, toolbarControlH),
		),
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle:     ebitenuiimage.NewNineSliceColor(inputIdleColor),
			Disabled: ebitenuiimage.NewNineSliceColor(inputDisabledColor),
		}),
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          toolbarTextColor,
			Disabled:      toolbarDisabledTextColor,
			Caret:         toolbarTextColor,
			DisabledCaret: toolbarDisabledTextColor,
		}),
		widget.TextInputOpts.Padding(&widget.Insets{
			Left:   textInputPaddingX,
			Right:  textInputPaddingX,
			Top:    textInputPaddingY,
			Bottom: textInputPaddingY,
		}),
		widget.TextInputOpts.Face(face),
		widget.TextInputOpts.CaretWidth(1),
		widget.TextInputOpts.Placeholder("Seed"),
	)

	toolbar.AddChild(seedInput)

	resetBtn := newToolbarButton("Reset", face)
	resetBtn.ClickedEvent.AddHandler(func(_ any) { onReset() })
	toolbar.AddChild(resetBtn)

	playPauseLabel := playPauseButtonLabel(workspace.activeSummary)
	playPauseBtn := newToolbarButton(playPauseLabel, face)
	playPauseBtn.ClickedEvent.AddHandler(func(_ any) { onPlayPause() })
	toolbar.AddChild(playPauseBtn)

	viewport := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ebitenuiimage.NewNineSliceColor(viewportFrameColor)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(0),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Left:   tabViewportPaddingX,
				Right:  tabViewportPaddingX,
				Top:    tabViewportPaddingY,
				Bottom: tabViewportPaddingY,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)

	tabStrip := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ebitenuiimage.NewNineSliceColor(tabStripColor)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(tabStripSpacing),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Left:   toolbarPaddingX,
				Right:  toolbarPaddingX,
				Top:    toolbarPaddingY,
				Bottom: toolbarPaddingY,
			}),
		)),
	)
	for _, tab := range workspace.tabs {
		tabStrip.AddChild(newSimulationTabButton(tab.id, tab.id == workspace.activeTabID, face, func(id string) func(*widget.ButtonClickedEventArgs) {
			return func(*widget.ButtonClickedEventArgs) {
				onSelect(id)
			}
		}(tab.id)))
	}
	tabStrip.AddChild(newAddTabButton(face, func(*widget.ButtonClickedEventArgs) {
		onBranch()
	}))

	tabBody := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ebitenuiimage.NewNineSliceColor(tabBodyColor)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Left:   tabViewportPaddingX,
				Right:  tabViewportPaddingX,
				Top:    tabViewportPaddingY,
				Bottom: tabViewportPaddingY,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)

	s := workspace.activeSummary
	infoText := fmt.Sprintf(
		"Tick: %d    Time: %s    Aircraft: %d    Airbases: %d    Threats: %d",
		s.tick, s.timestamp, s.aircraftCount, s.airbaseCount, s.threatCount,
	)
	infoLabel := widget.NewText(
		widget.TextOpts.Text(infoText, face, toolbarTextColor),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	tabBody.AddChild(infoLabel)

	root.AddChild(toolbar)
	viewport.AddChild(tabStrip)
	viewport.AddChild(tabBody)
	root.AddChild(viewport)

	return &ebitenui.UI{Container: root}, seedInput, infoLabel, nil
}

func newToolbarButton(label string, face *textv2.Face) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(toolbarButtonW, toolbarButtonH),
		),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:     ebitenuiimage.NewNineSliceColor(buttonIdleColor),
			Hover:    ebitenuiimage.NewNineSliceColor(buttonHoverColor),
			Pressed:  ebitenuiimage.NewNineSliceColor(buttonPressedColor),
			Disabled: ebitenuiimage.NewNineSliceColor(buttonDisabledColor),
		}),
		widget.ButtonOpts.Text(label, face, &widget.ButtonTextColor{
			Idle:     toolbarTextColor,
			Disabled: toolbarDisabledTextColor,
			Hover:    toolbarTextColor,
			Pressed:  toolbarTextColor,
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   toolbarPaddingX,
			Right:  toolbarPaddingX,
			Top:    toolbarPaddingY,
			Bottom: toolbarPaddingY,
		}),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {}),
	)
}

func newSimulationTabButton(label string, active bool, face *textv2.Face, handler func(*widget.ButtonClickedEventArgs)) *widget.Button {
	idleColor := tabIdleColor
	hoverColor := tabHoverColor
	pressedColor := tabPressedColor
	if active {
		idleColor = tabActiveColor
		hoverColor = tabActiveColor
		pressedColor = tabActiveColor
	}

	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(tabButtonMinW, tabButtonH),
		),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:     ebitenuiimage.NewNineSliceColor(idleColor),
			Hover:    ebitenuiimage.NewNineSliceColor(hoverColor),
			Pressed:  ebitenuiimage.NewNineSliceColor(pressedColor),
			Disabled: ebitenuiimage.NewNineSliceColor(buttonDisabledColor),
		}),
		widget.ButtonOpts.Text(label, face, &widget.ButtonTextColor{
			Idle:     toolbarTextColor,
			Disabled: toolbarDisabledTextColor,
			Hover:    toolbarTextColor,
			Pressed:  toolbarTextColor,
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   toolbarPaddingX,
			Right:  toolbarPaddingX,
			Top:    toolbarPaddingY,
			Bottom: toolbarPaddingY,
		}),
		widget.ButtonOpts.ClickedHandler(handler),
	)
}

func newAddTabButton(face *textv2.Face, handler func(*widget.ButtonClickedEventArgs)) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(addTabButtonW, tabButtonH),
		),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:     ebitenuiimage.NewNineSliceColor(tabIdleColor),
			Hover:    ebitenuiimage.NewNineSliceColor(tabHoverColor),
			Pressed:  ebitenuiimage.NewNineSliceColor(tabPressedColor),
			Disabled: ebitenuiimage.NewNineSliceColor(buttonDisabledColor),
		}),
		widget.ButtonOpts.Text("+", face, &widget.ButtonTextColor{
			Idle:     toolbarTextColor,
			Disabled: toolbarDisabledTextColor,
			Hover:    toolbarTextColor,
			Pressed:  toolbarTextColor,
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   toolbarPaddingX,
			Right:  toolbarPaddingX,
			Top:    toolbarPaddingY,
			Bottom: toolbarPaddingY,
		}),
		widget.ButtonOpts.ClickedHandler(handler),
	)
}

func seedFromInput(input *widget.TextInput) [32]byte {
	if input == nil {
		return [32]byte{}
	}
	text := strings.TrimSpace(input.GetText())
	if text == "" {
		return [32]byte{}
	}
	return sha256.Sum256([]byte(text))
}

func (s *textShell) refreshInfoLabel() {
	if s.infoLabel == nil {
		return
	}
	info, err := s.service.Simulation(s.workspace.activeTabID)
	if err != nil {
		return
	}
	ts := "—"
	if !info.Timestamp.IsZero() {
		ts = info.Timestamp.Format("15:04:05")
	}
	s.infoLabel.Label = fmt.Sprintf(
		"Tick: %d    Time: %s    Aircraft: %d    Airbases: %d    Threats: %d",
		info.Tick, ts,
		s.workspace.activeSummary.aircraftCount,
		s.workspace.activeSummary.airbaseCount,
		s.workspace.activeSummary.threatCount,
	)
}

func (s *textShell) togglePlayPause() error {
	info := s.workspace.activeSummary
	if !info.running {
		if err := s.service.StartSimulation(s.workspace.activeTabID); err != nil {
			return err
		}
	} else if !info.paused {
		if err := s.service.PauseSimulation(s.workspace.activeTabID); err != nil {
			return err
		}
	} else {
		if err := s.service.ResumeSimulation(s.workspace.activeTabID); err != nil {
			return err
		}
	}

	workspace, err := workspaceForActiveTab(s.service, s.workspace.activeTabID)
	if err != nil {
		return err
	}
	s.workspace = workspace
	return s.rebuildUI()
}

func playPauseButtonLabel(summary tabSummary) string {
	if summary.running && !summary.paused {
		return "Pause"
	}
	return "Play"
}

func loadToolbarFace(size float64) (*textv2.Face, error) {
	source, err := textv2.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		return nil, fmt.Errorf("load text toolbar font: %w", err)
	}
	face := textv2.Face(&textv2.GoTextFace{Source: source, Size: size})
	return &face, nil
}

func captureShellFrame(screen *ebiten.Image, screenshotOut string) error {
	if strings.TrimSpace(screenshotOut) == "" {
		return nil
	}
	frame := image.NewRGBA(screen.Bounds())
	pixels := make([]byte, 4*screen.Bounds().Dx()*screen.Bounds().Dy())
	screen.ReadPixels(pixels)
	copy(frame.Pix, pixels)

	if err := writePNG(screenshotOut, frame); err != nil {
		return fmt.Errorf("write QA screenshot %q: %w", screenshotOut, err)
	}
	return nil
}

func writePNG(path string, img image.Image) error {
	if img == nil {
		return fmt.Errorf("screenshot image is required")
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create screenshot directory %q: %w", dir, err)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create screenshot file %q: %w", path, err)
	}
	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		return fmt.Errorf("encode screenshot %q: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close screenshot file %q: %w", path, err)
	}

	return nil
}
