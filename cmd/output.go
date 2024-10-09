package cmd

import "fmt"

const (
	COLOR_RESET   = "\u001B[0m"
	COLOR_DANGER  = "\033[31m"
	COLOR_SUCCESS = "\033[32m"
	COLOR_INFO    = "\033[36m"
	COLOR_WARNING = "\033[0;33m"
)

func PrintfSuccess(text string, vars ...any) {
	println(FormatSuccess(text, vars...))
}

func FormatSuccess(text string, vars ...any) string {
	return FormatWithColor(COLOR_SUCCESS, text, vars...)
}

func PrintfInfo(text string, vars ...any) {
	println(FormatInfo(text, vars...))
}

func FormatInfo(text string, vars ...any) string {
	return FormatWithColor(COLOR_INFO, text, vars...)
}

func PrintfWarning(text string, vars ...any) {
	println(FormatWarning(text, vars...))
}

func FormatWarning(text string, vars ...any) string {
	return FormatWithColor(COLOR_WARNING, text, vars...)
}

func PrintfDanger(text string, vars ...any) {
	println(FormatDanger(text, vars...))
}

func FormatDanger(text string, vars ...any) string {
	return FormatWithColor(COLOR_DANGER, text, vars...)
}

func FormatWithColor(color, text string, vars ...any) string {
	return fmt.Sprintf(color+text+COLOR_RESET, vars...)
}
