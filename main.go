package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Bubble Tea item and selector model for interactive selection
type item string

func (i item) FilterValue() string { return string(i) }

type selectorModel struct {
	list     list.Model
	selected string
}

func (m *selectorModel) Init() tea.Cmd {
	return nil
}

func (m *selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if sel := m.list.SelectedItem(); sel != nil {
				m.selected = sel.FilterValue()
			}
			return m, tea.Quit
		case "ctrl+c", "q":
			m.selected = ""
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *selectorModel) View() string {
	return m.list.View()
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, i list.Item) {
	selected := index == m.Index()
	s := fmt.Sprintf("  %s", i.FilterValue())
	if selected {
		s = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true).Render("> " + i.FilterValue())
	}
	fmt.Fprintln(w, s)
}

var style = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	PaddingTop(1).
	PaddingLeft(2).
	Width(30)

var debugMode bool

func debugPrintln(a ...interface{}) {
	if debugMode {
		fmt.Println(a...)
	}
}

func debugPrintf(format string, a ...interface{}) {
	if debugMode {
		fmt.Printf(format, a...)
	}
}

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
		Use:   "wand [group] [machine]",
		Short: "A stylish CLI tool using lipgloss",
		Long:  style.Render("wand: A stylish CLI tool using lipgloss"),
		Args:  cobra.MaximumNArgs(2),
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
			var groupMap = make(map[string][]map[string]interface{})
			for _, entry := range data {
				for k, v := range entry {
					groups = append(groups, k)
					if arr, ok := v.([]interface{}); ok {
						for _, m := range arr {
							if mMap, ok := m.(map[string]interface{}); ok {
								groupMap[k] = append(groupMap[k], mMap)
							}
						}
					}
				}
			}
			if len(groups) == 0 {
				fmt.Println(style.Render("No groups found in config."))
				return
			}

			var selectedGroup string
			var selectedMachineName string
			var machines []map[string]interface{}

			if len(args) > 0 {
				// Group provided as arg
				selectedGroup = args[0]
				var found bool
				for _, g := range groups {
					if g == selectedGroup {
						found = true
						break
					}
				}
				if !found {
					fmt.Println(style.Render("Group not found: "), selectedGroup)
					return
				}
				machines = groupMap[selectedGroup]
				if len(machines) == 0 {
					fmt.Println(style.Render("No machines found in group."))
					return
				}
				if len(args) > 1 {
					// Machine provided as arg
					selectedMachineName = args[1]
					var foundMachine bool
					for _, m := range machines {
						if m["name"] == selectedMachineName {
							foundMachine = true
							break
						}
					}
					if !foundMachine {
						fmt.Println(style.Render("Machine not found in group: "), selectedMachineName)
						return
					}
				} else {
					// No machine arg, show machine selector
					machineItems := make([]list.Item, len(machines))
					for i, m := range machines {
						machineItems[i] = item(m["name"].(string))
					}
					machineList := list.New(machineItems, itemDelegate{}, 40, 12)
					machineList.Title = "Select a machine"
					p2 := tea.NewProgram(&selectorModel{list: machineList})
					m2, err := p2.Run()
					if err != nil {
						fmt.Println("Error running machine selector:", err)
						return
					}
					selectedMachineName = m2.(*selectorModel).selected
					if selectedMachineName == "" {
						fmt.Println("No machine selected.")
						return
					}
				}
			} else {
				// No args, show group selector
				groupItems := make([]list.Item, len(groups))
				for i, g := range groups {
					groupItems[i] = item(g)
				}
				groupList := list.New(groupItems, itemDelegate{}, 40, 10)
				groupList.Title = "Select a group"
				p := tea.NewProgram(&selectorModel{list: groupList})
				m, err := p.Run()
				if err != nil {
					fmt.Println("Error running group selector:", err)
					return
				}
				selectedGroup = m.(*selectorModel).selected
				if selectedGroup == "" {
					fmt.Println("No group selected.")
					return
				}
				machines = groupMap[selectedGroup]
				if len(machines) == 0 {
					fmt.Println(style.Render("No machines found in group."))
					return
				}
				// Show machine selector
				machineItems := make([]list.Item, len(machines))
				for i, m := range machines {
					machineItems[i] = item(m["name"].(string))
				}
				machineList := list.New(machineItems, itemDelegate{}, 40, 12)
				machineList.Title = "Select a machine"
				p2 := tea.NewProgram(&selectorModel{list: machineList})
				m2, err := p2.Run()
				if err != nil {
					fmt.Println("Error running machine selector:", err)
					return
				}
				selectedMachineName = m2.(*selectorModel).selected
				if selectedMachineName == "" {
					fmt.Println("No machine selected.")
					return
				}
			}

			// Find the selected machine
			var selectedMachine map[string]interface{}
			for _, m := range machines {
				if m["name"] == selectedMachineName {
					selectedMachine = m
					break
				}
			}
			if selectedMachine == nil {
				fmt.Println("Machine not found.")
				return
			}
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

				// Check if Microsoft Remote Desktop is running
				isRunning := false
				debugPrintln("[DEBUG] Checking if Microsoft Remote Desktop or Windows App is running...")
				isRunning = false
				checkCmd1 := exec.Command("pgrep", "-f", "Microsoft Remote Desktop")
				checkCmd2 := exec.Command("pgrep", "-f", "Windows App")
				if err := checkCmd1.Run(); err == nil {
					isRunning = true
					debugPrintln("[DEBUG] Microsoft Remote Desktop is already running.")
				} else if err := checkCmd2.Run(); err == nil {
					isRunning = true
					debugPrintln("[DEBUG] Windows App is already running.")
				} else {
					debugPrintln("[DEBUG] Neither Microsoft Remote Desktop nor Windows App is running.")
					// Print process list for debugging
					if debugMode {
						psCmd := exec.Command("ps", "aux")
						grepCmd := exec.Command("egrep", `Microsoft Remote Desktop|Windows App`)
						psOut, _ := psCmd.StdoutPipe()
						grepCmd.Stdin = psOut
						grepCmd.Stdout = os.Stdout
						_ = psCmd.Start()
						_ = grepCmd.Start()
						_ = psCmd.Wait()
						_ = grepCmd.Wait()
					}
				}
				if !isRunning {
					debugPrintln("[DEBUG] Launching Microsoft Remote Desktop and Windows App (background, no focus)...")
					// Try both apps
					_ = exec.Command("open", "-gj", "/Applications/Microsoft Remote Desktop.app").Run()
					_ = exec.Command("open", "-gj", "/Applications/Windows App.app").Run()
					// Wait for the app to launch, retry up to 20 times (15 seconds total)
					for i := 0; i < 20; i++ {
						time.Sleep(750 * time.Millisecond)
						debugPrintf("[DEBUG] Checking if Microsoft Remote Desktop or Windows App is running (attempt %d)...\n", i+1)
						checkCmd1 := exec.Command("pgrep", "-f", "Microsoft Remote Desktop")
						checkCmd2 := exec.Command("pgrep", "-f", "Windows App")
						if err := checkCmd1.Run(); err == nil {
							isRunning = true
							debugPrintln("[DEBUG] Microsoft Remote Desktop is now running.")
							break
						} else if err := checkCmd2.Run(); err == nil {
							isRunning = true
							debugPrintln("[DEBUG] Windows App is now running.")
							break
						}
					}
					if !isRunning {
						debugPrintln("[DEBUG] Neither Microsoft Remote Desktop nor Windows App started after waiting.")
					}
				}

				// Extra wait to ensure the app is fully initialized
				if isRunning {
					debugPrintln("[DEBUG] App detected as running, waiting 2 seconds to ensure it is ready...")
					time.Sleep(2 * time.Second)
				}

				fmt.Println(style.Render("Opening RDP connection..."))
				c := exec.Command("open", tmpFile.Name())
				c.Stdin = os.Stdin
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				err = c.Run()
				if err != nil {
					fmt.Println(style.Render("RDP connection failed:"), err)
				} else {
					debugPrintln("[DEBUG] RDP file opened successfully.")
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
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Show debug output")
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