package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/ready-to-release/eac/src/cli/internal/version"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags
var Version string

// init runs before the root command
func init() {
	RootCmd.AddCommand(versionCmd)

	// Get build time from executable file modification time
	buildTime := ""
	if exe, err := os.Executable(); err == nil {
		if stat, err := os.Stat(exe); err == nil {
			buildTime = stat.ModTime().Format(time.RFC3339)
		}
	}

	// Extract build info
	commit := ""
	timestamp := ""
	modified := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		commit = getSettingValue(info, "vcs.revision")
		timestamp = getSettingValue(info, "vcs.time")
		modified = getSettingValue(info, "vcs.modified")
	}

	// Set version with injected Version variable taking priority
	version.SetVersion(Version, timestamp, commit, buildTime, modified)
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of r2r CLI",
	Long:  `All software has versions. This is r2r CLI's`,
	Run: func(cmd *cobra.Command, args []string) {
		info := version.GetInfo()

		// Use the Version variable, defaulting to "undefined" if not set
		v := info.Version
		if v == "" {
			v = "undefined"
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Version:   %s\n", v)
		fmt.Fprintf(cmd.OutOrStdout(), "Time:      %s\n", info.Timestamp)
		fmt.Fprintf(cmd.OutOrStdout(), "BuildTime: %s\n", info.BuildTime)
		fmt.Fprintf(cmd.OutOrStdout(), "Revision:  %s%s\n", info.Commit,
			func() string {
				isModified, err := strconv.ParseBool(info.Modified)
				if err != nil {
					return ""
				}
				if isModified {
					return " (modified)"
				}
				return ""
			}(),
		)
	},
}

func getSettingValue(info *debug.BuildInfo, key string) string {
	for _, setting := range info.Settings {
		if setting.Key == key {
			return setting.Value
		}
	}
	return ""
}
