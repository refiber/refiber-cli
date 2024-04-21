package ui

import "github.com/charmbracelet/lipgloss"

var (
	TextError = lipgloss.NewStyle().Foreground(lipgloss.Color("#F44336")).Padding(0, 0, 0)
	TextTitle = lipgloss.NewStyle().Background(lipgloss.Color("#01FAC6")).Foreground(lipgloss.Color("#030303")).Bold(true).Padding(0, 1, 0)
	TextGreen = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	TextGray  = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
)
