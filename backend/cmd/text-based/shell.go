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
	"sort"
	"strings"
	"time"

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
	listEntryPaddingX   = 8
	listEntryPaddingY   = 4
	listFontSize        = 14
	listSliderMinHandle = 16
	listSliderPadding   = 2
	columnSpacing       = 12
	detailPanelSpacing  = 8
	detailPanelPaddingX = 8
	detailPanelPaddingY = 8
	detailPanelMinH     = 96
	recentEventsLimit   = 100
	eventsScrollMaxH    = 220
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
	listBgColor              = color.NRGBA{R: 26, G: 31, B: 42, A: 255}
	listSelectedTextColor    = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	listUnselectedTextColor  = color.NRGBA{R: 210, G: 216, B: 228, A: 255}
	listSelectedBgColor      = color.NRGBA{R: 54, G: 65, B: 83, A: 255}
	listSelectingBgColor     = color.NRGBA{R: 45, G: 54, B: 70, A: 255}
	listFocusedBgColor       = color.NRGBA{R: 40, G: 48, B: 63, A: 255}
	listSliderTrackColor     = color.NRGBA{R: 38, G: 45, B: 59, A: 255}
	listSliderHandleColor    = color.NRGBA{R: 70, G: 82, B: 103, A: 255}
	listSliderHandleHover    = color.NRGBA{R: 90, G: 104, B: 128, A: 255}
)

type aircraftEntry struct {
	tailNumber string
	needCount  int
	state      string
	posX       float64
	posY       float64
	needs      []services.Need
}

type airbaseEntry struct {
	id            string
	region        string
	aircraftCount int
	capabilities  map[string]services.AirbaseCapability
}

type threatEntry struct {
	id           string
	engagedCount int
}

type eventEntry struct {
	id     string
	text   string
	detail string
}

type shellUIResult struct {
	ui              *ebitenui.UI
	seedInput       *widget.TextInput
	infoLabel       *widget.Text
	aircraftContent *widget.Container
	aircraftButtons map[string]*widget.Button
	airbaseContent  *widget.Container
	airbaseButtons  map[string]*widget.Button
	threatContent   *widget.Container
	threatButtons   map[string]*widget.Button
	eventContent    *widget.Container
	eventButtons    []*widget.Button
	aircraftDetail  *widget.Text
	airbaseDetail   *widget.Text
	threatDetail    *widget.Text
	eventDetail     *widget.Text
	listFace        *textv2.Face
}

type textShell struct {
	options                 launchOptions
	service                 *services.SimulationService
	workspace               shellWorkspace
	ui                      *ebitenui.UI
	seedInput               *widget.TextInput
	infoLabel               *widget.Text
	aircraftContent         *widget.Container
	aircraftButtons         map[string]*widget.Button
	aircraftMap             map[string]aircraftEntry
	selectedAircraft        string
	airbaseContent          *widget.Container
	airbaseButtons          map[string]*widget.Button
	airbaseMap              map[string]airbaseEntry
	selectedAirbase         string
	threatContent           *widget.Container
	threatButtons           map[string]*widget.Button
	threatMap               map[string]threatEntry
	threatEngagedByThreatID map[string]map[string]bool
	selectedThreat          string
	eventContent            *widget.Container
	recentEvents            []eventEntry
	eventButtons            []*widget.Button
	eventsDirty             bool
	aircraftDetail          *widget.Text
	airbaseDetail           *widget.Text
	threatDetail            *widget.Text
	eventDetail             *widget.Text
	selectedEventID         string
	listFace                *textv2.Face
	subID                   uint64
	eventCh                 <-chan services.Event
	pendingSelectID         string
	pendingBranch           bool
	pendingReset            bool
	pendingPlayPause        bool
	aircraftDirty           bool
	airbaseDirty            bool
	threatDirty             bool
	lastAircraftPollTabID   string
	lastAircraftPollTail    string
	qaFrameArmed            bool
	qaFrameCaptured         bool
	qaErr                   error
	closed                  bool
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

	subID, eventCh := service.Broadcaster().Subscribe()

	shell := &textShell{
		options:                 options,
		service:                 service,
		workspace:               workspace,
		aircraftMap:             make(map[string]aircraftEntry),
		aircraftButtons:         make(map[string]*widget.Button),
		airbaseMap:              make(map[string]airbaseEntry),
		airbaseButtons:          make(map[string]*widget.Button),
		threatMap:               make(map[string]threatEntry),
		threatButtons:           make(map[string]*widget.Button),
		threatEngagedByThreatID: make(map[string]map[string]bool),
		recentEvents:            make([]eventEntry, 0, recentEventsLimit),
		eventButtons:            make([]*widget.Button, 0, recentEventsLimit),
		subID:                   subID,
		eventCh:                 eventCh,
	}
	shell.seedAircraftMap()
	shell.seedAirbaseMap()
	shell.seedThreatMap()
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
	s.drainEvents()
	if info, err := s.service.Simulation(s.workspace.activeTabID); err == nil {
		s.refreshInfoLabel(info)
		s.refreshSelectedAircraftFromService()
	}
	s.refreshAircraftList()
	s.refreshAirbaseList()
	s.refreshThreatList()
	s.refreshRecentEvents()
	s.refreshDetailLabels()
	if s.ui != nil {
		s.ui.Update()
	}
	if s.pendingBranch {
		s.pendingBranch = false
		workspace, err := branchBaseTab(s.service)
		if err != nil {
			return err
		}
		s.workspace = workspace
		s.lastAircraftPollTabID = ""
		s.lastAircraftPollTail = ""
		s.recentEvents = s.recentEvents[:0]
		s.selectedEventID = ""
		s.eventsDirty = true
		s.seedAircraftMap()
		s.seedAirbaseMap()
		s.seedThreatMap()
		if err := s.rebuildUI(); err != nil {
			return err
		}
	}
	if s.pendingReset {
		s.pendingReset = false
		seed := seedFromInput(s.seedInput)
		s.service.Broadcaster().Unsubscribe(s.subID)
		s.service.Reset()
		workspace, err := bootstrapWorkspaceWithSeed(s.service, seed)
		if err != nil {
			return err
		}
		s.workspace = workspace
		s.subID, s.eventCh = s.service.Broadcaster().Subscribe()
		s.lastAircraftPollTabID = ""
		s.lastAircraftPollTail = ""
		s.recentEvents = s.recentEvents[:0]
		s.selectedEventID = ""
		s.eventsDirty = true
		s.seedAircraftMap()
		s.seedAirbaseMap()
		s.seedThreatMap()
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
		s.lastAircraftPollTabID = ""
		s.lastAircraftPollTail = ""
		s.recentEvents = s.recentEvents[:0]
		s.selectedEventID = ""
		s.eventsDirty = true
		s.seedAircraftMap()
		s.seedAirbaseMap()
		s.seedThreatMap()
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
		s.service.Broadcaster().Unsubscribe(s.subID)
		s.service.Reset()
	}
}

func (s *textShell) rebuildUI() error {
	result, err := buildTextShellUI(
		s.workspace,
		s.aircraftMap, s.selectedAircraft,
		s.airbaseMap, s.selectedAirbase,
		s.threatMap, s.selectedThreat,
		s.recentEvents,
		s.selectedEventID,
		func(tabID string) {
			s.pendingSelectID = tabID
		},
		func() { s.pendingBranch = true },
		func() { s.pendingReset = true },
		func() { s.pendingPlayPause = true },
		func(tailNumber string) {
			s.selectedAircraft = tailNumber
			s.lastAircraftPollTail = ""
			s.updateAircraftSelection()
		},
		func(airbaseID string) {
			s.selectedAirbase = airbaseID
			s.updateAirbaseSelection()
		},
		func(threatID string) {
			s.selectedThreat = threatID
			s.updateThreatSelection()
		},
		func(eventID string) {
			s.selectedEventID = eventID
			s.updateEventSelection()
		},
	)
	if err != nil {
		return err
	}
	s.ui = result.ui
	s.seedInput = result.seedInput
	s.infoLabel = result.infoLabel
	s.aircraftContent = result.aircraftContent
	s.aircraftButtons = result.aircraftButtons
	s.airbaseContent = result.airbaseContent
	s.airbaseButtons = result.airbaseButtons
	s.threatContent = result.threatContent
	s.threatButtons = result.threatButtons
	s.eventContent = result.eventContent
	s.eventButtons = result.eventButtons
	s.aircraftDetail = result.aircraftDetail
	s.airbaseDetail = result.airbaseDetail
	s.threatDetail = result.threatDetail
	s.eventDetail = result.eventDetail
	s.listFace = result.listFace
	return nil
}

func buildTextShellUI(
	workspace shellWorkspace,
	aircraftMap map[string]aircraftEntry,
	selectedAircraft string,
	airbaseMap map[string]airbaseEntry,
	selectedAirbase string,
	threatMap map[string]threatEntry,
	selectedThreat string,
	recentEvents []eventEntry,
	selectedEventID string,
	onSelect func(string),
	onBranch func(),
	onReset func(),
	onPlayPause func(),
	onAircraftSelect func(string),
	onAirbaseSelect func(string),
	onThreatSelect func(string),
	onEventSelect func(string),
) (shellUIResult, error) {
	face, err := loadToolbarFace(toolbarFontSize)
	if err != nil {
		return shellUIResult{}, err
	}
	listFace, err := loadToolbarFace(listFontSize)
	if err != nil {
		return shellUIResult{}, err
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
			widget.RowLayoutOpts.Spacing(columnSpacing),
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

	columnsContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Stretch([]bool{true, true, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(columnSpacing, 0),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)

	col1 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(columnSpacing),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	aircraftHeader := widget.NewText(
		widget.TextOpts.Text("Aircraft", listFace, toolbarTextColor),
	)
	col1.AddChild(aircraftHeader)

	entries := aircraftMapToSortedEntries(aircraftMap)
	aircraftContent := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(0),
		)),
	)
	aircraftButtons := make(map[string]*widget.Button, len(entries))
	for _, entry := range entries {
		btn := newAircraftButton(entry, entry.tailNumber == selectedAircraft, listFace, onAircraftSelect)
		aircraftButtons[entry.tailNumber] = btn
		aircraftContent.AddChild(btn)
	}

	aircraftScroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(aircraftContent),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle:     ebitenuiimage.NewNineSliceColor(listBgColor),
			Disabled: ebitenuiimage.NewNineSliceColor(listBgColor),
			Mask:     ebitenuiimage.NewNineSliceColor(listBgColor),
		}),
		widget.ScrollContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(150, 0),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	col1.AddChild(aircraftScroll)

	columnsContainer.AddChild(col1)

	col2 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(columnSpacing),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	airbaseHeader := widget.NewText(
		widget.TextOpts.Text("Airbases", listFace, toolbarTextColor),
	)
	col2.AddChild(airbaseHeader)

	airbaseEntries := airbaseMapToSortedEntries(airbaseMap)
	airbaseContent := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(0),
		)),
	)
	airbaseButtons := make(map[string]*widget.Button, len(airbaseEntries))
	for _, entry := range airbaseEntries {
		btn := newAirbaseButton(entry, entry.id == selectedAirbase, listFace, onAirbaseSelect)
		airbaseButtons[entry.id] = btn
		airbaseContent.AddChild(btn)
	}
	airbaseScroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(airbaseContent),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle:     ebitenuiimage.NewNineSliceColor(listBgColor),
			Disabled: ebitenuiimage.NewNineSliceColor(listBgColor),
			Mask:     ebitenuiimage.NewNineSliceColor(listBgColor),
		}),
		widget.ScrollContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(150, 0),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	col2.AddChild(airbaseScroll)

	threatHeader := widget.NewText(
		widget.TextOpts.Text("Threats", listFace, toolbarTextColor),
	)
	col2.AddChild(threatHeader)

	threatEntries := threatMapToSortedEntries(threatMap)
	threatContent := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(0),
		)),
	)
	threatButtons := make(map[string]*widget.Button, len(threatEntries))
	for _, entry := range threatEntries {
		btn := newThreatButton(entry, entry.id == selectedThreat, listFace, onThreatSelect)
		threatButtons[entry.id] = btn
		threatContent.AddChild(btn)
	}
	threatScroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(threatContent),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle:     ebitenuiimage.NewNineSliceColor(listBgColor),
			Disabled: ebitenuiimage.NewNineSliceColor(listBgColor),
			Mask:     ebitenuiimage.NewNineSliceColor(listBgColor),
		}),
		widget.ScrollContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(150, 0),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	col2.AddChild(threatScroll)

	eventHeader := widget.NewText(
		widget.TextOpts.Text("Events", listFace, toolbarTextColor),
	)
	col2.AddChild(eventHeader)

	eventContent := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(0),
		)),
	)
	eventButtons := make([]*widget.Button, 0, len(recentEvents))
	for _, entry := range recentEvents {
		btn := newEventButton(entry, entry.id == selectedEventID, listFace, onEventSelect)
		eventButtons = append(eventButtons, btn)
		eventContent.AddChild(btn)
	}
	eventScroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(eventContent),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle:     ebitenuiimage.NewNineSliceColor(listBgColor),
			Disabled: ebitenuiimage.NewNineSliceColor(listBgColor),
			Mask:     ebitenuiimage.NewNineSliceColor(listBgColor),
		}),
		widget.ScrollContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(150, 0),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true, MaxHeight: eventsScrollMaxH}),
		),
	)
	col2.AddChild(eventScroll)

	columnsContainer.AddChild(col2)

	col3 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(columnSpacing),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	aircraftDetailPanel, aircraftDetailLabel := newDetailPanel(
		"Selected Aircraft",
		formatAircraftDetail(aircraftMap, selectedAircraft),
		listFace,
	)
	col3.AddChild(aircraftDetailPanel)

	airbaseDetailPanel, airbaseDetailLabel := newDetailPanel(
		"Selected Airbase",
		formatAirbaseDetail(airbaseMap, selectedAirbase),
		listFace,
	)
	col3.AddChild(airbaseDetailPanel)

	threatDetailPanel, threatDetailLabel := newDetailPanel(
		"Selected Threat",
		formatThreatDetail(threatMap, selectedThreat),
		listFace,
	)
	col3.AddChild(threatDetailPanel)

	eventDetailPanel, eventDetailLabel := newDetailPanel(
		"Selected Event",
		formatEventDetail(recentEvents, selectedEventID),
		listFace,
	)
	col3.AddChild(eventDetailPanel)

	columnsContainer.AddChild(col3)

	tabBody.AddChild(columnsContainer)

	root.AddChild(toolbar)
	viewport.AddChild(tabStrip)
	viewport.AddChild(tabBody)
	root.AddChild(viewport)

	return shellUIResult{
		ui:              &ebitenui.UI{Container: root},
		seedInput:       seedInput,
		infoLabel:       infoLabel,
		aircraftContent: aircraftContent,
		aircraftButtons: aircraftButtons,
		airbaseContent:  airbaseContent,
		airbaseButtons:  airbaseButtons,
		threatContent:   threatContent,
		threatButtons:   threatButtons,
		eventContent:    eventContent,
		eventButtons:    eventButtons,
		aircraftDetail:  aircraftDetailLabel,
		airbaseDetail:   airbaseDetailLabel,
		threatDetail:    threatDetailLabel,
		eventDetail:     eventDetailLabel,
		listFace:        listFace,
	}, nil
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

func newDetailPanel(title string, body string, face *textv2.Face) (*widget.Container, *widget.Text) {
	panel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ebitenuiimage.NewNineSliceColor(listBgColor)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(detailPanelSpacing),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Left:   detailPanelPaddingX,
				Right:  detailPanelPaddingX,
				Top:    detailPanelPaddingY,
				Bottom: detailPanelPaddingY,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(0, detailPanelMinH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)

	header := widget.NewText(
		widget.TextOpts.Text(title, face, toolbarTextColor),
	)
	label := widget.NewText(
		widget.TextOpts.Text(body, face, listUnselectedTextColor),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)

	panel.AddChild(header)
	panel.AddChild(label)

	return panel, label
}

func formatAircraftDetail(m map[string]aircraftEntry, selected string) string {
	if selected == "" {
		return "No aircraft selected"
	}
	entry, ok := m[selected]
	if !ok {
		return "No aircraft selected"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "TailNumber: %s\n", entry.tailNumber)
	fmt.Fprintf(&b, "State:      %s\n", entry.state)
	fmt.Fprintf(&b, "Position:   (%.1f, %.1f)\n", entry.posX, entry.posY)
	if len(entry.needs) == 0 {
		b.WriteString("Needs:      none")
	} else {
		b.WriteString("Needs:")
		for _, n := range entry.needs {
			fmt.Fprintf(&b, "\n  - %s  severity:%d  cap:%s  blocking:%v", n.Type, n.Severity, n.RequiredCapability, n.Blocking)
		}
	}
	return b.String()
}

func formatAirbaseDetail(m map[string]airbaseEntry, selected string) string {
	if selected == "" {
		return "No airbase selected"
	}
	entry, ok := m[selected]
	if !ok {
		return "No airbase selected"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "ID:       %s\n", entry.id)
	fmt.Fprintf(&b, "Region:   %s\n", entry.region)
	fmt.Fprintf(&b, "Aircraft: %d", entry.aircraftCount)
	if len(entry.capabilities) == 0 {
		b.WriteString("\nCapabilities: none")
		return b.String()
	}
	keys := make([]string, 0, len(entry.capabilities))
	for capability := range entry.capabilities {
		keys = append(keys, capability)
	}
	sort.Strings(keys)
	b.WriteString("\nCapabilities:")
	for _, capability := range keys {
		fmt.Fprintf(&b, "\n  - %s  recovery:%d", capability, entry.capabilities[capability].RecoveryMultiplierPermille)
	}
	return b.String()
}

func formatThreatDetail(m map[string]threatEntry, selected string) string {
	if selected == "" {
		return "No threat selected"
	}
	entry, ok := m[selected]
	if !ok {
		return "No threat selected"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "ID:      %s\n", entry.id)
	fmt.Fprintf(&b, "Engaged: %d", entry.engagedCount)
	return b.String()
}

func formatEventDetail(events []eventEntry, selected string) string {
	if selected == "" {
		return "No event selected"
	}
	for _, entry := range events {
		if entry.id == selected {
			return entry.detail
		}
	}
	return "No event selected"
}

func aircraftButtonLabel(entry aircraftEntry) string {
	return fmt.Sprintf("%-16s  Needs: %d  State: %s", entry.tailNumber, entry.needCount, entry.state)
}

func aircraftButtonImage(selected bool) *widget.ButtonImage {
	if selected {
		return &widget.ButtonImage{
			Idle:    ebitenuiimage.NewNineSliceColor(listSelectedBgColor),
			Hover:   ebitenuiimage.NewNineSliceColor(listSelectedBgColor),
			Pressed: ebitenuiimage.NewNineSliceColor(listSelectedBgColor),
		}
	}
	return &widget.ButtonImage{
		Idle:    ebitenuiimage.NewNineSliceColor(listBgColor),
		Hover:   ebitenuiimage.NewNineSliceColor(listFocusedBgColor),
		Pressed: ebitenuiimage.NewNineSliceColor(listSelectingBgColor),
	}
}

func aircraftButtonTextColor(selected bool) *widget.ButtonTextColor {
	if selected {
		return &widget.ButtonTextColor{
			Idle:    listSelectedTextColor,
			Hover:   listSelectedTextColor,
			Pressed: listSelectedTextColor,
		}
	}
	return &widget.ButtonTextColor{
		Idle:    listUnselectedTextColor,
		Hover:   listUnselectedTextColor,
		Pressed: listUnselectedTextColor,
	}
}

func newAircraftButton(entry aircraftEntry, selected bool, face *textv2.Face, onSelect func(string)) *widget.Button {
	label := aircraftButtonLabel(entry)
	tailNumber := entry.tailNumber
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(aircraftButtonImage(selected)),
		widget.ButtonOpts.Text(label, face, aircraftButtonTextColor(selected)),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   listEntryPaddingX,
			Right:  listEntryPaddingX,
			Top:    listEntryPaddingY,
			Bottom: listEntryPaddingY,
		}),
		widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			onSelect(tailNumber)
		}),
	)
}

func airbaseButtonLabel(entry airbaseEntry) string {
	short := entry.id
	if len(short) > 8 {
		short = short[:8]
	}
	return fmt.Sprintf("%-12s (%s)  Aircraft: %d", entry.region, short, entry.aircraftCount)
}

func newAirbaseButton(entry airbaseEntry, selected bool, face *textv2.Face, onSelect func(string)) *widget.Button {
	label := airbaseButtonLabel(entry)
	id := entry.id
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(aircraftButtonImage(selected)),
		widget.ButtonOpts.Text(label, face, aircraftButtonTextColor(selected)),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   listEntryPaddingX,
			Right:  listEntryPaddingX,
			Top:    listEntryPaddingY,
			Bottom: listEntryPaddingY,
		}),
		widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			onSelect(id)
		}),
	)
}

func threatButtonLabel(entry threatEntry) string {
	short := entry.id
	if len(short) > 8 {
		short = short[:8]
	}
	return fmt.Sprintf("%-8s  Engaged: %d", short, entry.engagedCount)
}

func formatRecentEvent(evt services.Event) (eventEntry, bool) {
	switch e := evt.(type) {
	case services.AircraftStateChangeEvent:
		return eventEntry{
			id:     fmt.Sprintf("%s:%s:%d:%s:%s", e.Type, e.TailNumber, e.Tick, e.OldState, e.NewState),
			text:   fmt.Sprintf("%s %s %s -> %s", e.Timestamp.Format("15:04:05"), shortID(e.TailNumber), e.OldState, e.NewState),
			detail: fmt.Sprintf("Type: %s\nTail: %s\nTick: %d\nState: %s -> %s\nTime: %s", e.Type, e.TailNumber, e.Tick, e.OldState, e.NewState, e.Timestamp.Format(time.RFC3339)),
		}, true
	case services.LandingAssignmentEvent:
		return eventEntry{
			id:     fmt.Sprintf("%s:%s:%d:%s:%s", e.Type, e.TailNumber, e.Tick, e.BaseID, e.Source),
			text:   fmt.Sprintf("%s assign %s -> %s (%s)", e.Timestamp.Format("15:04:05"), shortID(e.TailNumber), shortID(e.BaseID), string(e.Source)),
			detail: fmt.Sprintf("Type: %s\nTail: %s\nBase: %s\nSource: %s\nTick: %d\nTime: %s", e.Type, e.TailNumber, e.BaseID, e.Source, e.Tick, e.Timestamp.Format(time.RFC3339)),
		}, true
	case services.ThreatSpawnedEvent:
		return eventEntry{
			id:     fmt.Sprintf("%s:%s:%d", e.Type, e.Threat.ID, e.Tick),
			text:   fmt.Sprintf("%s threat spawned %s", e.Timestamp.Format("15:04:05"), shortID(e.Threat.ID)),
			detail: fmt.Sprintf("Type: %s\nThreat: %s\nTick: %d\nTime: %s", e.Type, e.Threat.ID, e.Tick, e.Timestamp.Format(time.RFC3339)),
		}, true
	case services.ThreatTargetedEvent:
		return eventEntry{
			id:     fmt.Sprintf("%s:%s:%d:%s", e.Type, e.Threat.ID, e.Tick, e.TailNumber),
			text:   fmt.Sprintf("%s threat %s targeted by %s", e.Timestamp.Format("15:04:05"), shortID(e.Threat.ID), shortID(e.TailNumber)),
			detail: fmt.Sprintf("Type: %s\nThreat: %s\nTail: %s\nTick: %d\nTime: %s", e.Type, e.Threat.ID, e.TailNumber, e.Tick, e.Timestamp.Format(time.RFC3339)),
		}, true
	case services.ThreatDespawnedEvent:
		return eventEntry{
			id:     fmt.Sprintf("%s:%s:%d", e.Type, e.Threat.ID, e.Tick),
			text:   fmt.Sprintf("%s threat despawned %s", e.Timestamp.Format("15:04:05"), shortID(e.Threat.ID)),
			detail: fmt.Sprintf("Type: %s\nThreat: %s\nTick: %d\nTime: %s", e.Type, e.Threat.ID, e.Tick, e.Timestamp.Format(time.RFC3339)),
		}, true
	case services.SimulationEndedEvent:
		return eventEntry{
			id:     fmt.Sprintf("%s:%d", e.Type, e.Tick),
			text:   fmt.Sprintf("%s simulation ended", e.Timestamp.Format("15:04:05")),
			detail: fmt.Sprintf("Type: %s\nTick: %d\nTime: %s", e.Type, e.Tick, e.Timestamp.Format(time.RFC3339)),
		}, true
	case services.SimulationClosedEvent:
		return eventEntry{
			id:     fmt.Sprintf("%s:%d:%s", e.Type, e.Tick, e.Reason),
			text:   fmt.Sprintf("%s simulation closed (%s)", e.Timestamp.Format("15:04:05"), e.Reason),
			detail: fmt.Sprintf("Type: %s\nReason: %s\nTick: %d\nTime: %s", e.Type, e.Reason, e.Tick, e.Timestamp.Format(time.RFC3339)),
		}, true
	case services.BranchCreatedEvent:
		return eventEntry{
			id:     fmt.Sprintf("%s:%s:%d", e.Type, e.BranchID, e.Tick),
			text:   fmt.Sprintf("%s branch created %s", e.SplitTimestamp.Format("15:04:05"), shortID(e.BranchID)),
			detail: fmt.Sprintf("Type: %s\nBranch: %s\nParent: %s\nTick: %d\nTime: %s", e.Type, e.BranchID, e.ParentID, e.Tick, e.SplitTimestamp.Format(time.RFC3339)),
		}, true
	default:
		return eventEntry{}, false
	}
}

func shortID(value string) string {
	if len(value) <= 8 {
		return value
	}
	return value[:8]
}

func newThreatButton(entry threatEntry, selected bool, face *textv2.Face, onSelect func(string)) *widget.Button {
	label := threatButtonLabel(entry)
	id := entry.id
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(aircraftButtonImage(selected)),
		widget.ButtonOpts.Text(label, face, aircraftButtonTextColor(selected)),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   listEntryPaddingX,
			Right:  listEntryPaddingX,
			Top:    listEntryPaddingY,
			Bottom: listEntryPaddingY,
		}),
		widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			onSelect(id)
		}),
	)
}

func newEventButton(entry eventEntry, selected bool, face *textv2.Face, onSelect func(string)) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(aircraftButtonImage(selected)),
		widget.ButtonOpts.Text(entry.text, face, aircraftButtonTextColor(selected)),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   listEntryPaddingX,
			Right:  listEntryPaddingX,
			Top:    listEntryPaddingY,
			Bottom: listEntryPaddingY,
		}),
		widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			onSelect(entry.id)
		}),
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

func (s *textShell) refreshInfoLabel(info services.SimulationInfo) {
	if s.infoLabel == nil {
		return
	}
	ts := "—"
	if !info.Timestamp.IsZero() {
		ts = info.Timestamp.Format("15:04:05")
	}
	s.workspace.activeSummary.tick = info.Tick
	s.workspace.activeSummary.timestamp = ts
	s.workspace.activeSummary.running = info.Running
	s.workspace.activeSummary.paused = info.Paused
	s.workspace.activeSummary.aircraftCount = len(s.aircraftMap)
	s.workspace.activeSummary.airbaseCount = len(s.airbaseMap)
	s.workspace.activeSummary.threatCount = len(s.threatMap)
	s.infoLabel.Label = fmt.Sprintf(
		"Tick: %d    Time: %s    Aircraft: %d    Airbases: %d    Threats: %d",
		info.Tick, ts,
		len(s.aircraftMap),
		len(s.airbaseMap),
		len(s.threatMap),
	)
}

func (s *textShell) refreshSelectedAircraftFromService() {
	if s.selectedAircraft == "" {
		s.lastAircraftPollTabID = s.workspace.activeTabID
		s.lastAircraftPollTail = ""
		return
	}
	aircrafts, err := s.service.Aircrafts(s.workspace.activeTabID)
	if err != nil {
		return
	}
	for _, ac := range aircrafts {
		if ac.TailNumber != s.selectedAircraft {
			continue
		}
		s.aircraftMap[ac.TailNumber] = aircraftEntry{
			tailNumber: ac.TailNumber,
			needCount:  len(ac.Needs),
			state:      ac.State,
			posX:       ac.Position.X,
			posY:       ac.Position.Y,
			needs:      ac.Needs,
		}
		s.aircraftDirty = true
		break
	}

	s.lastAircraftPollTabID = s.workspace.activeTabID
	s.lastAircraftPollTail = s.selectedAircraft
}

func (s *textShell) seedAircraftMap() {
	s.aircraftMap = make(map[string]aircraftEntry)
	aircrafts, err := s.service.Aircrafts(s.workspace.activeTabID)
	if err != nil {
		return
	}
	for _, ac := range aircrafts {
		s.aircraftMap[ac.TailNumber] = aircraftEntry{
			tailNumber: ac.TailNumber,
			needCount:  len(ac.Needs),
			state:      ac.State,
			posX:       ac.Position.X,
			posY:       ac.Position.Y,
			needs:      ac.Needs,
		}
	}
}

func (s *textShell) seedAirbaseMap() {
	s.airbaseMap = make(map[string]airbaseEntry)
	airbases, err := s.service.Airbases(s.workspace.activeTabID)
	if err != nil {
		return
	}
	aircraftsByBase := make(map[string]int)
	aircrafts, err := s.service.Aircrafts(s.workspace.activeTabID)
	if err == nil {
		for _, ac := range aircrafts {
			if ac.AssignedTo != nil {
				aircraftsByBase[*ac.AssignedTo]++
			}
		}
	}
	for _, ab := range airbases {
		s.airbaseMap[ab.ID] = airbaseEntry{
			id:            ab.ID,
			region:        ab.Region,
			aircraftCount: aircraftsByBase[ab.ID],
			capabilities:  ab.Capabilities,
		}
	}
}

func (s *textShell) seedThreatMap() {
	s.threatMap = make(map[string]threatEntry)
	s.threatEngagedByThreatID = make(map[string]map[string]bool)
	threats, err := s.service.Threats(s.workspace.activeTabID)
	if err != nil {
		return
	}
	for _, t := range threats {
		s.threatMap[t.ID] = threatEntry{
			id:           t.ID,
			engagedCount: 0,
		}
	}
}

func (s *textShell) drainEvents() {
	for {
		select {
		case evt, ok := <-s.eventCh:
			if !ok {
				return
			}
			if evt.EventSimulationID() != s.workspace.activeTabID {
				continue
			}
			s.appendRecentEvent(evt)
			switch e := evt.(type) {
			case services.AircraftStateChangeEvent:
				s.aircraftMap[e.Aircraft.TailNumber] = aircraftEntry{
					tailNumber: e.Aircraft.TailNumber,
					needCount:  len(e.Aircraft.Needs),
					state:      e.NewState,
					posX:       e.Aircraft.Position.X,
					posY:       e.Aircraft.Position.Y,
					needs:      e.Aircraft.Needs,
				}
				s.aircraftDirty = true
				s.airbaseDirty = true
			case services.AllAircraftPositionsEvent:
				for _, snap := range e.Positions {
					if existing, ok := s.aircraftMap[snap.TailNumber]; ok {
						existing.posX = snap.Position.X
						existing.posY = snap.Position.Y
						existing.state = snap.State
						if snap.TailNumber != s.selectedAircraft {
							existing.needCount = len(snap.Needs)
							existing.needs = snap.Needs
						}
						s.aircraftMap[snap.TailNumber] = existing
					}
				}
				s.aircraftDirty = true
			case services.ThreatSpawnedEvent:
				s.threatMap[e.Threat.ID] = threatEntry{
					id:           e.Threat.ID,
					engagedCount: 0,
				}
				s.threatDirty = true
			case services.ThreatTargetedEvent:
				if s.threatEngagedByThreatID[e.Threat.ID] == nil {
					s.threatEngagedByThreatID[e.Threat.ID] = make(map[string]bool)
				}
				s.threatEngagedByThreatID[e.Threat.ID][e.TailNumber] = true
				if te, ok := s.threatMap[e.Threat.ID]; ok {
					te.engagedCount = len(s.threatEngagedByThreatID[e.Threat.ID])
					s.threatMap[e.Threat.ID] = te
				}
				s.threatDirty = true
			case services.ThreatDespawnedEvent:
				delete(s.threatMap, e.Threat.ID)
				delete(s.threatEngagedByThreatID, e.Threat.ID)
				s.threatDirty = true
			}
		default:
			return
		}
	}
}

func aircraftMapToSortedEntries(m map[string]aircraftEntry) []aircraftEntry {
	entries := make([]aircraftEntry, 0, len(m))
	for _, e := range m {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].tailNumber < entries[j].tailNumber
	})
	return entries
}

func airbaseMapToSortedEntries(m map[string]airbaseEntry) []airbaseEntry {
	entries := make([]airbaseEntry, 0, len(m))
	for _, e := range m {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].region < entries[j].region
	})
	return entries
}

func threatMapToSortedEntries(m map[string]threatEntry) []threatEntry {
	entries := make([]threatEntry, 0, len(m))
	for _, e := range m {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].id < entries[j].id
	})
	return entries
}

func (s *textShell) refreshAircraftList() {
	if s.aircraftContent == nil || !s.aircraftDirty {
		return
	}
	s.aircraftDirty = false

	for tailNumber, entry := range s.aircraftMap {
		btn, exists := s.aircraftButtons[tailNumber]
		if exists {
			btn.SetText(aircraftButtonLabel(entry))
		} else {
			selected := tailNumber == s.selectedAircraft
			newBtn := newAircraftButton(entry, selected, s.listFace, func(tn string) {
				s.selectedAircraft = tn
				s.updateAircraftSelection()
			})
			s.aircraftButtons[tailNumber] = newBtn
			s.aircraftContent.AddChild(newBtn)
		}
	}

	for tailNumber, btn := range s.aircraftButtons {
		if _, exists := s.aircraftMap[tailNumber]; !exists {
			s.aircraftContent.RemoveChild(btn)
			delete(s.aircraftButtons, tailNumber)
		}
	}
}

func (s *textShell) updateAircraftSelection() {
	for tailNumber, btn := range s.aircraftButtons {
		selected := tailNumber == s.selectedAircraft
		btn.SetImage(aircraftButtonImage(selected))
	}
}

func (s *textShell) refreshAirbaseList() {
	if s.airbaseContent == nil || !s.airbaseDirty {
		return
	}
	s.airbaseDirty = false

	aircraftsByBase := make(map[string]int)
	aircrafts, err := s.service.Aircrafts(s.workspace.activeTabID)
	if err == nil {
		for _, ac := range aircrafts {
			if ac.AssignedTo != nil {
				aircraftsByBase[*ac.AssignedTo]++
			}
		}
	}
	for id, entry := range s.airbaseMap {
		entry.aircraftCount = aircraftsByBase[id]
		s.airbaseMap[id] = entry
		btn, exists := s.airbaseButtons[id]
		if exists {
			btn.SetText(airbaseButtonLabel(entry))
		} else {
			selected := id == s.selectedAirbase
			newBtn := newAirbaseButton(entry, selected, s.listFace, func(abID string) {
				s.selectedAirbase = abID
				s.updateAirbaseSelection()
			})
			s.airbaseButtons[id] = newBtn
			s.airbaseContent.AddChild(newBtn)
		}
	}

	for id, btn := range s.airbaseButtons {
		if _, exists := s.airbaseMap[id]; !exists {
			s.airbaseContent.RemoveChild(btn)
			delete(s.airbaseButtons, id)
		}
	}
}

func (s *textShell) updateAirbaseSelection() {
	for id, btn := range s.airbaseButtons {
		selected := id == s.selectedAirbase
		btn.SetImage(aircraftButtonImage(selected))
	}
}

func (s *textShell) refreshThreatList() {
	if s.threatContent == nil || !s.threatDirty {
		return
	}
	s.threatDirty = false

	for id, entry := range s.threatMap {
		btn, exists := s.threatButtons[id]
		if exists {
			btn.SetText(threatButtonLabel(entry))
		} else {
			selected := id == s.selectedThreat
			newBtn := newThreatButton(entry, selected, s.listFace, func(tID string) {
				s.selectedThreat = tID
				s.updateThreatSelection()
			})
			s.threatButtons[id] = newBtn
			s.threatContent.AddChild(newBtn)
		}
	}

	for id, btn := range s.threatButtons {
		if _, exists := s.threatMap[id]; !exists {
			s.threatContent.RemoveChild(btn)
			delete(s.threatButtons, id)
		}
	}
}

func (s *textShell) appendRecentEvent(evt services.Event) {
	entry, ok := formatRecentEvent(evt)
	if !ok {
		return
	}
	s.recentEvents = append([]eventEntry{entry}, s.recentEvents...)
	if len(s.recentEvents) > recentEventsLimit {
		s.recentEvents = s.recentEvents[:recentEventsLimit]
	}
	if s.selectedEventID == "" {
		s.selectedEventID = entry.id
	}
	selectedStillExists := false
	for _, existing := range s.recentEvents {
		if existing.id == s.selectedEventID {
			selectedStillExists = true
			break
		}
	}
	if !selectedStillExists {
		s.selectedEventID = ""
	}
	s.eventsDirty = true
}

func (s *textShell) refreshRecentEvents() {
	if s.eventContent == nil || !s.eventsDirty {
		return
	}
	s.eventsDirty = false

	for len(s.eventButtons) < len(s.recentEvents) {
		idx := len(s.eventButtons)
		entry := s.recentEvents[idx]
		btn := newEventButton(entry, entry.id == s.selectedEventID, s.listFace, func(eventID string) {
			s.selectedEventID = eventID
			s.updateEventSelection()
		})
		s.eventButtons = append(s.eventButtons, btn)
		s.eventContent.AddChild(btn)
	}
	for i, entry := range s.recentEvents {
		s.eventButtons[i].SetText(entry.text)
		s.eventButtons[i].SetImage(aircraftButtonImage(entry.id == s.selectedEventID))
	}
	for i := len(s.recentEvents); i < len(s.eventButtons); i++ {
		s.eventButtons[i].SetText("")
		s.eventButtons[i].SetImage(aircraftButtonImage(false))
	}
	if s.ui != nil && s.ui.Container != nil {
		s.ui.Container.RequestRelayout()
	}
}

func (s *textShell) updateEventSelection() {
	for i, btn := range s.eventButtons {
		selected := i < len(s.recentEvents) && s.recentEvents[i].id == s.selectedEventID
		btn.SetImage(aircraftButtonImage(selected))
	}
}

func (s *textShell) refreshDetailLabels() {
	if s.aircraftDetail != nil {
		label := formatAircraftDetail(s.aircraftMap, s.selectedAircraft)
		if s.aircraftDetail.Label != label {
			s.aircraftDetail.Label = label
			if s.ui != nil && s.ui.Container != nil {
				s.ui.Container.RequestRelayout()
			}
		}
	}
	if s.airbaseDetail != nil {
		label := formatAirbaseDetail(s.airbaseMap, s.selectedAirbase)
		if s.airbaseDetail.Label != label {
			s.airbaseDetail.Label = label
			if s.ui != nil && s.ui.Container != nil {
				s.ui.Container.RequestRelayout()
			}
		}
	}
	if s.threatDetail != nil {
		label := formatThreatDetail(s.threatMap, s.selectedThreat)
		if s.threatDetail.Label != label {
			s.threatDetail.Label = label
			if s.ui != nil && s.ui.Container != nil {
				s.ui.Container.RequestRelayout()
			}
		}
	}
	if s.eventDetail != nil {
		label := formatEventDetail(s.recentEvents, s.selectedEventID)
		if s.eventDetail.Label != label {
			s.eventDetail.Label = label
			if s.ui != nil && s.ui.Container != nil {
				s.ui.Container.RequestRelayout()
			}
		}
	}
}

func (s *textShell) updateThreatSelection() {
	for id, btn := range s.threatButtons {
		selected := id == s.selectedThreat
		btn.SetImage(aircraftButtonImage(selected))
	}
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
