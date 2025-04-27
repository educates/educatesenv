package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed educates versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		binDir := Config.Local.Folder
		files, err := os.ReadDir(binDir)
		if err != nil {
			return fmt.Errorf("error reading bin directory: %w", err)
		}

		// Find the active symlink target
		active := ""
		symlinkPath := filepath.Join(binDir, "educates")
		if linkTarget, err := os.Readlink(symlinkPath); err == nil {
			// If the symlink is relative, resolve it to absolute
			if !filepath.IsAbs(linkTarget) {
				linkTarget = filepath.Join(binDir, linkTarget)
			}
			active = linkTarget
		}

		var versions []string
		versionToPath := make(map[string]string)
		for _, f := range files {
			name := f.Name()
			if strings.HasPrefix(name, "educates-") && !f.IsDir() {
				ver := strings.TrimPrefix(name, "educates-")
				fullPath := filepath.Join(binDir, name)
				versions = append(versions, ver)
				versionToPath[ver] = fullPath
			}
		}
		sort.Sort(sort.Reverse(sort.StringSlice(versions)))

		if len(versions) == 0 {
			fmt.Println("No educates versions installed.")
			return nil
		}

		for _, ver := range versions {
			mark := " "
			if versionToPath[ver] == active {
				mark = "*"
			}
			fmt.Printf("%s %s\n", mark, ver)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
