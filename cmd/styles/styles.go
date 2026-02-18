package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	PurpleDeep   = lipgloss.Color("#8B6BB8")
	PurpleMid    = lipgloss.Color("#9B7BE6")
	PurpleLight  = lipgloss.Color("#BDA7FF")
	PurplePale   = lipgloss.Color("#DDD2FF")

	Primary   = PurpleDeep
	Secondary = PurpleLight
	Success   = lipgloss.Color("51")  // Cyan-ish
	Warning   = lipgloss.Color("214") // Orange
	Error     = lipgloss.Color("196") // Red
	Muted     = lipgloss.Color("243") // Gray
	Info      = lipgloss.Color("86")  // Cyan

	// Progress gradient (hex values are required by bubbles/progress gradient parser)
	ProgressGradientFromHex = "#00D7D7"
	ProgressGradientToHex   = "#FF5FD7"

	// App title styles
	AppTitleStyle = lipgloss.NewStyle().
			Foreground(Info).
			Bold(true).
			MarginBottom(1)

	// Header styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(PurpleMid).
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
			Foreground(PurpleDeep).
			Bold(true)

	SecondaryStyle = lipgloss.NewStyle().
			Foreground(PurpleLight).
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
			Foreground(PurplePale)

	// Progress styles
	ProgressLabelStyle = lipgloss.NewStyle().
				Foreground(PurpleMid).
				Bold(true)

	// Box/Border styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PurpleMid).
			Padding(1, 2)

	SubtleBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(PurpleLight).
			Padding(0, 1)
)
