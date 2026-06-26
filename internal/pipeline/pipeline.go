// Package pipeline implements the deen GUI's data model: a source input
// followed by an ordered chain of plugin transforms, where each step's output
// feeds the next (similar to Burp Suite's Decoder). It has no GUI dependency
// so it can be unit tested without a display.
package pipeline

import (
	"bytes"
	"flag"
	"fmt"

	"github.com/takeshixx/deen/internal/plugins"
)

// Step is a single transform in the pipeline.
type Step struct {
	Plugin    string            // base plugin name, e.g. "base64"
	Unprocess bool              // run the decode direction
	Options   map[string]string // flag name -> value (string form)

	override    []byte // user-edited content that replaces the computed output
	hasOverride bool
	output      []byte
	err         error
}

// Pipeline holds the source input and the chain of steps.
type Pipeline struct {
	source []byte
	steps  []*Step
}

// New returns an empty pipeline.
func New() *Pipeline { return &Pipeline{} }

// Steps returns the current steps.
func (p *Pipeline) Steps() []*Step { return p.steps }

// Len returns the number of steps.
func (p *Pipeline) Len() int { return len(p.steps) }

// Source returns the current source input.
func (p *Pipeline) Source() []byte { return p.source }

// Output returns the output of step i.
func (p *Pipeline) Output(i int) []byte { return p.steps[i].output }

// Err returns the error (if any) produced by step i.
func (p *Pipeline) Err(i int) error { return p.steps[i].err }

// Result returns the output of the last step, or the source if there are none.
func (p *Pipeline) Result() []byte {
	if len(p.steps) == 0 {
		return p.source
	}
	return p.steps[len(p.steps)-1].output
}

// SetSource sets the source input. Editing the source invalidates any
// downstream manual edits.
func (p *Pipeline) SetSource(b []byte) {
	p.source = b
	p.clearOverrides(0)
	p.Compute()
}

// AddStep appends a transform and returns its index.
func (p *Pipeline) AddStep(plugin string, unprocess bool) int {
	p.steps = append(p.steps, &Step{Plugin: plugin, Unprocess: unprocess, Options: map[string]string{}})
	p.Compute()
	return len(p.steps) - 1
}

// RemoveStep removes the step at index i.
func (p *Pipeline) RemoveStep(i int) {
	if i < 0 || i >= len(p.steps) {
		return
	}
	p.steps = append(p.steps[:i], p.steps[i+1:]...)
	p.Compute()
}

// SetPlugin changes the transform of step i. This clears manual edits from i
// onward, since the transform output changes.
func (p *Pipeline) SetPlugin(i int, plugin string, unprocess bool) {
	if i < 0 || i >= len(p.steps) {
		return
	}
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
	p.clearOverrides(i + 1)
	p.steps[i].override = data
	p.steps[i].hasOverride = true
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
		prev = s.output
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
		return nil, err
	}
	return buf.Bytes(), nil
}
