//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func registrySet(key, name, value string) error {
	args := []string{"add", key}
	if name == "" {
		args = append(args, "/ve")
	} else {
		args = append(args, "/v", name)
	}
	args = append(args, "/t", "REG_SZ", "/d", value, "/f")

	cmd := exec.Command("reg.exe", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("reg.exe: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func registerSparrowFileAssociations() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return err
	}

	quotedExe := `"` + exe + `"`
	openCommand := quotedExe + ` "%1"`
	icon := exe + ",0"

	// Per-user registration: no administrator privileges required.
	if err := registrySet(`HKCU\Software\Classes\SparrowHTML`, "", "Sparrow HTML Document"); err != nil {
		return err
	}
	if err := registrySet(`HKCU\Software\Classes\SparrowHTML\DefaultIcon`, "", icon); err != nil {
		return err
	}
	if err := registrySet(`HKCU\Software\Classes\SparrowHTML\shell\open\command`, "", openCommand); err != nil {
		return err
	}

	if err := registrySet(`HKCU\Software\Classes\SparrowPDF`, "", "Sparrow PDF Document"); err != nil {
		return err
	}
	if err := registrySet(`HKCU\Software\Classes\SparrowPDF\DefaultIcon`, "", icon); err != nil {
		return err
	}
	if err := registrySet(`HKCU\Software\Classes\SparrowPDF\shell\open\command`, "", openCommand); err != nil {
		return err
	}

	for _, ext := range []string{".html", ".htm"} {
		if err := registrySet(`HKCU\Software\Classes\`+ext, "", "SparrowHTML"); err != nil {
			return err
		}
	}
	if err := registrySet(`HKCU\Software\Classes\.pdf`, "", "SparrowPDF"); err != nil {
		return err
	}

	// Also register the application itself so Windows can display its icon
	// in "Open with" and default-app selection.
	appKey := `HKCU\Software\Classes\Applications\` + filepath.Base(exe)
	if err := registrySet(appKey, "FriendlyAppName", "Sparrow Browser"); err != nil {
		return err
	}
	if err := registrySet(appKey+`\DefaultIcon`, "", icon); err != nil {
		return err
	}
	if err := registrySet(appKey+`\shell\open\command`, "", openCommand); err != nil {
		return err
	}

	// Advertise Sparrow as a browser candidate in Windows Default Apps.
	capabilities := `HKCU\Software\Clients\StartMenuInternet\Sparrow\Capabilities`
	if err := registrySet(capabilities, "ApplicationName", "Sparrow Browser"); err != nil {
		return err
	}
	if err := registrySet(capabilities, "ApplicationDescription", "Sparrow Browser"); err != nil {
		return err
	}
	if err := registrySet(capabilities+`\FileAssociations`, ".html", "SparrowHTML"); err != nil {
		return err
	}
	if err := registrySet(capabilities+`\FileAssociations`, ".htm", "SparrowHTML"); err != nil {
		return err
	}
	if err := registrySet(capabilities+`\FileAssociations`, ".pdf", "SparrowPDF"); err != nil {
		return err
	}
	if err := registrySet(capabilities+`\URLAssociations`, "http", "SparrowURL"); err != nil {
		return err
	}
	if err := registrySet(capabilities+`\URLAssociations`, "https", "SparrowURL"); err != nil {
		return err
	}
	if err := registrySet(`HKCU\Software\Classes\SparrowURL`, "", "Sparrow URL"); err != nil {
		return err
	}
	if err := registrySet(`HKCU\Software\Classes\SparrowURL\DefaultIcon`, "", icon); err != nil {
		return err
	}
	if err := registrySet(`HKCU\Software\Classes\SparrowURL\shell\open\command`, "", openCommand); err != nil {
		return err
	}
	if err := registrySet(`HKCU\Software\RegisteredApplications`, "Sparrow Browser", `Software\Clients\StartMenuInternet\Sparrow\Capabilities`); err != nil {
		return err
	}

	// Refresh Explorer's icon/association cache.
	_ = exec.Command("ie4uinit.exe", "-show").Run()
	return nil
}

func unregisterSparrowFileAssociations() error {
	// We only remove the Sparrow-specific ProgIDs/application registration.
	for _, key := range []string{
		`HKCU\Software\Classes\SparrowHTML`,
		`HKCU\Software\Classes\SparrowPDF`,
		`HKCU\Software\Classes\SparrowURL`,
		`HKCU\Software\Clients\StartMenuInternet\Sparrow`,
	} {
		_ = exec.Command("reg.exe", "delete", key, "/f").Run()
	}
	_ = exec.Command("reg.exe", "delete", `HKCU\Software\RegisteredApplications`, "/v", "Sparrow Browser", "/f").Run()
	return nil
}
