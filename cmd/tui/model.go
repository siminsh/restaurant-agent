package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"restaurant-agent/pkg/agent"
)

// -- Message types --

type agentResponseMsg struct {
	response  string
	toolCalls int
}

type agentErrorMsg struct {
	err error
}

// -- Chat message --

type chatMessage struct {
	role      string // "user", "assistant", "error"
	content   string
	toolCalls int
}

// -- Styles --

var (
	primaryColor   = lipgloss.Color("#FF6B35")
	secondaryColor = lipgloss.Color("#004E89")
	accentColor    = lipgloss.Color("#2EC4B6")
	dimColor       = lipgloss.Color("#666666")
	errorColor     = lipgloss.Color("#E63946")

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 2)

	userLabelStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	userMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(secondaryColor).
			Padding(0, 2)

	assistantLabelStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	toolBadgeStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	inputBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor)
)

// -- Model --

type model struct {
	agent     *agent.Agent
	sessionID string

	textarea textarea.Model
	spinner  spinner.Model
	viewport viewport.Model

	messages []chatMessage
	loading  bool
	ready    bool
	width    int
	height   int

	renderer *glamour.TermRenderer
}

func newModel(ag *agent.Agent, sessionID string) model {
	ta := textarea.New()
	ta.Placeholder = "Ask about orders, inventory, staff, or say 'morning briefing'..."
	ta.Focus()
	ta.CharLimit = 500
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.KeyMap.InsertNewline.SetEnabled(false)

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = spinnerStyle

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(80),
	)

	return model{
		agent:     ag,
		sessionID: sessionID,
		textarea:  ta,
		spinner:   sp,
		renderer:  renderer,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		inputHeight := 5
		viewportHeight := m.height - headerHeight - inputHeight

		if !m.ready {
			m.viewport = viewport.New(m.width, viewportHeight)
			m.viewport.SetContent(m.renderWelcome())
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = viewportHeight
		}

		m.textarea.SetWidth(m.width - 4)

		m.renderer, _ = glamour.NewTermRenderer(
			glamour.WithStylePath("dark"),
			glamour.WithWordWrap(m.width-6),
		)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.loading {
				return m, nil
			}
			return m, tea.Quit
		case "enter":
			if m.loading {
				return m, nil
			}
			input := strings.TrimSpace(m.textarea.Value())
			if input == "" {
				return m, nil
			}
			switch input {
			case "/quit", "/exit":
				return m, tea.Quit
			case "/clear":
				m.messages = nil
				m.agent.ClearSession(m.sessionID)
				m.viewport.SetContent(m.renderWelcome())
				m.textarea.Reset()
				return m, nil
			}

			m.messages = append(m.messages, chatMessage{
				role:    "user",
				content: input,
			})
			m.textarea.Reset()
			m.loading = true
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			return m, m.sendMessage(input)
		}

	case agentResponseMsg:
		m.loading = false
		rendered := msg.response
		if m.renderer != nil {
			if r, err := m.renderer.Render(msg.response); err == nil {
				rendered = strings.TrimRight(r, "\n")
			}
		}
		m.messages = append(m.messages, chatMessage{
			role:      "assistant",
			content:   rendered,
			toolCalls: msg.toolCalls,
		})
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case agentErrorMsg:
		m.loading = false
		m.messages = append(m.messages, chatMessage{
			role:    "error",
			content: fmt.Sprintf("Error: %v", msg.err),
		})
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	if !m.loading {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	header := m.renderHeader()
	chat := m.viewport.View()
	input := m.renderInput()

	return lipgloss.JoinVertical(lipgloss.Left, header, chat, input)
}

// -- Render helpers --

func (m model) renderHeader() string {
	title := "  De Gouden Lepel - Operations Agent"
	help := "  /clear = reset  |  ctrl+c = quit"

	left := headerStyle.Render(title)
	right := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD0B0")).
		Background(primaryColor).
		Padding(0, 2).
		Render(help)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	mid := lipgloss.NewStyle().
		Background(primaryColor).
		Render(strings.Repeat(" ", gap))

	bar := left + mid + right + "\n"
	return bar
}

func (m model) renderMessages() string {
	var b strings.Builder

	for _, msg := range m.messages {
		switch msg.role {
		case "user":
			label := userLabelStyle.Render("  You")
			b.WriteString(label + "\n")
			maxW := m.width - 8
			if maxW < 20 {
				maxW = 20
			}
			b.WriteString(userMsgStyle.Width(maxW).Render(msg.content))
			b.WriteString("\n\n")

		case "assistant":
			label := assistantLabelStyle.Render("  Operations Agent")
			toolInfo := ""
			if msg.toolCalls > 0 {
				toolInfo = toolBadgeStyle.Render(
					fmt.Sprintf("  [%d tool call(s)]", msg.toolCalls),
				)
			}
			b.WriteString(label + toolInfo + "\n")
			b.WriteString(msg.content)
			b.WriteString("\n")

		case "error":
			b.WriteString(errorStyle.Render("  "+msg.content) + "\n\n")
		}
	}

	if m.loading {
		thinking := "\n" + spinnerStyle.Render(m.spinner.View()) + " Thinking...\n"
		b.WriteString(thinking)
	}

	return b.String()
}

func (m model) renderInput() string {
	if m.loading {
		return inputBorderStyle.Width(m.width - 4).Render(
			spinnerStyle.Render(m.spinner.View()) + " Agent is working...",
		)
	}
	return inputBorderStyle.Width(m.width - 4).Render(m.textarea.View())
}

func (m model) renderWelcome() string {
	welcome := `# De Gouden Lepel - Operations Agent

Welcome! I'm your AI operations assistant. Try asking:

- **"Give me my morning briefing"** — full operations overview
- **"How did we do today?"** — order and revenue summary
- **"What are our top selling items this week?"**
- **"Which channel brings in the most revenue?"**
- **"Who's working today?"** — staff schedule
- **"What items are low on stock?"** — inventory alerts
- **"Can we make 20 biefstukken?"** — menu feasibility check
- **"Order more zalm, we're running low"**

Type your message below and press **Enter**.
`
	if m.renderer != nil {
		if r, err := m.renderer.Render(welcome); err == nil {
			return r
		}
	}
	return welcome
}

func (m model) sendMessage(input string) tea.Cmd {
	ag := m.agent
	sid := m.sessionID
	return func() tea.Msg {
		resp, err := ag.Chat(context.Background(), sid, input)
		if err != nil {
			return agentErrorMsg{err: err}
		}
		return agentResponseMsg{
			response:  resp.Response,
			toolCalls: resp.ToolCalls,
		}
	}
}
