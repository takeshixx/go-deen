// Package pipeline implements the deen GUI's data model: a source input
// followed by an ordered chain of plugin transforms, where each step's output
// feeds the next (similar to Burp Suite's Decoder). It has no GUI dependency
// so it can be unit tested without a display.
package pipeline

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/takeshixx/deen/internal/plugins"
)

// Step is a single transform in the pipeline.
type Step struct {
	Plugin    string            // base plugin name, e.g. "base64"
	Unprocess bool              // run the decode direction
	Options   map[string]string // flag name -> value (string form)
	Disabled  bool              // bypass this step without removing it

	override    []byte // user-edited content that replaces the computed output
	hasOverride bool
	output      []byte
	effective   []byte
	err         error
}

// Pipeline holds the source input and the chain of steps.
type Pipeline struct {
	source []byte
	steps  []*Step

	undo []snapshot
	redo []snapshot
}

// New returns an empty pipeline.
func New() *Pipeline { return &Pipeline{} }

const maxHistory = 100

type stepSnapshot struct {
	Plugin      string
	Unprocess   bool
	Options     map[string]string
	Disabled    bool
	Override    []byte
	HasOverride bool
}

type snapshot struct {
	Source []byte
	Steps  []stepSnapshot
}

type chainFile struct {
	Version int             `json:"version"`
	Source  []byte          `json:"source,omitempty"`
	Steps   []chainFileStep `json:"steps"`
}

type chainFileStep struct {
	Plugin      string            `json:"plugin"`
	Unprocess   bool              `json:"unprocess,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
	Disabled    bool              `json:"disabled,omitempty"`
	Override    []byte            `json:"override,omitempty"`
	HasOverride bool              `json:"hasOverride,omitempty"`
}

// Steps returns the current steps.
func (p *Pipeline) Steps() []*Step { return p.steps }

// Len returns the number of steps.
func (p *Pipeline) Len() int { return len(p.steps) }

// Source returns the current source input.
func (p *Pipeline) Source() []byte { return p.source }

// Clear resets the source and all steps while keeping undo history available.
func (p *Pipeline) Clear() {
	p.record()
	p.source = nil
	p.steps = nil
	p.Compute()
}

// Output returns the displayed output of step i. Disabled steps still compute
// their transform for display, but their effective chain output is bypassed.
func (p *Pipeline) Output(i int) []byte { return p.steps[i].output }

// Input returns the bytes used as input for step i.
func (p *Pipeline) Input(i int) []byte {
	if i <= 0 {
		return p.source
	}
	if i > len(p.steps) {
		return nil
	}
	return p.steps[i-1].effective
}

// Err returns the error (if any) produced by step i.
func (p *Pipeline) Err(i int) error { return p.steps[i].err }

// Result returns the output of the last step, or the source if there are none.
func (p *Pipeline) Result() []byte {
	if len(p.steps) == 0 {
		return p.source
	}
	return p.steps[len(p.steps)-1].effective
}

// ExportJSON serializes the editable pipeline state, including binary source
// input and manual output overrides. Byte slices are encoded as base64 by JSON.
func (p *Pipeline) ExportJSON() ([]byte, error) {
	return p.exportJSON(true)
}

// ExportJSONWithoutSource serializes the transform chain without source input
// or manual output overrides. It is intended for shareable recipes that should
// not leak the user's data.
func (p *Pipeline) ExportJSONWithoutSource() ([]byte, error) {
	return p.exportJSON(false)
}

func (p *Pipeline) exportJSON(includeSource bool) ([]byte, error) {
	cf := chainFile{
		Version: 1,
		Steps:   make([]chainFileStep, 0, len(p.steps)),
	}
	if includeSource {
		cf.Source = append([]byte(nil), p.source...)
	}
	for _, step := range p.steps {
		opts := make(map[string]string, len(step.Options))
		for k, v := range step.Options {
			opts[k] = v
		}
		cfs := chainFileStep{
			Plugin:    step.Plugin,
			Unprocess: step.Unprocess,
			Options:   opts,
			Disabled:  step.Disabled,
		}
		if includeSource {
			cfs.Override = append([]byte(nil), step.override...)
			cfs.HasOverride = step.hasOverride
		}
		cf.Steps = append(cf.Steps, cfs)
	}
	return json.MarshalIndent(cf, "", "  ")
}

// ImportJSON replaces the pipeline with a serialized chain. The previous state
// is retained in undo history so a user can recover from an accidental import.
func (p *Pipeline) ImportJSON(data []byte) error {
	var cf chainFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return err
	}
	if cf.Version != 1 {
		return fmt.Errorf("unsupported chain version %d", cf.Version)
	}

	s := snapshot{
		Source: append([]byte(nil), cf.Source...),
		Steps:  make([]stepSnapshot, 0, len(cf.Steps)),
	}
	for i, step := range cf.Steps {
		pluginName := step.Plugin
		if step.Plugin != "" {
			plugin, _, ok := plugins.Resolve(step.Plugin)
			if !ok {
				return fmt.Errorf("step %d: unknown plugin %q", i+1, step.Plugin)
			}
			pluginName = plugin.Name
		}
		opts := normalizeStepOptions(pluginName, step.Options)
		unprocess := step.Unprocess && plugins.CanDecode(pluginName)
		s.Steps = append(s.Steps, stepSnapshot{
			Plugin:      pluginName,
			Unprocess:   unprocess,
			Options:     opts,
			Disabled:    step.Disabled,
			Override:    append([]byte(nil), step.Override...),
			HasOverride: step.HasOverride,
		})
	}

	p.record()
	p.restore(s)
	return nil
}

// SetSource sets the source input. Editing the source invalidates any
// downstream manual edits.
func (p *Pipeline) SetSource(b []byte) {
	p.setSource(append([]byte(nil), b...))
}

// SetSourceOwned sets the source input without copying b. Callers must not
// mutate b after handing it to the pipeline.
func (p *Pipeline) SetSourceOwned(b []byte) {
	p.setSource(b)
}

func (p *Pipeline) setSource(b []byte) {
	p.record()
	p.source = b
	p.clearOverrides(0)
	p.Compute()
}

// AddStep appends a transform and returns its index.
func (p *Pipeline) AddStep(plugin string, unprocess bool) int {
	return p.AddStepWithOptions(plugin, unprocess, nil)
}

// AddStepWithOptions appends a transform with initial option values and returns
// its index.
func (p *Pipeline) AddStepWithOptions(plugin string, unprocess bool, options map[string]string) int {
	p.record()
	if unprocess && !plugins.CanDecode(plugin) {
		unprocess = false
	}
	opts := normalizeStepOptions(plugin, options)
	p.steps = append(p.steps, &Step{Plugin: plugin, Unprocess: unprocess, Options: opts})
	p.Compute()
	return len(p.steps) - 1
}

func normalizeStepOptions(plugin string, options map[string]string) map[string]string {
	opts := make(map[string]string, len(options))
	for k, v := range options {
		opts[k] = v
	}
	if plugin != "aes" {
		return opts
	}
	nonce, hasNonce := opts["nonce"]
	iv, hasIV := opts["iv"]
	switch {
	case hasNonce && (!hasIV || iv == ""):
		opts["iv"] = nonce
		delete(opts, "nonce")
	case hasNonce && hasIV && nonce == iv:
		delete(opts, "nonce")
	}
	return opts
}

// RemoveStep removes the step at index i.
func (p *Pipeline) RemoveStep(i int) {
	if i < 0 || i >= len(p.steps) {
		return
	}
	p.record()
	p.steps = append(p.steps[:i], p.steps[i+1:]...)
	p.Compute()
}

// MoveStep moves a step to a new index.
func (p *Pipeline) MoveStep(from, to int) {
	if from < 0 || from >= len(p.steps) || to < 0 || to >= len(p.steps) || from == to {
		return
	}
	p.record()
	step := p.steps[from]
	p.steps = append(p.steps[:from], p.steps[from+1:]...)
	p.steps = append(p.steps[:to], append([]*Step{step}, p.steps[to:]...)...)
	p.clearOverrides(min(from, to))
	p.Compute()
}

// DuplicateStep copies a step and inserts it directly after the original.
func (p *Pipeline) DuplicateStep(i int) {
	if i < 0 || i >= len(p.steps) {
		return
	}
	p.record()
	p.steps = append(p.steps[:i+1], append([]*Step{cloneStep(p.steps[i])}, p.steps[i+1:]...)...)
	p.clearOverrides(i + 1)
	p.Compute()
}

// SetStepDisabled enables or disables a step. Disabled steps pass their input
// through unchanged for the chain, while still keeping their transform output
// visible on the step itself.
func (p *Pipeline) SetStepDisabled(i int, disabled bool) {
	if i < 0 || i >= len(p.steps) || p.steps[i].Disabled == disabled {
		return
	}
	p.record()
	p.steps[i].Disabled = disabled
	p.clearOverrides(i + 1)
	p.Compute()
}

// SetPlugin changes the transform of step i. This clears manual edits from i
// onward, since the transform output changes.
func (p *Pipeline) SetPlugin(i int, plugin string, unprocess bool) {
	if i < 0 || i >= len(p.steps) {
		return
	}
	if unprocess && !plugins.CanDecode(plugin) {
		unprocess = false
	}
	p.record()
	p.steps[i].Plugin = plugin
	p.steps[i].Unprocess = unprocess
	p.clearOverrides(i)
	p.Compute()
}

// SetOption sets a plugin option (flag) value on step i.
func (p *Pipeline) SetOption(i int, name, value string) {
	if i < 0 || i >= len(p.steps) {
		return
	}
	p.record()
	if p.steps[i].Options == nil {
		p.steps[i].Options = map[string]string{}
	}
	p.steps[i].Options[name] = value
	p.clearOverrides(i)
	p.Compute()
}

// EditOutput overrides the content of step i with user-edited data and
// recomputes everything below it.
func (p *Pipeline) EditOutput(i int, data []byte) {
	if i < 0 || i >= len(p.steps) {
		return
	}
	p.record()
	p.clearOverrides(i + 1)
	p.steps[i].override = data
	p.steps[i].hasOverride = true
	p.Compute()
}

// CanUndo reports whether Undo can restore an earlier pipeline state.
func (p *Pipeline) CanUndo() bool { return len(p.undo) > 0 }

// CanRedo reports whether Redo can re-apply a state that was undone.
func (p *Pipeline) CanRedo() bool { return len(p.redo) > 0 }

// Undo restores the previous pipeline state. It returns false when there is no
// earlier state available.
func (p *Pipeline) Undo() bool {
	if !p.CanUndo() {
		return false
	}
	current := p.snapshot()
	prev := p.undo[len(p.undo)-1]
	p.undo = p.undo[:len(p.undo)-1]
	p.redo = appendLimited(p.redo, current)
	p.restore(prev)
	return true
}

// Redo reapplies the most recently undone pipeline state. It returns false when
// there is no state available.
func (p *Pipeline) Redo() bool {
	if !p.CanRedo() {
		return false
	}
	current := p.snapshot()
	next := p.redo[len(p.redo)-1]
	p.redo = p.redo[:len(p.redo)-1]
	p.undo = appendLimited(p.undo, current)
	p.restore(next)
	return true
}

func (p *Pipeline) record() {
	p.undo = appendLimited(p.undo, p.snapshot())
	p.redo = nil
}

func appendLimited(history []snapshot, s snapshot) []snapshot {
	history = append(history, s)
	if len(history) > maxHistory {
		copy(history, history[len(history)-maxHistory:])
		history = history[:maxHistory]
	}
	return history
}

func (p *Pipeline) snapshot() snapshot {
	s := snapshot{
		Source: p.source,
		Steps:  make([]stepSnapshot, 0, len(p.steps)),
	}
	for _, step := range p.steps {
		opts := make(map[string]string, len(step.Options))
		for k, v := range step.Options {
			opts[k] = v
		}
		s.Steps = append(s.Steps, stepSnapshot{
			Plugin:      step.Plugin,
			Unprocess:   step.Unprocess,
			Options:     opts,
			Disabled:    step.Disabled,
			Override:    append([]byte(nil), step.override...),
			HasOverride: step.hasOverride,
		})
	}
	return s
}

func (p *Pipeline) restore(s snapshot) {
	p.source = s.Source
	p.steps = make([]*Step, 0, len(s.Steps))
	for _, ss := range s.Steps {
		opts := make(map[string]string, len(ss.Options))
		for k, v := range ss.Options {
			opts[k] = v
		}
		p.steps = append(p.steps, &Step{
			Plugin:      ss.Plugin,
			Unprocess:   ss.Unprocess,
			Options:     opts,
			Disabled:    ss.Disabled,
			override:    append([]byte(nil), ss.Override...),
			hasOverride: ss.HasOverride,
		})
	}
	p.Compute()
}

// clearOverrides drops manual edits for all steps with index >= from.
func (p *Pipeline) clearOverrides(from int) {
	for j := from; j < len(p.steps); j++ {
		p.steps[j].hasOverride = false
		p.steps[j].override = nil
	}
}

// Compute (re)evaluates every step's output from the source down.
func (p *Pipeline) Compute() {
	prev := p.source
	for _, s := range p.steps {
		switch {
		case s.hasOverride:
			s.output, s.err = s.override, nil
		case s.Plugin == "":
			s.output, s.err = prev, nil // passthrough until a transform is chosen
		default:
			s.output, s.err = runStep(s, prev)
		}
		if s.Disabled {
			s.effective = prev
		} else {
			s.effective = s.output
		}
		prev = s.effective
	}
}

func cloneStep(s *Step) *Step {
	opts := make(map[string]string, len(s.Options))
	for k, v := range s.Options {
		opts[k] = v
	}
	return &Step{
		Plugin:      s.Plugin,
		Unprocess:   s.Unprocess,
		Options:     opts,
		Disabled:    s.Disabled,
		override:    append([]byte(nil), s.override...),
		hasOverride: s.hasOverride,
	}
}

// runStep executes a single plugin transform.
func runStep(s *Step, in []byte) ([]byte, error) {
	cmd := s.Plugin
	if s.Unprocess {
		cmd = "." + s.Plugin
	}
	plugin, unprocess, ok := plugins.Resolve(cmd)
	if !ok {
		return nil, fmt.Errorf("unknown plugin %q", s.Plugin)
	}
	fn := plugin.Process
	if unprocess {
		if plugin.Unprocess == nil {
			return nil, fmt.Errorf("%s does not support decoding", s.Plugin)
		}
		fn = plugin.Unprocess
	}

	fs := flag.NewFlagSet(s.Plugin, flag.ContinueOnError)
	if plugin.RegisterFlags != nil {
		plugin.RegisterFlags(fs)
	}
	for name, value := range s.Options {
		if fs.Lookup(name) != nil {
			_ = fs.Set(name, value)
		}
	}

	var buf bytes.Buffer
	if err := fn(bytes.NewReader(in), &buf, fs); err != nil {
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}
