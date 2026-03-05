package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "faaah",
	Short: "Play a sound every time a shell command fails",
	Long:  "faaah hooks into your shell (bash/zsh) and plays a sound\nevery time a command exits with a non-zero status code.",
}
