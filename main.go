package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func getConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("This action is irreversible. Do you want to continue? (yes/no): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		if response == "yes" {
			return true
		} else if response == "no" {
			return false
		} else {
			fmt.Println("Invalid input. Please enter 'yes' or 'no'.")
		}
	}
}

func checkAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}

func waitForExit() {
	fmt.Print("Press any key to exit...")
	bufio.NewReader(os.Stdin).ReadByte()
}

// Ejecuta el desinstalador principal de Visual Studio
func runUninstaller() error {
	desinstalador := `C:\Program Files (x86)\Microsoft Visual Studio\Installer\vs_installer.exe`
	args := []string{"/uninstall", "/norestart"}

	fmt.Println("Ejecutando el desinstalador de Visual Studio...")
	cmd := exec.Command(desinstalador, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error ejecutando el desinstalador: %v\n", err)
		return err
	}

	fmt.Printf("Salida del desinstalador: %s\n", output)
	return nil
}

// Ejecuta el programa InstallCleanup.exe
func runCleanup() error {
	cleanup := `C:\Program Files (x86)\Microsoft Visual Studio\Installer\InstallCleanup.exe`

	fmt.Println("Ejecutando InstallCleanup.exe...")
	cmd := exec.Command(cleanup)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error ejecutando InstallCleanup: %v\n", err)
		return err
	}

	fmt.Printf("Salida de InstallCleanup: %s\n", output)
	return nil
}

// Manual folder cleanup
func cleanFolders() {
	folders := []string{
		`C:\Program Files (x86)\Microsoft Visual Studio`,
		`C:\Program Files\Microsoft Visual Studio`,
		`C:\ProgramData\Microsoft\visualstudio\packages`,
	}

	tempDir := os.TempDir()
	folders = append(folders, tempDir)

	fmt.Println("Starting folder cleanup...")

	for _, folder := range folders {
		fmt.Printf("Attempting to delete: %s\n", folder)
		if err := os.RemoveAll(folder); err != nil {
			fmt.Printf("Could not delete folder: %s. Error: %v\n", folder, err)
		} else {
			fmt.Printf("Folder deleted: %s\n", folder)
		}
	}
}

// Manual registry keys cleanup
func cleanRegistryKeys() {
	keys := []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\Policies\Microsoft\VisualStudio\Setup`},
		{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\VisualStudio\Setup`},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\VisualStudio\Setup`},
	}

	fmt.Println("Starting registry keys cleanup...")

	for _, key := range keys {
		fmt.Printf("Attempting to delete key: %s\n", key.path)
		if err := registry.DeleteKey(key.root, key.path); err != nil {
			if syscall.ERROR_FILE_NOT_FOUND == err.(syscall.Errno) {
				fmt.Printf("Key not found: %s\n", key.path)
			} else {
				fmt.Printf("Could not delete key: %s. Error: %v\n", key.path, err)
			}
		} else {
			fmt.Printf("Key deleted: %s\n", key.path)
		}
	}
}

// Remove shortcuts from Start Menu and Desktop
func cleanShortcuts() {
	shortcuts := []string{
		filepath.Join(os.Getenv("APPDATA"), `Microsoft\Windows\Start Menu\Programs\Visual Studio Installer.lnk`),
		filepath.Join(os.Getenv("APPDATA"), `Microsoft\Windows\Start Menu\Programs\Visual Studio 2022.lnk`),
		filepath.Join(os.Getenv("USERPROFILE"), `Desktop\Visual Studio Installer.lnk`),
		filepath.Join(os.Getenv("USERPROFILE"), `Desktop\Visual Studio 2022.lnk`),
	}

	fmt.Println("Starting shortcuts cleanup...")

	for _, shortcut := range shortcuts {
		fmt.Printf("Attempting to delete shortcut: %s\n", shortcut)
		if err := os.Remove(shortcut); err != nil {
			fmt.Printf("Could not delete shortcut: %s. Error: %v\n", shortcut, err)
		} else {
			fmt.Printf("Shortcut deleted: %s\n", shortcut)
		}
	}
}

// Perform manual cleanup
func manualCleanup() {
	fmt.Println("Starting manual cleanup of Visual Studio...")
	cleanFolders()
	cleanRegistryKeys()
	cleanShortcuts()
	fmt.Println("Manual cleanup completed.")
}

func main() {
	if !checkAdmin() {
		fmt.Println("Run as admin is required")
		waitForExit()
		return
	}

	if getConfirmation() {
		if err := runUninstaller(); err != nil {
			fmt.Println("No se pudo ejecutar el desinstalador principal. Continuando con InstallCleanup...")
		}

		if err := runCleanup(); err != nil {
			fmt.Println("Ocurrió un error al ejecutar InstallCleanup. Continuando con limpieza manual...")
		}

		manualCleanup()
	} else {
		fmt.Println("Operation cancelled.")
		waitForExit()
	}
}
