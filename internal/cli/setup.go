package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/KhizerRehan/qarz-cli/internal/config"
	"github.com/KhizerRehan/qarz-cli/internal/gauth"
)

func newSetupCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Walk the one-time Google Calendar setup",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			reader := bufio.NewReader(app.inputOrStdin())

			credentialsPath, err := config.CredentialsPath()
			if err != nil {
				return err
			}
			tokenPath, err := config.TokenPath()
			if err != nil {
				return err
			}

			fmt.Fprintln(app.Out, "qarz needs its own Google Calendar access. Five steps, three of them in your browser.")
			fmt.Fprintln(app.Out)

			switch _, err := os.Stat(credentialsPath); {
			case os.IsNotExist(err):
				if err := walkCredentials(app, reader, credentialsPath); err != nil {
					return err
				}
			case err != nil:
				return fmt.Errorf("check %s: %w", credentialsPath, err)
			default:
				fmt.Fprintf(app.Out, "Found existing client credentials at %s, keeping them.\n\n", credentialsPath)
			}

			if gauth.HasToken(tokenPath) {
				again, err := confirm(app, reader, "You are already signed in. Sign in again?", false)
				if err != nil {
					return err
				}
				if !again {
					fmt.Fprintln(app.Out, "Keeping the existing sign-in.")
				} else if err := gauth.Login(cmd.Context(), credentialsPath, tokenPath, app.Out); err != nil {
					return err
				}
			} else if err := gauth.Login(cmd.Context(), credentialsPath, tokenPath, app.Out); err != nil {
				return err
			}

			fmt.Fprintln(app.Out)
			cfg, err := askConfig(app, reader)
			if err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}

			path, err := config.ConfigPath()
			if err != nil {
				return err
			}

			fmt.Fprintf(app.Out, "\nSetup complete. Settings are in %s and the Google token in %s.\n", path, tokenPath)
			return nil
		},
	}
}

func walkCredentials(app *App, reader *bufio.Reader, credentialsPath string) error {
	steps := []struct {
		title string
		url   string
		body  string
	}{
		{
			title: "Create a Google Cloud project",
			url:   "https://console.cloud.google.com/projectcreate",
			body:  "Name it anything you like. qarz is the only thing that will use it.",
		},
		{
			title: "Enable the Google Calendar API for that project",
			url:   "https://console.cloud.google.com/apis/library/calendar-json.googleapis.com",
			body:  "Make sure the new project is selected in the picker at the top, then press Enable.",
		},
		{
			title: "Configure the OAuth consent screen",
			url:   "https://console.cloud.google.com/apis/credentials/consent",
			body:  "Choose External, fill in the app name and your own email, and add yourself as a test user.",
		},
		{
			title: "Create an OAuth client of type Desktop app",
			url:   "https://console.cloud.google.com/apis/credentials",
			body:  "Credentials, then Create credentials, then OAuth client ID. Pick Desktop app, create it, and download the JSON.",
		},
	}

	for i, step := range steps {
		fmt.Fprintf(app.Out, "Step %d of 5 — %s\n", i+1, step.title)
		fmt.Fprintf(app.Out, "  %s\n", step.url)
		fmt.Fprintf(app.Out, "  %s\n", step.body)
		if _, err := prompt(app, reader, "  Press Enter once that is done"); err != nil {
			return err
		}
		fmt.Fprintln(app.Out)
	}

	fmt.Fprintln(app.Out, "Step 5 of 5 — point qarz at the JSON you just downloaded")
	for {
		answer, err := prompt(app, reader, "  Path to the downloaded client JSON")
		if err != nil {
			return err
		}

		answer = expandHome(strings.Trim(answer, "\"' "))
		if answer == "" {
			fmt.Fprintln(app.Out, "  A path is required.")
			continue
		}

		raw, err := os.ReadFile(answer)
		if err != nil {
			fmt.Fprintf(app.Out, "  Could not read that file: %v\n", err)
			continue
		}
		if err := os.WriteFile(credentialsPath, raw, 0o600); err != nil {
			return fmt.Errorf("write %s: %w", credentialsPath, err)
		}
		if !gauth.CredentialsUsable(credentialsPath) {
			fmt.Fprintln(app.Out, "  That file is not a Google desktop client JSON. Download it again from the Credentials page.")
			continue
		}

		fmt.Fprintf(app.Out, "  Stored at %s\n\n", credentialsPath)
		return nil
	}
}

func askConfig(app *App, reader *bufio.Reader) (config.Config, error) {
	existing, err := config.Load()
	if err != nil {
		return config.Config{}, err
	}

	cfg := existing
	fields := []struct {
		label   string
		current *string
	}{
		{label: "Calendar to project onto", current: &cfg.CalendarID},
		{label: "Timezone", current: &cfg.Timezone},
		{label: "Currency", current: &cfg.Currency},
	}

	for _, field := range fields {
		answer, err := prompt(app, reader, fmt.Sprintf("%s [%s]", field.label, *field.current))
		if err != nil {
			return config.Config{}, err
		}
		if answer != "" {
			*field.current = answer
		}
	}

	if _, err := cfg.Location(); err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

func prompt(app *App, reader *bufio.Reader, label string) (string, error) {
	fmt.Fprintf(app.Out, "%s: ", label)

	line, err := reader.ReadString('\n')
	if err != nil && !strings.Contains(err.Error(), "EOF") {
		return "", fmt.Errorf("read your answer: %w", err)
	}
	return strings.TrimSpace(line), nil
}

func confirm(app *App, reader *bufio.Reader, question string, defaultYes bool) (bool, error) {
	hint := "y/N"
	if defaultYes {
		hint = "Y/n"
	}

	answer, err := prompt(app, reader, fmt.Sprintf("%s [%s]", question, hint))
	if err != nil {
		return false, err
	}

	switch strings.ToLower(answer) {
	case "":
		return defaultYes, nil
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return home + strings.TrimPrefix(path, "~")
}

func (a *App) inputOrStdin() io.Reader {
	if a.In != nil {
		return a.In
	}
	return os.Stdin
}
