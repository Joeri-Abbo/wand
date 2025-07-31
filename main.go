package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var style = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	PaddingTop(1).
	PaddingLeft(2).
	Width(30)

func main() {
	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit the config file in JSON",
		Run: func(cmd *cobra.Command, args []string) {
			configPath := filepath.Join(os.Getenv("HOME"), ".wand", "config.json")
			if _, err := os.Stat(filepath.Dir(configPath)); os.IsNotExist(err) {
				os.MkdirAll(filepath.Dir(configPath), 0700)
			}
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			c := exec.Command(editor, configPath)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			err := c.Run()
			if err != nil {
				fmt.Println(style.Render("Failed to open editor:"), err)
			}
		},
	}

	rootCmd := &cobra.Command{
		Use:   "wand",
		Short: "A stylish CLI tool using lipgloss",
		Long:  style.Render("wand: A stylish CLI tool using lipgloss"),
		Run: func(cmd *cobra.Command, args []string) {
			configPath := filepath.Join(os.Getenv("HOME"), ".wand", "config.json")
			f, err := os.Open(configPath)
			if err != nil {
				fmt.Println(style.Render("Config file not found. Use 'wand edit' to create it."))
				return
			}
			defer f.Close()
			var data []map[string]interface{}
			dec := json.NewDecoder(f)
			if err := dec.Decode(&data); err != nil {
				fmt.Println(style.Render("Failed to parse config file."))
				return
			}
			var groups []string
			for _, entry := range data {
				for k := range entry {
					groups = append(groups, k)
				}
			}
			if len(groups) == 0 {
				fmt.Println(style.Render("No groups found in config."))
				return
			}
			groupStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).PaddingLeft(2)
			fmt.Println(style.Render("Select a group:"))
			for i, g := range groups {
				fmt.Printf("%s %s\n", groupStyle.Render(fmt.Sprintf("[%d]", i+1)), g)
			}
			fmt.Print(style.Render("Enter number: "))
			var sel int
			_, err = fmt.Scanln(&sel)
			if err != nil || sel < 1 || sel > len(groups) {
				fmt.Println(style.Render("Invalid selection."))
				return
			}
			selectedGroup := groups[sel-1]
			// Find machines in selected group
			var machines []map[string]interface{}
			for _, entry := range data {
				if v, ok := entry[selectedGroup]; ok {
					if arr, ok := v.([]interface{}); ok {
						for _, m := range arr {
							if mMap, ok := m.(map[string]interface{}); ok {
								machines = append(machines, mMap)
							}
						}
					}
				}
			}
			if len(machines) == 0 {
				fmt.Println(style.Render("No machines found in group."))
				return
			}
			machineStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).PaddingLeft(2)
			fmt.Println(style.Render("Available machines:"))
			for i, m := range machines {
				name, _ := m["name"].(string)
				fmt.Printf("%s %s\n", machineStyle.Render(fmt.Sprintf("[%d]", i+1)), name)
			}
			fmt.Print(style.Render("Select a machine by number: "))
			var msel int
			_, err = fmt.Scanln(&msel)
			if err != nil || msel < 1 || msel > len(machines) {
				fmt.Println(style.Render("Invalid machine selection."))
				return
			}
			selectedMachine := machines[msel-1]
		   name, _ := selectedMachine["name"].(string)
		   user, _ := selectedMachine["user"].(string)
		   host, _ := selectedMachine["host"].(string)
		   connType, _ := selectedMachine["connection"].(string)
		   if connType == "rdp" {
			   // Generate a temporary .rdp file
			   tmpFile, err := os.CreateTemp("", "wand-*.rdp")
			   if err != nil {
				   fmt.Println(style.Render("Failed to create temp RDP file:"), err)
				   return
			   }
			   defer os.Remove(tmpFile.Name())
			   rdpContent := fmt.Sprintf("full address:s:%s\nusername:s:%s\n", host, user)
			   if _, err := tmpFile.WriteString(rdpContent); err != nil {
				   fmt.Println(style.Render("Failed to write RDP file:"), err)
				   tmpFile.Close()
				   return
			   }
			   tmpFile.Close()
			   fmt.Println(style.Render("Opening RDP connection..."))
			   c := exec.Command("open", tmpFile.Name())
			   c.Stdin = os.Stdin
			   c.Stdout = os.Stdout
			   c.Stderr = os.Stderr
			   err = c.Run()
			   if err != nil {
				   fmt.Println(style.Render("RDP connection failed:"), err)
			   }
			   return
		   }
		   // Default to SSH
		   var identityFile string
		   if ids, ok := selectedMachine["identityFile"].([]interface{}); ok && len(ids) > 0 {
			   if id, ok := ids[0].(string); ok && id != "" {
				   identityFile = id
			   }
		   }
		   fmt.Println(style.Render("Selected machine:"), name)
		   // Build ssh command
		   sshArgs := []string{}
		   // Handle jump host
		   if usesJump, ok := selectedMachine["uses_jumpHost"].(bool); ok && usesJump {
			   if jumpHosts, ok := selectedMachine["jumpHost"].([]interface{}); ok && len(jumpHosts) > 0 {
				   var jumpSpecs []string
				   var jumpIdentities []string
				   for _, jh := range jumpHosts {
					   if jhName, ok := jh.(string); ok {
						   for _, m := range machines {
							   if m["name"] == jhName {
								   jUser, _ := m["user"].(string)
								   jHost, _ := m["host"].(string)
								   if jHost != "" {
									   if jUser != "" {
										   jumpSpecs = append(jumpSpecs, fmt.Sprintf("%s@%s", jUser, jHost))
									   } else {
										   jumpSpecs = append(jumpSpecs, jHost)
									   }
								   }
								   // Add jump host identityFile if present
								   if ids, ok := m["identityFile"].([]interface{}); ok && len(ids) > 0 {
									   if id, ok := ids[0].(string); ok && id != "" {
										   jumpIdentities = append(jumpIdentities, os.ExpandEnv(id))
									   }
								   }
							   }
						   }
					   }
				   }
				   if len(jumpSpecs) > 0 {
					   sshArgs = append(sshArgs, "-J", jumpSpecs[0])
				   }
				   // Add -i for each jump host identity file
				   for _, jid := range jumpIdentities {
					   sshArgs = append(sshArgs, "-i", jid)
				   }
			   }
		   }
		   if identityFile != "" {
			   sshArgs = append(sshArgs, "-i", os.ExpandEnv(identityFile))
		   }
		   var dest string
		   if user != "" && host != "" {
			   dest = fmt.Sprintf("%s@%s", user, host)
		   } else if host != "" {
			   dest = host
		   }
		   if dest != "" {
			   sshArgs = append(sshArgs, dest)
		   }
		   fmt.Println(style.Render("Connecting via SSH..."))
		   c := exec.Command("ssh", sshArgs...)
		   c.Stdin = os.Stdin
		   c.Stdout = os.Stdout
		   c.Stderr = os.Stderr
		   err = c.Run()
		   if err != nil {
			   fmt.Println(style.Render("SSH connection failed:"), err)
		   }
		},
	}
	rootCmd.AddCommand(editCmd)
	rootCmd.Execute()
}

// containsShellSpecial returns true if the string contains shell-special characters
func containsShellSpecial(s string) bool {
	for _, c := range s {
		if c == ' ' || c == '\t' || c == '\'' || c == '"' || c == '$' || c == '`' || c == '\\' {
			return true
		}
	}
	return false
}

// escapeSingleQuotes escapes single quotes for shell
func escapeSingleQuotes(s string) string {
	// Replace every ' with '\''
	// This is the standard way to escape single quotes in POSIX shells
	return stringReplaceAll(s, "'", "'\\''")
}

// stringReplaceAll is a helper for Go < 1.12 compatibility
func stringReplaceAll(s, old, new string) string {
	// Use strings.ReplaceAll if available, otherwise fallback to Replace
	// (for Go 1.12+)
	// This is a simplified version for this use case
	for {
		idx := indexOf(s, old)
		if idx < 0 {
			break
		}
		s = s[:idx] + new + s[idx+len(old):]
	}
	return s
}

// indexOf returns the index of the first instance of substr in s, or -1 if not present
func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
