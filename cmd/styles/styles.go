package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Primary   = lipgloss.Color("212") // Pink
	Secondary = lipgloss.Color("99")  // Purple
	Success   = lipgloss.Color("42")  // Green
	Warning   = lipgloss.Color("214") // Orange
	Error     = lipgloss.Color("196") // Red
	Muted     = lipgloss.Color("243") // Gray
	Info      = lipgloss.Color("86")  // Cyan

	// Progress gradient (hex values are required by bubbles/progress gradient parser)
	ProgressGradientFromHex = "#00D7D7"
	ProgressGradientToHex   = "#FF5FD7"

	// App title styles
	AppTitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			MarginBottom(1)

	// Header styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Info)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	PrimaryStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	SecondaryStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	// Interactive styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginTop(1)

	// Timer/Stopwatch styles
	TimerStyle = lipgloss.NewStyle().
			Foreground(Info).
			Bold(true)

	TimerLabelStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	// Progress styles
	ProgressLabelStyle = lipgloss.NewStyle().
				Foreground(Secondary).
				Bold(true)

	// Box/Border styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Secondary).
			Padding(1, 2)

	SubtleBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(Muted).
			Padding(0, 1)
)
