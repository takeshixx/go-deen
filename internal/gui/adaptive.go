//go:build gui

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

const (
	compactSidebarBreakpoint         = 960
	compactStagesBreakpoint          = 780
	expandedSidebarWidth     float32 = 226
	collapsedSidebarWidth    float32 = 68
)

// sidebarShellLayout always reserves an icon rail. At wide sizes the user's
// preference controls whether labels are shown; at compact sizes the rail is
// collapsed automatically and can expand temporarily over the workspace.
type sidebarShellLayout struct {
	gui     *DeenGUI
	leading fyne.CanvasObject
	content fyne.CanvasObject
}

func (layout *sidebarShellLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	compact := size.Width < compactSidebarBreakpoint
	if layout.gui.compactSidebar != compact {
		layout.gui.setCompactSidebar(compact)
	}
	expanded := layout.gui.sidebarOpen
	if compact {
		expanded = layout.gui.sidebarDrawerOpen
	}
	layout.gui.setSidebarCollapsed(!expanded)
	setAdaptiveObjectVisible(layout.leading, true)

	leadingWidth := collapsedSidebarWidth
	if expanded {
		leadingWidth = expandedSidebarWidth
	}
	reservedWidth := leadingWidth
	if compact && expanded {
		reservedWidth = collapsedSidebarWidth
	}
	gap := theme.Padding()
	layout.content.Move(fyne.NewPos(reservedWidth+gap, 0))
	layout.content.Resize(fyne.NewSize(max(size.Width-reservedWidth-gap, 0), size.Height))
	layout.leading.Move(fyne.NewPos(0, 0))
	layout.leading.Resize(fyne.NewSize(leadingWidth, size.Height))
}

func (layout *sidebarShellLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	content := layout.content.MinSize()
	return fyne.NewSize(content.Width+theme.Padding()+collapsedSidebarWidth, max(content.Height, layout.leading.MinSize().Height))
}

// adaptiveLeadingLayout presents a leading navigation panel beside content at
// comfortable widths and over the content as a drawer at compact widths. The
// container using this layout must order content first and the leading panel
// second so the compact drawer paints above the workspace.
type adaptiveLeadingLayout struct {
	leading    fyne.CanvasObject
	content    fyne.CanvasObject
	breakpoint float32
	isOpen     func(compact bool) bool
	onCompact  func(compact bool)

	compact     bool
	initialized bool
}

func newAdaptiveLeadingContainer(layout *adaptiveLeadingLayout) *fyne.Container {
	shell := container.NewWithoutLayout(layout.content, layout.leading)
	shell.Layout = layout
	// Establish a wide initial layout without inflating MinSize. Fyne's normal
	// container constructor starts at MinSize, which is intentionally smaller
	// than the breakpoint and would briefly create an off-screen drawer before
	// the window receives its restored size.
	height := max(layout.content.MinSize().Height, layout.leading.MinSize().Height)
	shell.Resize(fyne.NewSize(layout.breakpoint, height))
	return shell
}

func (layout *adaptiveLeadingLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	compact := size.Width < layout.breakpoint
	if !layout.initialized || compact != layout.compact {
		layout.compact = compact
		layout.initialized = true
		if layout.onCompact != nil {
			layout.onCompact(compact)
		}
	}

	open := true
	if layout.isOpen != nil {
		open = layout.isOpen(compact)
	}
	setAdaptiveObjectVisible(layout.leading, open)

	layout.content.Move(fyne.NewPos(0, 0))
	layout.content.Resize(size)
	if !open {
		return
	}

	leadingWidth := layout.leading.MinSize().Width
	layout.leading.Move(fyne.NewPos(0, 0))
	layout.leading.Resize(fyne.NewSize(leadingWidth, size.Height))
	if compact {
		return
	}

	gap := theme.Padding()
	layout.content.Move(fyne.NewPos(leadingWidth+gap, 0))
	layout.content.Resize(fyne.NewSize(max(size.Width-leadingWidth-gap, 0), size.Height))
}

func (layout *adaptiveLeadingLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	content := layout.content.MinSize()
	if layout.compact || layout.leading == nil || !layout.leading.Visible() {
		return content
	}
	leading := layout.leading.MinSize()
	return fyne.NewSize(content.Width+theme.Padding()+leading.Width, max(content.Height, leading.Height))
}

func setAdaptiveObjectVisible(object fyne.CanvasObject, visible bool) {
	if object == nil || object.Visible() == visible {
		return
	}
	if visible {
		object.Show()
	} else {
		object.Hide()
	}
}
