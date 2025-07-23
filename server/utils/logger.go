package utils

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

func InitLogger() *log.Logger {
	styles := log.DefaultStyles()
	styles.Keys["GET"] = lipgloss.NewStyle().Foreground(lipgloss.Color("#86ff33"))
	styles.Keys["POST"] = lipgloss.NewStyle().Foreground(lipgloss.Color("#2980b9"))
	styles.Keys["PUT"] = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7dc6f"))
	styles.Keys["DELETE"] = lipgloss.NewStyle().Foreground(lipgloss.Color("#e74c3c"))

	logger := log.New(os.Stderr)
	logger.SetStyles(styles)
	return logger
}
