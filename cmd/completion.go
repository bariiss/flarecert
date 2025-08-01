package cmd

import (
	"os"
	"strings"

	"github.com/bariiss/flarecert/internal/config"
	"github.com/bariiss/flarecert/internal/dns"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

  $ source <(flarecert completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ flarecert completion bash > /etc/bash_completion.d/flarecert
  # macOS:
  $ flarecert completion bash > $(brew --prefix)/etc/bash_completion.d/flarecert

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ flarecert completion zsh > "${fpath[1]}/_flarecert"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ flarecert completion fish | source

  # To load completions for each session, execute once:
  $ flarecert completion fish > ~/.config/fish/completions/flarecert.fish

PowerShell:

  PS> flarecert completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> flarecert completion powershell > flarecert.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

// GetDomainCompletions returns domain suggestions from Cloudflare zones
func GetDomainCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Try to load config and get zones
	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	provider, err := dns.NewCloudflareProvider(cfg.CloudflareAPIToken, cfg.CloudflareEmail, cfg.DNSTimeout, false)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	zones, err := provider.ListZones()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, zone := range zones {
		if zone.Status == "active" {
			// Add the zone itself
			if strings.HasPrefix(zone.Name, toComplete) {
				suggestions = append(suggestions, zone.Name)
			}
			// Add wildcard version
			wildcard := "*." + zone.Name
			if strings.HasPrefix(wildcard, toComplete) {
				suggestions = append(suggestions, wildcard)
			}
			// Add www version
			www := "www." + zone.Name
			if strings.HasPrefix(www, toComplete) {
				suggestions = append(suggestions, www)
			}
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
