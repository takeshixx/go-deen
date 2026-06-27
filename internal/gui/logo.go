//go:build gui

package gui

import "fyne.io/fyne/v2"

var deenLogoResource = fyne.NewStaticResource("deen-logo.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
	<rect width="64" height="64" rx="14" fill="#0f172a"/>
	<path d="M22 17v30" fill="none" stroke="#e5e7eb" stroke-width="6" stroke-linecap="round"/>
	<path d="M22 17c19 0 27 8 27 15s-8 15-27 15" fill="none" stroke="#e5e7eb" stroke-width="6" stroke-linecap="round" stroke-linejoin="round"/>
	<path d="M22 32h18" fill="none" stroke="#38bdf8" stroke-width="5" stroke-linecap="round"/>
	<circle cx="22" cy="17" r="5.5" fill="#0f172a" stroke="#22c55e" stroke-width="4"/>
	<circle cx="22" cy="32" r="5.5" fill="#0f172a" stroke="#38bdf8" stroke-width="4"/>
	<circle cx="22" cy="47" r="5.5" fill="#0f172a" stroke="#f59e0b" stroke-width="4"/>
	<circle cx="40" cy="32" r="5.5" fill="#0f172a" stroke="#38bdf8" stroke-width="4"/>
</svg>`))
