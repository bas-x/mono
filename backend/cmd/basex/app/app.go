package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	textv2 "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/bas-x/basex/services"
)

type RuntimeConfig struct {
	Config  Config
	Service *services.SimulationService
}

type Runtime struct {
	config       Config
	session      *Session
	state        *State
	viewport     Viewport
	ui           *ebitenui.UI
	startButton  *widget.Button
	statusLabel  *widget.Label
	timeLabel    *widget.Label
	airbaseRows  []*widget.Button
	aircraftRows []*widget.Button
	uiFace       *textv2.Face
	uiDirty      bool
	draggingMap  bool
	lastMouseX   int
	lastMouseY   int
	qaRan        bool
	closed       bool
}

func New(cfg RuntimeConfig) (*Runtime, error) {
	if cfg.Service == nil {
		return nil, fmt.Errorf("service is required")
	}
	faceSource, err := textv2.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		return nil, fmt.Errorf("create ui font: %w", err)
	}
	face := textv2.Face(&textv2.GoTextFace{Source: faceSource, Size: 16})
	return &Runtime{
		config:   cfg.Config,
		session:  NewSession(cfg.Service),
		state:    NewState(),
		viewport: DefaultViewport(),
		uiFace:   &face,
		uiDirty:  true,
	}, nil
}

func (r *Runtime) Close() {
	if r.closed {
		return
	}
	r.closed = true
	r.session.Close()
}

func (r *Runtime) Update() error {
	events := r.session.DrainEvents()
	refreshUI := false
	for _, event := range events {
		r.state.ApplyEvent(event)
		if eventRequiresUIRefresh(event) {
			refreshUI = true
		}
	}
	if refreshUI {
		r.uiDirty = true
	}
	if !r.qaRan && len(r.config.QASteps) > 0 {
		if err := r.runQAScript(); err != nil {
			r.state.SetError(err)
			return err
		}
		r.qaRan = true
		if err := r.writeArtifacts(); err != nil {
			return err
		}
		if r.config.ExitAfterQA {
			return ebiten.Termination
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	r.handleInteractiveInput()
	r.handleMouseInput()
	if err := r.syncUI(); err != nil {
		return err
	}
	if r.ui != nil {
		r.refreshDynamicUI()
		r.ui.Update()
	}
	return nil
}

func (r *Runtime) Draw(screen *ebiten.Image) {
	r.drawToScreen(screen)
}

func (r *Runtime) handleInteractiveInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		r.state.Error = ""
		if _, _, _, _, err := r.createSimulation(); err != nil {
			r.state.SetError(err)
		}
		r.uiDirty = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		r.state.Error = ""
		if err := r.startSimulation(); err != nil {
			r.state.SetError(err)
		}
		r.uiDirty = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		r.state.Error = ""
		if err := r.togglePause(); err != nil {
			r.state.SetError(err)
		}
		r.uiDirty = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		r.state.Error = ""
		if err := r.resetSimulation(); err != nil {
			r.state.SetError(err)
		}
		r.uiDirty = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		r.selectNextAirbase()
		r.uiDirty = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		r.selectNextAircraft()
		r.uiDirty = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		r.state.Error = ""
		if err := r.refreshState(); err != nil {
			r.state.SetError(err)
		}
		r.uiDirty = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		r.viewport = r.viewport.Pan(r.viewport.Width*0.03, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		r.viewport = r.viewport.Pan(-r.viewport.Width*0.03, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		r.viewport = r.viewport.Pan(0, -r.viewport.Height*0.03)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		r.viewport = r.viewport.Pan(0, r.viewport.Height*0.03)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		r.viewport = r.viewport.Zoom(1.2)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		r.viewport = r.viewport.Zoom(1 / 1.2)
	}
}

func (r *Runtime) handleMouseInput() {
	mx, my := ebiten.CursorPosition()
	mapRect := mapInnerRect(r.config.WindowWidth, r.config.WindowHeight)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if image.Pt(mx, my).In(mapRect) {
			if r.selectMapEntityAt(mx, my) {
				r.uiDirty = true
			} else {
				r.draggingMap = true
				r.lastMouseX = mx
				r.lastMouseY = my
			}
		}
	}
	if r.draggingMap && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		dx := mx - r.lastMouseX
		dy := my - r.lastMouseY
		if dx != 0 || dy != 0 {
			proj := newProjection(image.Rect(0, 0, max(mapRect.Dx(), 1), max(mapRect.Dy(), 1)), r.viewport)
			if proj.scale > 0 {
				r.viewport = r.viewport.Pan(-float64(dx)/proj.scale, -float64(dy)/proj.scale)
			}
			r.lastMouseX = mx
			r.lastMouseY = my
		}
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		r.draggingMap = false
	}
	if image.Pt(mx, my).In(mapRect) {
		_, wheelY := ebiten.Wheel()
		if wheelY > 0 {
			r.zoomAtCursor(mx, my, 1.15)
		} else if wheelY < 0 {
			r.zoomAtCursor(mx, my, 1/1.15)
		}
	}
}

func (r *Runtime) Layout(outsideWidth, outsideHeight int) (int, int) {
	return r.config.WindowWidth, r.config.WindowHeight
}

func (r *Runtime) runQAScript() error {
	for _, step := range r.config.QASteps {
		switch {
		case step == "boot":
			continue
		case step == "create":
			info, airbases, aircraft, threats, err := r.createSimulation()
			if err != nil {
				return err
			}
			_ = info
			_ = airbases
			_ = aircraft
			_ = threats
		case step == "start":
			if err := r.startSimulation(); err != nil {
				return err
			}
		case step == "pause":
			if err := r.pauseSimulation(); err != nil {
				return err
			}
		case step == "resume":
			if err := r.resumeSimulation(); err != nil {
				return err
			}
		case step == "reset":
			if err := r.resetSimulation(); err != nil {
				return err
			}
		case strings.HasPrefix(step, "select-airbase-"):
			var index int
			_, err := fmt.Sscanf(step, "select-airbase-%d", &index)
			if err != nil {
				return err
			}
			r.state.SelectAirbaseIndex(index)
			if index >= 0 && index < len(r.state.Airbases) {
				r.viewport = r.viewport.FocusAirbase(r.state.Airbases[index])
			}
		case strings.HasPrefix(step, "select-aircraft-"):
			var index int
			_, err := fmt.Sscanf(step, "select-aircraft-%d", &index)
			if err != nil {
				return err
			}
			r.state.SelectAircraftIndex(index)
		case step == "pan-right":
			r.viewport = r.viewport.Pan(r.viewport.Width*0.1, 0)
		case step == "mouse-pan-right":
			r.viewport = r.viewport.Pan(-r.viewport.Width*0.1, 0)
		case step == "zoom-in":
			r.viewport = r.viewport.Zoom(1.25)
		case step == "mouse-zoom-in":
			r.viewport = r.viewport.ZoomAt(1.25, 0.5, 0.5)
		case strings.HasPrefix(step, "click-airbase-"):
			var index int
			_, err := fmt.Sscanf(step, "click-airbase-%d", &index)
			if err != nil {
				return err
			}
			r.selectAirbase(index)
		case strings.HasPrefix(step, "click-aircraft-"):
			var index int
			_, err := fmt.Sscanf(step, "click-aircraft-%d", &index)
			if err != nil {
				return err
			}
			r.selectAircraft(index)
		default:
			return fmt.Errorf("unsupported qa step %q", step)
		}
	}
	return nil
}

func (r *Runtime) createSimulation() (services.SimulationInfo, []services.Airbase, []services.Aircraft, []services.Threat, error) {
	info, airbases, aircraft, threats, err := r.session.Create(r.config.Seed)
	if err != nil {
		return services.SimulationInfo{}, nil, nil, nil, err
	}
	r.state.SetSnapshot(info, airbases, aircraft, threats)
	r.state.addEvent("simulation created", true)
	if len(airbases) > 0 {
		r.viewport = r.viewport.FocusAirbase(airbases[0])
	}
	return info, airbases, aircraft, threats, nil
}

func (r *Runtime) startSimulation() error {
	if r.state.Paused {
		return r.resumeSimulation()
	}
	simulationID, err := mustSimulationID(r.state)
	if err != nil {
		return err
	}
	if err := r.session.Start(simulationID); err != nil {
		return err
	}
	r.state.addEvent("simulation started", true)
	return r.refreshState()
}

func (r *Runtime) pauseSimulation() error {
	simulationID, err := mustSimulationID(r.state)
	if err != nil {
		return err
	}
	if err := r.session.Pause(simulationID); err != nil {
		return err
	}
	r.state.addEvent("simulation paused", true)
	return r.refreshState()
}

func (r *Runtime) resumeSimulation() error {
	simulationID, err := mustSimulationID(r.state)
	if err != nil {
		return err
	}
	if err := r.session.Resume(simulationID); err != nil {
		return err
	}
	r.state.addEvent("simulation resumed", true)
	return r.refreshState()
}

func (r *Runtime) togglePause() error {
	if r.state.Paused {
		return r.resumeSimulation()
	}
	return r.pauseSimulation()
}

func (r *Runtime) resetSimulation() error {
	simulationID, err := mustSimulationID(r.state)
	if err != nil {
		return err
	}
	if err := r.session.Reset(simulationID); err != nil {
		return err
	}
	r.state.Clear()
	r.state.addEvent("simulation reset", true)
	r.viewport = DefaultViewport()
	return nil
}

func (r *Runtime) refreshState() error {
	simulationID, err := mustSimulationID(r.state)
	if err != nil {
		return err
	}
	info, airbases, aircraft, threats, err := r.session.Snapshot(simulationID)
	if err != nil {
		return err
	}
	r.state.SetSnapshot(info, airbases, aircraft, threats)
	r.state.addEvent("state refreshed", true)
	return nil
}

func (r *Runtime) selectNextAirbase() {
	if len(r.state.Airbases) == 0 {
		return
	}
	current := 0
	for i := range r.state.Airbases {
		if r.state.Airbases[i].ID == r.state.SelectedAirbaseID {
			current = i + 1
			break
		}
	}
	current %= len(r.state.Airbases)
	r.selectAirbase(current)
}

func (r *Runtime) selectNextAircraft() {
	if len(r.state.Aircraft) == 0 {
		return
	}
	current := 0
	for i := range r.state.Aircraft {
		if r.state.Aircraft[i].TailNumber == r.state.SelectedAircraft {
			current = i + 1
			break
		}
	}
	current %= len(r.state.Aircraft)
	r.selectAircraft(current)
}

func (r *Runtime) selectAirbase(index int) {
	if index < 0 || index >= len(r.state.Airbases) {
		return
	}
	r.state.SelectAirbaseIndex(index)
	r.state.addEvent("selected airbase "+shortID(r.state.Airbases[index].ID), true)
	r.viewport = r.viewport.FocusAirbase(r.state.Airbases[index])
}

func (r *Runtime) selectAircraft(index int) {
	if index < 0 || index >= len(r.state.Aircraft) {
		return
	}
	r.state.SelectAircraftIndex(index)
	r.state.addEvent("selected aircraft "+shortID(r.state.Aircraft[index].TailNumber), true)
}

func (r *Runtime) selectMapEntityAt(screenX, screenY int) bool {
	mapRect := mapInnerRect(r.config.WindowWidth, r.config.WindowHeight)
	surfacePoint := image.Pt(screenX-mapRect.Min.X, screenY-mapRect.Min.Y)
	proj := newProjection(image.Rect(0, 0, max(mapRect.Dx(), 1), max(mapRect.Dy(), 1)), r.viewport)
	for i, aircraft := range r.state.Aircraft {
		if aircraft.Position.X == 0 && aircraft.Position.Y == 0 {
			continue
		}
		pt := proj.projectServicePoint(aircraft.Position)
		if distanceSquared(surfacePoint, pt) <= 64 {
			r.selectAircraft(i)
			return true
		}
	}
	for i, base := range r.state.Airbases {
		pt := proj.projectServicePoint(base.Location)
		if distanceSquared(surfacePoint, pt) <= 100 {
			r.selectAirbase(i)
			return true
		}
	}
	return false
}

func (r *Runtime) zoomAtCursor(screenX, screenY int, scale float64) {
	mapRect := mapInnerRect(r.config.WindowWidth, r.config.WindowHeight)
	proj := newProjection(image.Rect(0, 0, max(mapRect.Dx(), 1), max(mapRect.Dy(), 1)), r.viewport)
	surfacePoint := image.Pt(screenX-mapRect.Min.X, screenY-mapRect.Min.Y)
	relX, relY := proj.relativePoint(surfacePoint)
	r.viewport = r.viewport.ZoomAt(scale, relX, relY)
}

func (r *Runtime) writeArtifacts() error {
	if err := r.syncUI(); err != nil {
		return err
	}
	if r.config.DumpStatePath != "" {
		if err := writeJSON(r.config.DumpStatePath, struct {
			*State
			Viewport Viewport `json:"viewport"`
		}{State: r.state, Viewport: r.viewport}); err != nil {
			return err
		}
	}
	if r.config.DumpEventsPath != "" {
		if err := writeJSON(r.config.DumpEventsPath, struct {
			Events []string `json:"events"`
		}{Events: append([]string(nil), r.state.EventLog...)}); err != nil {
			return err
		}
	}
	if r.config.ScreenshotOut != "" {
		frame, err := r.captureFrameImage(r.config.WindowWidth, r.config.WindowHeight)
		if err != nil {
			return err
		}
		if err := writePNG(r.config.ScreenshotOut, frame); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runtime) drawToScreen(screen *ebiten.Image) {
	bg := ebiten.NewImageFromImage(r.renderImage(screen.Bounds().Dx(), screen.Bounds().Dy()))
	screen.DrawImage(bg, nil)
	if r.ui != nil {
		r.ui.Draw(screen)
	}
}

func (r *Runtime) captureFrameImage(width, height int) (*image.RGBA, error) {
	screen := ebiten.NewImage(width, height)
	r.drawToScreen(screen)
	pix := make([]byte, 4*width*height)
	screen.ReadPixels(pix)
	out := image.NewRGBA(image.Rect(0, 0, width, height))
	copy(out.Pix, pix)
	return out, nil
}

func (r *Runtime) syncUI() error {
	if !r.uiDirty && r.ui != nil {
		return nil
	}
	ui, err := buildSidebarUI(r)
	if err != nil {
		return err
	}
	r.ui = ui
	r.uiDirty = false
	r.refreshDynamicUI()
	return nil
}

func (r *Runtime) refreshDynamicUI() {
	if r.statusLabel != nil {
		r.statusLabel.Label = fmt.Sprintf("Status: %s | Tick: %d", r.state.Status, r.state.Tick)
	}
	if r.startButton != nil {
		label := "Start simulation"
		if r.state.Paused {
			label = "Resume simulation"
		} else if r.state.Running {
			label = "Pause simulation"
		}
		r.startButton.SetText(label)
	}
	if r.timeLabel != nil {
		if r.state.CurrentTime.IsZero() {
			r.timeLabel.Label = "Time: n/a"
		} else {
			r.timeLabel.Label = fmt.Sprintf("Time: %s", r.state.CurrentTime.Format("2006-01-02 15:04:05"))
		}
	}
	for i, btn := range r.airbaseRows {
		if btn == nil {
			continue
		}
		text := ""
		selected := false
		if i < len(r.state.Airbases) {
			base := r.state.Airbases[i]
			text = truncateText(fmt.Sprintf("%d. %s (%s)", i+1, shortID(base.ID), base.Region), 28)
			selected = base.ID == r.state.SelectedAirbaseID
		}
		btn.SetText(text)
		btn.SetImage(listRowButtonImage(selected))
	}
	for i, btn := range r.aircraftRows {
		if btn == nil {
			continue
		}
		text := ""
		selected := false
		if i < len(r.state.Aircraft) {
			aircraft := r.state.Aircraft[i]
			text = truncateText(fmt.Sprintf("%d. %s | %s | needs:%d", i+1, shortID(aircraft.TailNumber), aircraft.State, len(aircraft.Needs)), 30)
			selected = aircraft.TailNumber == r.state.SelectedAircraft
		}
		btn.SetText(text)
		btn.SetImage(listRowButtonImage(selected))
	}
}

func eventRequiresUIRefresh(event services.Event) bool {
	return false
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func writePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func truncateText(text string, maxLen int) string {
	if maxLen <= 0 || len(text) <= maxLen {
		return text
	}
	if maxLen <= 1 {
		return text[:maxLen]
	}
	return text[:maxLen-1] + "…"
}
