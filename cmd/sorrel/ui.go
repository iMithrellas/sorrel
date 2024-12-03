package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	urlInput        textinput.Model
	startTimestamp  textinput.Model
	endTimestamp    textinput.Model
	focusedInputIdx int
}

func initialModel(initialURL, initialTimestamp string) model {
	url := textinput.New()
	url.Placeholder = "Enter YouTube URL"
	url.SetValue(initialURL)
	url.Focus()
	url.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	start := textinput.New()
	start.Placeholder = "Enter start timestamp (e.g., 00:01:30)"
	start.SetValue(initialTimestamp)

	end := textinput.New()
	end.Placeholder = "Enter end timestamp or duration"

	return model{
		urlInput:        url,
		startTimestamp:  start,
		endTimestamp:    end,
		focusedInputIdx: 0, // Start with URL input focused
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.focusedInputIdx = (m.focusedInputIdx + 1) % 3
			m.updateFocus()
		case "shift+tab":
			m.focusedInputIdx = (m.focusedInputIdx - 1 + 3) % 3
			m.updateFocus()
		}
	}

	var cmd tea.Cmd
	switch m.focusedInputIdx {
	case 0:
		m.urlInput, cmd = m.urlInput.Update(msg)
	case 1:
		m.startTimestamp, cmd = m.startTimestamp.Update(msg)
	case 2:
		m.endTimestamp, cmd = m.endTimestamp.Update(msg)
	}

	return m, cmd
}

func (m *model) updateFocus() {
	m.urlInput.Blur()
	m.startTimestamp.Blur()
	m.endTimestamp.Blur()

	switch m.focusedInputIdx {
	case 0:
		m.urlInput.Focus()
	case 1:
		m.startTimestamp.Focus()
	case 2:
		m.endTimestamp.Focus()
	}
}

func (m model) View() string {
	return fmt.Sprintf(
		"URL: %s\n\nStart Timestamp: %s\n\nEnd Timestamp: %s\n\n[Tab to switch, q to quit]",
		m.urlInput.View(),
		m.startTimestamp.View(),
		m.endTimestamp.View(),
	)
}
