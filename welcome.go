package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/common-nighthawk/go-figure"
	"github.com/gliderlabs/ssh"
)

func welcomeFooter(m welcome) string {
	s1 := lipgloss.NewStyle().Foreground(lipgloss.Color("#565665"))
	s2 := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#565665"))

	return strings.Join([]string{
		s2.Render("↑/k"),
		s1.Render("up •"),
		s2.Render("↓/j"),
		s1.Render("down •"),
		s2.Render("enter"),
		s1.Render("select •"),
		s2.Render("q"),
		s1.Render("adiós"),
	}, " ")
}

func markdownFooter(m welcome) string {
	s1 := lipgloss.NewStyle().Foreground(lipgloss.Color("#565665"))
	s2 := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#565665"))

	return strings.Join([]string{
		s2.Render("↑/k"),
		s1.Render("up •"),
		s2.Render("↓/j"),
		s1.Render("down •"),
		s2.Render("esc"),
		s1.Render("go back"),
	}, " ")
}

type choice struct {
	title       string
	description string
	src         string
	content     string
}

type welcome struct {
	w              int
	h              int
	headerMarkdown string
	footerMarkdown string
	cursor         int
	selected       bool
	viewportReady  bool
	viewport       viewport.Model
	choices        []choice
	help           help.Model
}

func (m welcome) Init() tea.Cmd {
	for i, choice := range m.choices {
		m.choices[i].content = prepareChoiceContent(choice, m)
	}
	return nil
}

func InitialiseWelcomeScreen(pty ssh.Pty) welcome {
	width := pty.Window.Width
	height := pty.Window.Height

	return welcome{
		w:              width,
		h:              height,
		headerMarkdown: "My name is **Antoni**, welcome to the ssh version of my website. Choose what you want to read about below.",
		footerMarkdown: "",
		cursor:         0,
		selected:       false,
		choices: []choice{
			{
				title:       "Resume",
				description: "This is my resume as seen on read.cv/antoni",
				src:         "data/resume.md",
			},
			{
				title:       "Links",
				description: "Links to all of my social media accounts",
				src:         "data/links.md",
			},
		},
	}
}

func (m welcome) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.selected {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.viewportReady = false
				m.selected = false
				return m, nil
			}
		}

		var (
			cmd  tea.Cmd
			cmds []tea.Cmd
		)

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			selectedChoice := m.choices[m.cursor]

			footerHeight := lipgloss.Height(markdownFooter(m)) + 2

			m.viewport = viewport.New(m.w, m.h-footerHeight)
			m.viewport.SetContent(selectedChoice.content)
			m.viewportReady = true

			m.selected = true
		}
	}

	return m, nil
}

func (m welcome) View() string {
	if m.selected {
		if !m.viewportReady {
			return "\n Loading..."
		}

		return m.viewport.View() + "\n\n" + markdownFooter(m) + "\n"
	} else {
		m.viewportReady = false
		return renderWelcomeScreen(m)
	}
}

func renderWelcomeScreen(m welcome) string {

	nameFigure := figure.NewFigure("Welcome", "", true)
	s := nameFigure.String()

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.w),
	)
	md, _ := r.Render(m.headerMarkdown)
	s += md

	for i, choice := range m.choices {
		isSelected := m.cursor == i

		listContainerStyle := lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder())
		if isSelected {
			listContainerStyle = listContainerStyle.BorderForeground(lipgloss.Color("#FF75B7"))
		}
		listTitleStyle := lipgloss.NewStyle().Bold(true)

		s += listContainerStyle.Render(listTitleStyle.Render(choice.title)+" - "+choice.description) + "\n"
	}

	contentHeight := lipgloss.Height(s)
	footerHeight := lipgloss.Height(welcomeFooter(m))
	spaceLinesRequired := m.h - contentHeight - footerHeight
	extraSpaces := ""
	if spaceLinesRequired > 0 {
		extraSpaces = strings.Repeat("\n", spaceLinesRequired)
	}

	return s + extraSpaces + welcomeFooter(m)
}

func prepareChoiceContent(c choice, m welcome) string {
	src := c.src

	if strings.HasSuffix(src, ".md") {
		rawMarkdown := ""
		if strings.HasPrefix(src, "http") {
			res, _ := http.Get(src)
			body, _ := ioutil.ReadAll(res.Body)
			rawMarkdown = string(body)
		} else {
			rawFile, _ := os.ReadFile(src)
			rawMarkdown = string(rawFile)
		}

		return markdownRenderer(rawMarkdown, m.w)
	}

	return "Unknown format"
}

func markdownRenderer(rawMarkdown string, w int) string {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(w),
	)
	out, _ := r.Render(rawMarkdown)
	return out
}
