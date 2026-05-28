package daemon

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install spotifyd binary for your system",
	RunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS == "darwin" {
			fmt.Print("On macOS, spotifyd will be installed via Homebrew. Proceed? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(strings.TrimSpace(response)) == "y" {
				fmt.Println("Installing spotifyd via brew...")
				installCmd := exec.Command("brew", "install", "spotifyd")
				installCmd.Stdout = os.Stdout
				installCmd.Stderr = os.Stderr
				if err := installCmd.Run(); err != nil {
					return fmt.Errorf("brew install failed: %w", err)
				}
				fmt.Println("spotifyd installed successfully via Homebrew.")
				return nil
			}
			return fmt.Errorf("installation aborted")
		}

		if runtime.GOOS != "linux" {
			return fmt.Errorf("unsupported system: %s", runtime.GOOS)
		}

		archMap := map[string]string{"amd64": "x86_64", "arm64": "aarch64", "arm": "armv7"}
		archName := archMap[runtime.GOARCH]
		if archName == "" {
			return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
		}

		targetAsset := fmt.Sprintf("spotifyd-linux-%s-default.tar.gz", archName)
		fmt.Printf("Looking for release: %s\n", targetAsset)

		resp, err := http.Get("https://api.github.com/repos/Spotifyd/spotifyd/releases/latest")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var release struct {
			Assets []struct {
				Name               string `json:"name"`
				BrowserDownloadURL string `json:"browser_download_url"`
			} `json:"assets"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return err
		}

		var downloadURL string
		for _, asset := range release.Assets {
			if asset.Name == targetAsset {
				downloadURL = asset.BrowserDownloadURL
				break
			}
		}

		if downloadURL == "" {
			return fmt.Errorf("could not find asset %s in latest release", targetAsset)
		}

		binDir := filepath.Join(config.DataDir(), "bin")
		os.MkdirAll(binDir, 0755)

		fmt.Printf("Downloading %s...\n", downloadURL)
		assetResp, err := http.Get(downloadURL)
		if err != nil {
			return err
		}
		defer assetResp.Body.Close()

		gr, err := gzip.NewReader(assetResp.Body)
		if err != nil {
			return err
		}
		defer gr.Close()

		tr := tar.NewReader(gr)
		var installed bool
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, "spotifyd") {
				targetFile := filepath.Join(binDir, "spotifyd")
				f, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
				if err != nil {
					return err
				}
				if _, err := io.Copy(f, tr); err != nil {
					f.Close()
					return err
				}
				f.Close()
				fmt.Printf("Successfully installed spotifyd to %s\n", targetFile)
				installed = true
				break
			}
		}

		if !installed {
			return fmt.Errorf("spotifyd binary not found in tarball")
		}

		installLinuxDeps()

		return nil
	},
}

func installLinuxDeps() {
	var pm string
	var args []string

	if _, err := exec.LookPath("apt-get"); err == nil {
		pm = "apt-get"
		args = []string{"install", "-y", "libasound2-dev", "libssl-dev", "libpulse-dev", "libdbus-1-dev"}
	} else if _, err := exec.LookPath("dnf"); err == nil {
		pm = "dnf"
		args = []string{"install", "-y", "alsa-lib-devel", "make", "gcc"}
	} else if _, err := exec.LookPath("pacman"); err == nil {
		pm = "pacman"
		args = []string{"-S", "--noconfirm", "base-devel", "alsa-lib", "libogg", "libpulse", "dbus"}
	} else if _, err := exec.LookPath("zypper"); err == nil {
		pm = "zypper"
		args = []string{"install", "-y", "alsa-devel", "make", "gcc"}
	} else {
		fmt.Println("Could not detect package manager. Please manually install dependencies.")
		return
	}

	fmt.Printf("\nspotifyd requires dependencies to run. Detected package manager: %s\n", pm)
	fmt.Printf("Command: sudo %s %s\n", pm, strings.Join(args, " "))
	fmt.Print("Proceed with installing dependencies? (y/n): ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(strings.TrimSpace(response)) == "y" {
		cmd := exec.Command("sudo", append([]string{pm}, args...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to install dependencies: %v\n", err)
		} else {
			fmt.Println("Dependencies installed successfully.")
		}
	} else {
		fmt.Println("Skipping dependency installation.")
	}
}

func init() {
	DaemonCmd.AddCommand(installCmd)
}
