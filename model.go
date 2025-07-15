package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	username     string
	pet          Pet
	width        int
	height       int
	lastMessage  string
	updates      chan GameUpdate
	txtStyle     lipgloss.Style
	titleStyle   lipgloss.Style
	healthStyle  lipgloss.Style
	actionStyle  lipgloss.Style
	quitStyle    lipgloss.Style
}

type updateMsg GameUpdate

func NewModel(username string, renderer *lipgloss.Renderer, width, height int) Model {
	updates := Subscribe(username)
	
	return Model{
		username:    username,
		pet:         GetGameState(),
		width:       width,
		height:      height,
		updates:     updates,
		txtStyle:    renderer.NewStyle().Foreground(lipgloss.Color("15")),
		titleStyle:  renderer.NewStyle().Foreground(lipgloss.Color("13")).Bold(true),
		healthStyle: renderer.NewStyle().Foreground(lipgloss.Color("10")),
		actionStyle: renderer.NewStyle().Foreground(lipgloss.Color("8")),
		quitStyle:   renderer.NewStyle().Foreground(lipgloss.Color("8")).Italic(true),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitForUpdate(m.updates),
		tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tea.KeyMsg{}
		}),
	)
}

func waitForUpdate(updates chan GameUpdate) tea.Cmd {
	return func() tea.Msg {
		return updateMsg(<-updates)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case updateMsg:
		m.pet = msg.Pet
		if msg.Message != "" {
			m.lastMessage = msg.Message
		}
		return m, waitForUpdate(m.updates)

	case tea.KeyMsg:
		switch msg.String() {
		case "f":
			m.lastMessage = FeedPet(m.username)
		case "p":
			m.lastMessage = PetPet(m.username)
		case "h":
			m.lastMessage = HitPet(m.username)
		case "q", "ctrl+c":
			Unsubscribe(m.username)
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	// Title
	title := "🐾 SSH TAMAGOTCHI 🐾"
	b.WriteString(m.titleStyle.Render(title))
	b.WriteString("\n\n")

	// Pet info
	petName := fmt.Sprintf("Name: %s", m.pet.Name)
	b.WriteString(m.txtStyle.Render(petName))
	b.WriteString("\n")

	// Health bar
	healthBar := m.renderHealthBar()
	b.WriteString(healthBar)
	b.WriteString("\n")

	// Mood with emoji
	mood := m.renderMood()
	b.WriteString(m.txtStyle.Render(mood))
	b.WriteString("\n\n")

	// Recent actions
	b.WriteString(m.txtStyle.Render("Recent Actions:"))
	b.WriteString("\n")
	
	if len(m.pet.Actions) == 0 {
		b.WriteString(m.actionStyle.Render("  No actions yet..."))
		b.WriteString("\n")
	} else {
		for i := len(m.pet.Actions) - 1; i >= 0; i-- {
			action := m.pet.Actions[i]
			actionText := fmt.Sprintf("  %s %s", action.User, m.actionToText(action.Type))
			b.WriteString(m.actionStyle.Render(actionText))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Last message
	if m.lastMessage != "" {
		b.WriteString(m.healthStyle.Render(">>> " + m.lastMessage))
		b.WriteString("\n\n")
	}

	// Controls
	controls := "f=feed  p=pet  h=hit  q=quit"
	b.WriteString(m.quitStyle.Render(controls))
	b.WriteString("\n")

	// Connected as
	user := fmt.Sprintf("Connected as: %s", m.username)
	b.WriteString(m.quitStyle.Render(user))

	return b.String()
}

func (m Model) renderHealthBar() string {
	healthPercent := float64(m.pet.Health) / float64(m.pet.MaxHealth)
	barWidth := 20
	filledBars := int(healthPercent * float64(barWidth))
	
	var bar strings.Builder
	bar.WriteString("Health: ")
	
	// Health bar
	for i := 0; i < barWidth; i++ {
		if i < filledBars {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}
	
	healthText := fmt.Sprintf(" %d/%d", m.pet.Health, m.pet.MaxHealth)
	
	var style lipgloss.Style
	if m.pet.Health <= 0 {
		style = m.txtStyle.Foreground(lipgloss.Color("9")) // Red
	} else if m.pet.Health < 30 {
		style = m.txtStyle.Foreground(lipgloss.Color("11")) // Yellow
	} else {
		style = m.healthStyle // Green
	}
	
	return style.Render(bar.String() + healthText)
}

func (m Model) renderMood() string {
	var emoji string
	switch m.pet.Mood {
	case "happy":
		emoji = "😊"
	case "content":
		emoji = "😌"
	case "neutral":
		emoji = "😐"
	case "sad":
		emoji = "😢"
	case "angry":
		emoji = "😠"
	case "dying":
		emoji = "💀"
	case "dead":
		emoji = "☠️"
	case "revived":
		emoji = "✨"
	default:
		emoji = "🤔"
	}
	
	return fmt.Sprintf("Mood: %s %s", emoji, m.pet.Mood)
}

func (m Model) actionToText(actionType string) string {
	switch actionType {
	case "fed":
		return "fed me! 🍎"
	case "petted":
		return "petted me! 💕"
	case "hit":
		return "hit me! 💥"
	case "killed":
		return "killed me! ☠️"
	case "revived":
		return "revived me! ✨"
	case "died":
		return "I died from neglect! 💀"
	default:
		return actionType
	}
}
