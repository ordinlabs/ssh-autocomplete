package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const (
	markerBegin = "# BEGIN ssh-autocomplete"
	markerEnd   = "# END ssh-autocomplete"
)

// generateCommand returns the command string that completion scripts should use
// to invoke ssh-autocomplete. In dev mode (go run), this returns a command that
// runs from the project directory, otherwise it returns "ssh-autocomplete".
func generateCommand() string {
	exe, err := os.Executable()
	if err != nil {
		return "ssh-autocomplete"
	}

	// go run compiles to a temp directory. Detect that.
	tmpDir := os.TempDir()
	if strings.HasPrefix(exe, tmpDir) || strings.Contains(exe, "go-build") {
		// We're running via go run — find the project directory
		// Use the working directory as the project root
		wd, err := os.Getwd()
		if err != nil {
			return "ssh-autocomplete"
		}
		return fmt.Sprintf("go run -C \"%s\" .", wd)
	}

	return "ssh-autocomplete"
}

func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Install SSH host autocompletion into your shell profile",
		Long:  "Detects your shell, writes completion scripts to ~/.ordin/, and adds the source line to your shell profile. Use 'setup help' to see manual instructions instead.",
		RunE:  runSetup,
	}

	cmd.AddCommand(newSetupHelpCmd())

	return cmd
}

func newSetupHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "help",
		Short: "Print manual setup instructions",
		Long:  "Prints the manual setup instructions for all supported shells without modifying any files.",
		RunE:  runSetupHelp,
	}
}

func runSetupHelp(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine home directory: %w", err)
	}

	ordinDir := filepath.Join(home, ".ordin")
	bashScript := filepath.Join(ordinDir, "ssh_complete.bash")
	zshScript := filepath.Join(ordinDir, "ssh_complete.zsh")
	psScript := filepath.Join(ordinDir, "ssh_complete.ps1")
	psProfile := filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")

	fmt.Fprintln(cmd.OutOrStdout(), "To set up SSH host autocompletion, configure your shell of choice:")
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintln(cmd.OutOrStdout(), "---- BASH ----")
	fmt.Fprintln(cmd.OutOrStdout(), "Add the following to your ~/.bash_profile:")
	fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", markerBegin)
	fmt.Fprintf(cmd.OutOrStdout(), "  . \"%s\"\n", bashScript)
	fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", markerEnd)
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintln(cmd.OutOrStdout(), "---- ZSH ----")
	fmt.Fprintln(cmd.OutOrStdout(), "Add the following to your ~/.zshrc:")
	fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", markerBegin)
	fmt.Fprintf(cmd.OutOrStdout(), "  . \"%s\"\n", zshScript)
	fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", markerEnd)
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintln(cmd.OutOrStdout(), "---- POWERSHELL ----")
	fmt.Fprintln(cmd.OutOrStdout(), "Open your PS profile file with the command: notepad $PROFILE")
	fmt.Fprintf(cmd.OutOrStdout(), "Or open: %s in an editor of your choosing\n", psProfile)
	fmt.Fprintln(cmd.OutOrStdout(), "Add the following to the file:")
	fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", markerBegin)
	fmt.Fprintf(cmd.OutOrStdout(), "  . %s\n", psScript)
	fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", markerEnd)
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintln(cmd.OutOrStdout(), "Note: The marker comments allow 'ssh-autocomplete setup' to detect and update your configuration automatically.")

	return nil
}

func runSetup(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine home directory: %w", err)
	}

	ordinDir := filepath.Join(home, ".ordin")
	if err := os.MkdirAll(ordinDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s directory: %w", ordinDir, err)
	}

	shell := detectShell()
	if shell == "" {
		fmt.Fprintln(cmd.OutOrStdout(), "Could not detect your shell. Run 'ssh-autocomplete setup help' for manual instructions.")
		return nil
	}

	invokeCmd := generateCommand()
	if invokeCmd != "ssh-autocomplete" {
		fmt.Fprintf(cmd.OutOrStdout(), "Dev mode detected, using: %s\n", invokeCmd)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Detected shell: %s\n", shell)

	var scriptPath, profilePath, sourceBlock string

	switch shell {
	case "bash":
		scriptPath = filepath.Join(ordinDir, "ssh_complete.bash")
		profilePath = filepath.Join(home, ".bash_profile")
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			profilePath = filepath.Join(home, ".bashrc")
		}
		sourceBlock = fmt.Sprintf(". \"%s\"", scriptPath)
		if err := os.WriteFile(scriptPath, []byte(bashCompletionScript(invokeCmd)), 0644); err != nil {
			return fmt.Errorf("failed to write completion script: %w", err)
		}

	case "zsh":
		scriptPath = filepath.Join(ordinDir, "ssh_complete.zsh")
		profilePath = filepath.Join(home, ".zshrc")
		sourceBlock = fmt.Sprintf(". \"%s\"", scriptPath)
		if err := os.WriteFile(scriptPath, []byte(zshCompletionScript(invokeCmd)), 0644); err != nil {
			return fmt.Errorf("failed to write completion script: %w", err)
		}

	case "powershell":
		scriptPath = filepath.Join(ordinDir, "ssh_complete.ps1")
		profilePath = detectPowerShellProfile(home)
		sourceBlock = fmt.Sprintf(". \"%s\"", scriptPath)
		if err := os.WriteFile(scriptPath, []byte(powershellCompletionScript(invokeCmd)), 0644); err != nil {
			return fmt.Errorf("failed to write completion script: %w", err)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Completion script written to: %s\n", scriptPath)
	fmt.Fprintf(cmd.OutOrStdout(), "Shell profile: %s\n", profilePath)

	// Check for legacy (unmarked) installs first
	legacyLines := findLegacyLines(profilePath)
	if len(legacyLines) > 0 && !hasMarkerBlock(profilePath) {
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "A previous ssh-autocomplete setup was detected in your profile (without update markers):")
		for _, line := range legacyLines {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", strings.TrimSpace(line))
		}
		fmt.Fprintln(cmd.OutOrStdout(), "")
		if !promptYesNo(cmd, "Remove the old lines and install the new managed version?") {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted. No changes made to your profile.")
			return nil
		}
		if err := removeLegacyLines(profilePath); err != nil {
			return fmt.Errorf("failed to remove legacy lines: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Removed legacy lines.")
	}

	// Check if already installed with markers
	existing := hasMarkerBlock(profilePath)
	if existing {
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "An existing ssh-autocomplete block was found in your profile.")
		if !promptYesNo(cmd, "Update it?") {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted. No changes made to your profile.")
			return nil
		}
		if err := replaceMarkerBlock(profilePath, sourceBlock); err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Profile updated successfully.")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintf(cmd.OutOrStdout(), "The following will be added to %s:\n", profilePath)
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", markerBegin)
		for _, line := range strings.Split(sourceBlock, "\n") {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", line)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", markerEnd)
		fmt.Fprintln(cmd.OutOrStdout(), "")

		if !promptYesNo(cmd, "Proceed?") {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted. No changes made to your profile.")
			return nil
		}
		if err := appendMarkerBlock(profilePath, sourceBlock); err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Profile updated successfully.")
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Restart your shell or source your profile to activate autocompletion.")
	return nil
}

func newUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove SSH host autocompletion from your shell profile",
		Long:  "Removes the ssh-autocomplete block from your shell profile and deletes the completion scripts from ~/.ordin/.",
		RunE:  runUninstall,
	}
}

func runUninstall(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine home directory: %w", err)
	}

	ordinDir := filepath.Join(home, ".ordin")
	shell := detectShell()

	if shell == "" {
		fmt.Fprintln(cmd.OutOrStdout(), "Could not detect your shell. You may need to manually remove the ssh-autocomplete block from your profile.")
		fmt.Fprintln(cmd.OutOrStdout(), "Look for lines between '# BEGIN ssh-autocomplete' and '# END ssh-autocomplete'.")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Detected shell: %s\n", shell)

	var profilePath string
	var scriptFiles []string

	switch shell {
	case "bash":
		profilePath = filepath.Join(home, ".bash_profile")
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			profilePath = filepath.Join(home, ".bashrc")
		}
		scriptFiles = []string{filepath.Join(ordinDir, "ssh_complete.bash")}

	case "zsh":
		profilePath = filepath.Join(home, ".zshrc")
		scriptFiles = []string{filepath.Join(ordinDir, "ssh_complete.zsh")}

	case "powershell":
		profilePath = detectPowerShellProfile(home)
		scriptFiles = []string{filepath.Join(ordinDir, "ssh_complete.ps1")}
	}

	if !hasMarkerBlock(profilePath) {
		// Check for legacy lines even without markers
		legacyLines := findLegacyLines(profilePath)
		if len(legacyLines) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No marker block found, but legacy ssh-autocomplete lines were detected:")
			for _, line := range legacyLines {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", strings.TrimSpace(line))
			}
			fmt.Fprintln(cmd.OutOrStdout(), "")
			if !promptYesNo(cmd, "Remove these lines?") {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}
			if err := removeLegacyLines(profilePath); err != nil {
				return fmt.Errorf("failed to remove legacy lines: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Removed legacy lines from profile.")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "No ssh-autocomplete configuration found in your profile. Nothing to uninstall.")
			return nil
		}
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Will remove the ssh-autocomplete block from: %s\n", profilePath)
		if !promptYesNo(cmd, "Proceed?") {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
			return nil
		}

		if err := removeMarkerBlock(profilePath); err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Removed ssh-autocomplete block from profile.")
	}

	// Remove script files
	for _, f := range scriptFiles {
		if _, err := os.Stat(f); err == nil {
			os.Remove(f)
			fmt.Fprintf(cmd.OutOrStdout(), "Removed: %s\n", f)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Uninstall complete. Restart your shell to apply changes.")
	return nil
}

// detectShell determines the current shell type.
func detectShell() string {
	if runtime.GOOS == "windows" {
		// On Windows, check for PowerShell indicators
		if os.Getenv("PSModulePath") != "" {
			return "powershell"
		}
		return "powershell" // default to powershell on Windows
	}

	// Unix-like: check SHELL env var
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return ""
	}

	base := filepath.Base(shellPath)
	switch {
	case strings.Contains(base, "zsh"):
		return "zsh"
	case strings.Contains(base, "bash"):
		return "bash"
	default:
		return ""
	}
}

// detectPowerShellProfile returns the path to the PowerShell profile.
func detectPowerShellProfile(home string) string {
	// Check for PowerShell Core (pwsh) profile first
	pwshProfile := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	if _, err := os.Stat(pwshProfile); err == nil {
		return pwshProfile
	}
	// Fall back to Windows PowerShell
	return filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
}

// legacyPatterns are strings that indicate an older version of ssh-autocomplete
// was sourced in a profile without using marker comments.
var legacyPatterns = []string{
	"ssh_complete.bash",
	"ssh_complete.zsh",
	"ssh_complete.ps1",
	"ssh-autocomplete",
	".wsm/ssh-autocomplete",
	"_ssh_host_complete",
	"_complete_ssh_hosts",
	"ssh autocomplete generate",
}

// hasMarkerBlock checks if a file contains the ssh-autocomplete marker block.
func hasMarkerBlock(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), markerBegin)
}

// findLegacyLines scans a profile file for lines that look like a previous
// (unmarked) ssh-autocomplete setup. Returns the matching line numbers (1-indexed)
// and the line contents. Only considers lines outside of an existing marker block.
func findLegacyLines(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")
	var matches []string
	inBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == markerBegin {
			inBlock = true
			continue
		}
		if trimmed == markerEnd {
			inBlock = false
			continue
		}
		if inBlock {
			continue
		}
		// Skip comments that aren't source lines
		if trimmed == "" {
			continue
		}
		for _, pattern := range legacyPatterns {
			if strings.Contains(line, pattern) {
				matches = append(matches, line)
				break
			}
		}
	}

	return matches
}

// removeLegacyLines removes lines matching legacy patterns from a profile file.
// Only removes lines outside of marker blocks.
func removeLegacyLines(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var result []string
	inBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == markerBegin {
			inBlock = true
			result = append(result, line)
			continue
		}
		if trimmed == markerEnd {
			inBlock = false
			result = append(result, line)
			continue
		}
		if inBlock {
			result = append(result, line)
			continue
		}

		// Check if this line matches a legacy pattern
		isLegacy := false
		if trimmed != "" {
			for _, pattern := range legacyPatterns {
				if strings.Contains(line, pattern) {
					isLegacy = true
					break
				}
			}
		}
		if !isLegacy {
			result = append(result, line)
		}
	}

	return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}

// appendMarkerBlock appends a marker-delimited block to a file.
func appendMarkerBlock(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	block := fmt.Sprintf("\n%s\n%s\n%s\n", markerBegin, content, markerEnd)
	_, err = f.WriteString(block)
	return err
}

// replaceMarkerBlock replaces an existing marker block in a file with new content.
func replaceMarkerBlock(path, content string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var result []string
	inBlock := false

	for _, line := range lines {
		if strings.TrimSpace(line) == markerBegin {
			inBlock = true
			result = append(result, markerBegin)
			result = append(result, strings.Split(content, "\n")...)
			continue
		}
		if strings.TrimSpace(line) == markerEnd {
			inBlock = false
			result = append(result, markerEnd)
			continue
		}
		if !inBlock {
			result = append(result, line)
		}
	}

	return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}

// removeMarkerBlock removes the marker-delimited block from a file.
func removeMarkerBlock(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var result []string
	inBlock := false

	for _, line := range lines {
		if strings.TrimSpace(line) == markerBegin {
			inBlock = true
			continue
		}
		if strings.TrimSpace(line) == markerEnd {
			inBlock = false
			continue
		}
		if !inBlock {
			result = append(result, line)
		}
	}

	// Trim trailing empty lines that were left from the block removal
	for len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}
	output := strings.Join(result, "\n") + "\n"

	return os.WriteFile(path, []byte(output), 0644)
}

// promptYesNo asks the user a yes/no question and returns true for yes.
func promptYesNo(cmd *cobra.Command, question string) bool {
	fmt.Fprintf(cmd.OutOrStdout(), "%s [y/N]: ", question)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

func bashCompletionScript(invokeCmd string) string {
	return fmt.Sprintf(`# SSH host completion for bash
_ssh_host_complete() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local hosts
    hosts=$(%s list 2>/dev/null)
    COMPREPLY=($(compgen -W "$hosts" -- "$cur"))
    return 0
}

complete -F _ssh_host_complete ssh
complete -F _ssh_host_complete scp
complete -F _ssh_host_complete ssh-copy-id
`, invokeCmd)
}

func zshCompletionScript(invokeCmd string) string {
	return fmt.Sprintf(`# SSH host completion for zsh
_ssh_host_complete() {
    local hosts
    hosts=(${(f)"$(%s list 2>/dev/null)"})
    compadd -a hosts
}

compdef _ssh_host_complete ssh
compdef _ssh_host_complete scp
compdef _ssh_host_complete ssh-copy-id
`, invokeCmd)
}

func powershellCompletionScript(invokeCmd string) string {
	return fmt.Sprintf(`# SSH host completion for PowerShell
Register-ArgumentCompleter -CommandName ssh, scp, sftp -Native -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $hosts = %s list 2>$null
    if ($wordToComplete -match '^(?<user>[-\w/\\]+)@(?<host>[-.\w]+)$') {
        $hosts | Where-Object { $_ -like "$($Matches['host'].ToString())*" } `+"`"+`
            | ForEach-Object { "$($Matches['user'].ToString())@$_" }
    }
    else {
        $hosts | Where-Object { $_ -like "$wordToComplete*" } `+"`"+`
            | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
    }
}
`, invokeCmd)
}
