package main

/*
#cgo linux pkg-config: gtk+-3.0
#include <gtk/gtk.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/auriora/onemount/cmd/common"
	"github.com/auriora/onemount/internal/ui"
	"github.com/auriora/onemount/internal/ui/systemd"
	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/logging"
	"github.com/coreos/go-systemd/v22/unit"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	flag "github.com/spf13/pflag"
)

// setupLogging configures the logger based on the configuration
func setupLogging(config *common.Config) error {
	// Set the global log level
	logging.SetGlobalLevel(common.StringToLevel(config.LogLevel))

	// Configure the log output
	var output io.Writer
	switch config.LogOutput {
	case "STDOUT":
		output = os.Stdout
	case "STDERR":
		output = os.Stderr
	default:
		// Open the log file
		file, err := os.OpenFile(config.LogOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			logging.Error().Err(err).Str("path", config.LogOutput).Msg("Failed to open log file, falling back to STDOUT")
			output = os.Stdout
		} else {
			output = file
		}
	}

	// Set up the logger with console formatting
	logging.DefaultLogger = logging.New(logging.NewConsoleWriterWithOptions(output, "15:04:05"))
	return nil
}

func usage() {
	fmt.Printf(`onemount-launcher - Manage and configure onemount mountpoints

Usage: onemount-launcher [options]

Valid options:
`)
	flag.PrintDefaults()
}

func main() {
	logLevel := flag.StringP("log", "l", "",
		"Set logging level/verbosity for the filesystem. "+
			"Can be one of: fatal, error, warn, info, debug, trace")
	logOutput := flag.StringP("log-output", "o", "",
		"Set the output location for logs. "+
			"Can be STDOUT, STDERR, or a file path. Default is STDOUT.")
	cacheDir := flag.StringP("cache-dir", "c", "",
		"Change the default cache directory used by onemount. "+
			"Will be created if it does not already exist.")
	configPath := flag.StringP("config-file", "f", common.DefaultConfigPath(),
		"A YAML-formatted configuration file used by onemount.")
	versionFlag := flag.BoolP("version", "v", false, "Display program version.")
	help := flag.BoolP("help", "h", false, "Displays this help message.")
	flag.Usage = usage
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *versionFlag {
		fmt.Println("onemount-launcher", common.Version())
		os.Exit(0)
	}

	// loading config can emit an unformatted log message, so we do this first with a basic logger
	logging.DefaultLogger = logging.New(logging.NewConsoleWriterWithOptions(os.Stderr, "15:04:05"))

	// command line options override config options
	config := common.LoadConfig(*configPath)

	// Now configure the logger based on the configuration
	err := setupLogging(config)
	if err != nil {
		fmt.Printf("Failed to setup logging: %s\n", err)
		return
	}
	if *cacheDir != "" {
		config.CacheDir = *cacheDir
	}
	if *logLevel != "" {
		config.LogLevel = *logLevel
	}
	if *logOutput != "" {
		config.LogOutput = *logOutput
	}

	logging.SetGlobalLevel(common.StringToLevel(config.LogLevel))

	logging.Info().Msgf("onemount-launcher %s", common.Version())

	app, err := gtk.ApplicationNew("com.github.auriora.onemount", glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		logging.Fatal().Err(err).Msg("Could not create application.")
	}
	app.Connect("activate", func(application *gtk.Application) {
		activateCallback(application, config, *configPath)
	})
	os.Exit(app.Run(nil))
}

// activateCallback is what actually sets up the application
func activateCallback(app *gtk.Application, config *common.Config, configPath string) {
	window, _ := gtk.ApplicationWindowNew(app)
	window.SetDefaultSize(550, 400)

	header, _ := gtk.HeaderBarNew()
	header.SetShowCloseButton(true)
	header.SetTitle("OneMount")
	window.SetTitlebar(header)

	err := window.SetIconFromFile("/usr/share/icons/onemount/onemount.svg")
	if err != nil {
		logging.Warn().Err(err).Msg("Could not find logo.")
	}

	listbox, _ := gtk.ListBoxNew()
	window.Add(listbox)

	switches := make(map[string]*gtk.Switch)

	mountpointBtn, _ := gtk.ButtonNewFromIconName("list-add-symbolic", gtk.ICON_SIZE_BUTTON)
	mountpointBtn.SetTooltipText("Add a new OneDrive account.")
	mountpointBtn.Connect("clicked", func(button *gtk.Button) {
		mount := ui.DirChooser("Select a mountpoint")
		if !ui.MountpointIsValid(mount) {
			logging.Error().Str("mountpoint", mount).
				Msg("Mountpoint was not valid (or user cancelled the operation). " +
					"Mountpoint must be an empty directory.")
			if mount != "" {
				ui.Dialog(
					"Mountpoint was not valid, mountpoint must be an empty directory "+
						"(there might be hidden files).", gtk.MESSAGE_ERROR, window)
			}
			return
		}

		escapedMount := unit.UnitNamePathEscape(mount)
		systemdUnit := systemd.TemplateUnit(systemd.OneMountServiceTemplate, escapedMount)
		logging.Info().
			Str("mountpoint", mount).
			Str("systemdUnit", systemdUnit).
			Msg("Creating mountpoint.")

		if err := systemd.UnitSetActive(systemdUnit, true); err != nil {
			logging.Error().Err(err).Msg("Failed to start unit.")
			return
		}

		row, sw := newMountRow(*config, mount)
		switches[mount] = sw
		listbox.Insert(row, -1)

		go xdgOpenDir(mount)
	})
	header.PackStart(mountpointBtn)

	// create a menubutton and assign a popover menu
	menuBtn, _ := gtk.MenuButtonNew()
	icon, _ := gtk.ImageNewFromIconName("open-menu-symbolic", gtk.ICON_SIZE_BUTTON)
	menuBtn.SetImage(icon)
	popover, _ := gtk.PopoverNew(menuBtn)
	menuBtn.SetPopover(popover)
	popover.SetBorderWidth(8)

	// add buttons to menu
	popoverBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	settings, _ := gtk.ModelButtonNew()
	settings.SetLabel("Settings")
	settings.Connect("clicked", func(button *gtk.ModelButton) {
		newSettingsDialog(config, configPath, window)
	})
	popoverBox.PackStart(settings, false, true, 0)

	// print version and link to repo
	about, _ := gtk.ModelButtonNew()
	about.SetLabel("About")
	about.Connect("clicked", func(button *gtk.ModelButton) {
		aboutDialog, _ := gtk.AboutDialogNew()
		aboutDialog.SetProgramName("OneMount Launcher")
		aboutDialog.SetAuthors([]string{"Bruce Cherrington", "https://github.com/auriora"})
		aboutDialog.SetWebsite("https://github.com/auriora/onemount")
		aboutDialog.SetWebsiteLabel("github.com/auriora/onemount")
		aboutDialog.SetVersion(fmt.Sprintf("onemount %s", common.Version()))
		aboutDialog.SetLicenseType(gtk.LICENSE_GPL_3_0)
		logo, err := gtk.ImageNewFromFile("/usr/share/icons/onemount/onemount-128.png")
		if err != nil {
			logging.Warn().Err(err).Msg("Could not find logo.")
		} else {
			aboutDialog.SetLogo(logo.GetPixbuf())
		}
		aboutDialog.SetTransientFor(window)
		aboutDialog.Connect("response", aboutDialog.Destroy)
		aboutDialog.Run()
	})
	popoverBox.PackStart(about, false, true, 0)

	popoverBox.ShowAll()
	popover.Add(popoverBox)
	popover.SetPosition(gtk.POS_BOTTOM)
	header.PackEnd(menuBtn)

	mounts := ui.GetKnownMounts(config.CacheDir)
	for _, mount := range mounts {
		mount = unit.UnitNamePathUnescape(mount)

		logging.Info().Str("mount", mount).Msg("Found existing mount.")

		row, sw := newMountRow(*config, mount)
		switches[mount] = sw
		listbox.Insert(row, -1)
	}

	listbox.Connect("row-activated", func() {
		row := listbox.GetSelectedRow()
		mount, _ := row.GetName()
		unitName := systemd.TemplateUnit(systemd.OneMountServiceTemplate,
			unit.UnitNamePathEscape(mount))

		logging.Debug().
			Str("mount", mount).
			Str("unit", unitName).
			Str("signal", "row-activated").
			Msg("")

		active, _ := systemd.UnitIsActive(unitName)
		if !active {
			err := systemd.UnitSetActive(unitName, true)
			if err != nil {
				logging.Error().
					Err(err).
					Str("unit", unitName).
					Msg("Could not set unit state to active.")
			}

		}
		switches[mount].SetActive(true)

		go xdgOpenDir(mount)
	})

	window.ShowAll()
}

// xdgOpenDir opens a folder in the user's default file browser.
// Should be invoked as a goroutine to not block the main app.
func xdgOpenDir(mount string) {
	logging.Debug().Str("dir", mount).Msg("Opening directory.")
	if mount == "" || !ui.PollUntilAvail(mount, -1) {
		logging.Error().
			Str("dir", mount).
			Msg("Either directory was invalid or exceeded timeout waiting for fs to become available.")
		return
	}
	cURI := C.CString("file://" + mount)
	C.g_app_info_launch_default_for_uri(cURI, nil, nil)
	C.free(unsafe.Pointer(cURI))
}

// newMountRow constructs a new ListBoxRow with the controls for an individual mountpoint.
// mount is the path to the new mountpoint.
func newMountRow(config common.Config, mount string) (*gtk.ListBoxRow, *gtk.Switch) {
	row, _ := gtk.ListBoxRowNew()
	row.SetSelectable(true)
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	row.Add(box)

	escapedMount := unit.UnitNamePathEscape(mount)
	unitName := systemd.TemplateUnit(systemd.OneMountServiceTemplate, escapedMount)

	driveName, err := common.GetXDGVolumeInfoName(filepath.Join(mount, ".xdg-volume-info"))
	if err != nil {
		logging.Error().
			Err(err).
			Str("mountpoint", mount).
			Msg("Could not determine user-specified acccount name.")
	}

	tildePath := ui.EscapeHome(mount)
	accountName, err := graph.GetAccountName(config.CacheDir, escapedMount)
	label, _ := gtk.LabelNew("")
	if driveName != "" {
		// we have a user-assigned name for the user's drive
		label.SetMarkup(fmt.Sprintf("%s <span style=\"italic\" weight=\"light\">(%s)</span>    ",
			driveName, tildePath,
		))
	} else if err == nil {
		// fs isn't mounted, so just use user principal name from AAD
		label, _ = gtk.LabelNew("")
		label.SetMarkup(fmt.Sprintf("%s <span style=\"italic\" weight=\"light\">(%s)</span>    ",
			accountName, tildePath,
		))
	} else {
		// something went wrong and all we have is the mountpoint name
		logging.Error().
			Err(err).
			Str("mountpoint", mount).
			Msg("Could not determine user principal name.")
		label, _ = gtk.LabelNew(tildePath)
	}
	box.PackStart(label, false, false, 5)

	// a switch to start/stop the mountpoint
	mountToggle, _ := gtk.SwitchNew()
	active, err := systemd.UnitIsActive(unitName)
	if err == nil {
		mountToggle.SetActive(active)
	} else {
		logging.Error().Err(err).Msg("Error checking unit active state.")
	}
	mountToggle.SetTooltipText("Mount or unmount selected OneDrive account")
	mountToggle.SetVAlign(gtk.ALIGN_CENTER)
	mountToggle.Connect("state-set", func() {
		logging.Info().
			Str("signal", "state-set").
			Str("mount", mount).
			Str("unitName", unitName).
			Bool("active", mountToggle.GetActive()).
			Msg("Changing systemd unit active state.")
		err := systemd.UnitSetActive(unitName, mountToggle.GetActive())
		if err != nil {
			logging.Error().
				Err(err).
				Str("unit", unitName).
				Msg("Could not change systemd unit active state.")
		}
	})

	mountpointSettingsBtn, _ := gtk.MenuButtonNew()
	icon, _ := gtk.ImageNewFromIconName("emblem-system-symbolic", gtk.ICON_SIZE_BUTTON)
	mountpointSettingsBtn.SetImage(icon)
	popover, _ := gtk.PopoverNew(mountpointSettingsBtn)
	mountpointSettingsBtn.SetPopover(popover)
	popover.SetBorderWidth(8)
	popoverBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)

	if accountName != "" {
		accountLabel, _ := gtk.LabelNew(accountName)
		popoverBox.Add(accountLabel)
	}
	// rename the mount by rewriting the .xdg-volume-info file
	renameMountpointEntry, _ := gtk.EntryNew()
	renameMountpointEntry.SetTooltipText("The label that your file browser uses for this drive")
	renameMountpointEntry.SetText(driveName)
	// runs on enter
	renameMountpointEntry.Connect("activate", func(entry *gtk.Entry) {
		newName, err := entry.GetText()
		ctx := logging.DefaultLogger.With().
			Str("signal", "clicked").
			Str("mount", mount).
			Str("unitName", unitName).
			Str("oldName", driveName).
			Str("newName", newName).
			Logger()
		if err != nil {
			ctx.Error().Err(err).Msg("Failed to get new drive name.")
			return
		}
		if driveName == newName {
			ctx.Info().Msg("New name is same as old name, ignoring.")
			return
		}
		ctx.Info().
			Msg("Renaming mount.")
		popover.GrabFocus()

		err = systemd.UnitSetActive(unitName, true)
		if err != nil {
			ctx.Error().Err(err).Msg("Failed to start mount for rename.")
			return
		}
		mountToggle.SetActive(true)

		if ui.PollUntilAvail(mount, -1) {
			xdgVolumeInfo := common.TemplateXDGVolumeInfo(newName)
			driveName = newName

			// Write the XDG volume info file to update the mount display name
			xdgPath := filepath.Join(mount, ".xdg-volume-info")
			err = os.WriteFile(xdgPath, []byte(xdgVolumeInfo), 0644)
			if err != nil {
				ctx.Error().
					Err(err).
					Str("path", xdgPath).
					Str("content", xdgVolumeInfo).
					Msg("Failed to write XDG volume info file - mount may be read-only or not fully ready")
				// Don't return here - the rename operation can still succeed even if XDG file write fails
			} else {
				ctx.Info().
					Str("path", xdgPath).
					Str("newName", newName).
					Msg("Successfully updated XDG volume info file")
			}
		} else {
			ctx.Error().Msg("Mount never became ready - cannot update display name")
		}
		// update label in UI now
		label.SetMarkup(fmt.Sprintf("%s <span style=\"italic\" weight=\"light\">(%s)</span>    ",
			newName, tildePath,
		))

		ui.Dialog("Drive rename will take effect on next filesystem start.", gtk.MESSAGE_INFO, nil)
		ctx.Info().Msg("Drive rename will take effect on next filesystem start.")
	})
	popoverBox.Add(renameMountpointEntry)

	separator, _ := gtk.SeparatorMenuItemNew()
	popoverBox.Add(separator)

	// create a button to enable/disable the mountpoint
	unitEnabledBtn, _ := gtk.CheckButtonNewWithLabel("Start Drive on Login")
	unitEnabledBtn.SetTooltipText("Start this drive automatically when you login")
	enabled, err := systemd.UnitIsEnabled(unitName)
	if err == nil {
		unitEnabledBtn.SetActive(enabled)
	} else {
		logging.Error().Err(err).Msg("Error checking unit enabled state.")
	}
	unitEnabledBtn.Connect("toggled", func() {
		logging.Info().
			Str("signal", "toggled").
			Str("mount", mount).
			Str("unitName", unitName).
			Bool("enabled", unitEnabledBtn.GetActive()).
			Msg("Changing systemd unit enabled state.")
		err := systemd.UnitSetEnabled(unitName, unitEnabledBtn.GetActive())
		if err != nil {
			logging.Error().
				Err(err).
				Str("unit", unitName).
				Msg("Could not change systemd unit enabled state.")
		}
	})
	popoverBox.PackStart(unitEnabledBtn, false, true, 0)

	// button to delete the mount
	deleteMountpointBtn, _ := gtk.ModelButtonNew()
	deleteMountpointBtn.SetLabel("Remove Drive")
	deleteMountpointBtn.SetTooltipText("Remove OneDrive account from local computer")
	deleteMountpointBtn.Connect("clicked", func(button *gtk.ModelButton) {
		logging.Trace().
			Str("signal", "clicked").
			Str("mount", mount).
			Str("unitName", unitName).
			Msg("Request to delete drive.")

		if ui.CancelDialog(nil, "<span weight=\"bold\">Remove drive?</span>",
			"This will remove all data for this drive from your local computer. "+
				"It can also be used to \"reset\" the drive to its original state.") {
			logging.Info().
				Str("signal", "clicked").
				Str("mount", mount).
				Str("unitName", unitName).
				Msg("Deleting mount.")
			err := systemd.UnitSetEnabled(unitName, false)
			if err != nil {
				logging.Error().Err(err).Msg("Could not disable unit.")
				return
			}

			err = systemd.UnitSetActive(unitName, false)
			if err != nil {
				logging.Error().Err(err).Msg("Could not deactivate unit.")
				return
			}

			cachedir, _ := os.UserCacheDir()
			err = os.RemoveAll(fmt.Sprintf("%s/onemount/%s/", cachedir, escapedMount))
			if err != nil {
				logging.Error().Err(err).Msg("Could not remove mount.")
				return
			}

			row.Destroy()
		}
	})
	popoverBox.PackStart(deleteMountpointBtn, false, true, 0)

	// ok show everything in the mount settings menu
	popoverBox.ShowAll()
	popover.Add(popoverBox)
	popover.SetPosition(gtk.POS_BOTTOM)

	// add all widgets to row in the right order
	box.PackEnd(mountpointSettingsBtn, false, false, 0)
	box.PackEnd(mountToggle, false, false, 0)

	// name is used by "row-activated" callback
	row.SetName(mount)
	row.ShowAll()
	return row, mountToggle
}

func newSettingsDialog(config *common.Config, configPath string, parent gtk.IWindow) {
	const offset = 15

	settingsDialog, _ := gtk.DialogNew()
	settingsDialog.SetResizable(false)
	settingsDialog.SetTitle("Settings")

	// log level settings
	settingsRowLog, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, offset)
	logLevelLabel, _ := gtk.LabelNew("Log Level")
	settingsRowLog.PackStart(logLevelLabel, false, false, 0)

	logLevelSelector, _ := gtk.ComboBoxTextNew()
	for i, entry := range common.LogLevels() {
		logLevelSelector.AppendText(entry)
		if entry == config.LogLevel {
			logLevelSelector.SetActive(i)
		}
	}
	logLevelSelector.Connect("changed", func(box *gtk.ComboBoxText) {
		config.LogLevel = box.GetActiveText()
		logging.Debug().
			Str("newLevel", config.LogLevel).
			Msg("Log level changed.")
		logging.SetGlobalLevel(common.StringToLevel(config.LogLevel))
		err := config.WriteConfig(configPath)
		if err != nil {
			logging.Error().Err(err).Msg("Could not write config.")
			return
		}
	})
	settingsRowLog.PackEnd(logLevelSelector, false, false, 0)

	// log output settings
	settingsRowLogOutput, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, offset)
	logOutputLabel, _ := gtk.LabelNew("Log Output")
	settingsRowLogOutput.PackStart(logOutputLabel, false, false, 0)

	logOutputEntry, _ := gtk.EntryNew()
	logOutputEntry.SetText(config.LogOutput)
	logOutputEntry.SetTooltipText("Set to STDOUT, STDERR, or a file path")
	logOutputEntry.SetSizeRequest(200, 0)
	logOutputEntry.Connect("activate", func(entry *gtk.Entry) {
		newOutput, err := entry.GetText()
		if err != nil {
			logging.Error().Err(err).Msg("Could not get log output text.")
			return
		}
		if newOutput == config.LogOutput {
			return // No change
		}
		config.LogOutput = newOutput
		logging.Debug().
			Str("newOutput", config.LogOutput).
			Msg("Log output changed.")
		err = config.WriteConfig(configPath)
		if err != nil {
			logging.Error().Err(err).Msg("Could not write config.")
			return
		}
		// Apply the change immediately
		if err := setupLogging(config); err != nil {
			logging.Error().Err(err).Msg("Failed to update logging configuration")
		}
	})
	settingsRowLogOutput.PackEnd(logOutputEntry, false, false, 0)

	// cache dir settings
	settingsRowCacheDir, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, offset)
	cacheDirLabel, _ := gtk.LabelNew("Cache Directory")
	settingsRowCacheDir.PackStart(cacheDirLabel, false, false, 0)

	cacheDirPicker, _ := gtk.ButtonNew()
	cacheDirPicker.SetLabel(ui.EscapeHome(config.CacheDir))
	cacheDirPicker.SetSizeRequest(200, 0)
	cacheDirPicker.Connect("clicked", func(button *gtk.Button) {
		oldPath, _ := button.GetLabel()
		oldPath = ui.UnescapeHome(oldPath)
		path := ui.DirChooser("Select an empty directory to use for storage")
		if !ui.CancelDialog(settingsDialog, "Remount all drives?", "") {
			return
		}
		logging.Warn().
			Str("oldPath", oldPath).
			Str("newPath", path).
			Msg("All active drives will be remounted to move cache directory.")

		// actually perform the stop+move op
		isMounted := make([]string, 0)
		for _, mount := range ui.GetKnownMounts(oldPath) {
			unitName := systemd.TemplateUnit(systemd.OneMountServiceTemplate, mount)
			logging.Info().
				Str("mount", mount).
				Str("unit", unitName).
				Msg("Disabling mount.")
			if mounted, _ := systemd.UnitIsActive(unitName); mounted {
				isMounted = append(isMounted, unitName)
			}

			err := systemd.UnitSetActive(unitName, false)
			if err != nil {
				ui.Dialog("Could not disable mount: "+err.Error(),
					gtk.MESSAGE_ERROR, settingsDialog)
				logging.Error().
					Err(err).
					Str("mount", mount).
					Str("unit", unitName).
					Msg("Could not disable mount.")
				return
			}

			err = os.Rename(filepath.Join(oldPath, mount), filepath.Join(path, mount))
			if err != nil {
				ui.Dialog("Could not move cache for mount: "+err.Error(),
					gtk.MESSAGE_ERROR, settingsDialog)
				logging.Error().
					Err(err).
					Str("mount", mount).
					Str("unit", unitName).
					Msg("Could not move cache for mount.")
				return
			}
		}

		// remount drives that were mounted before
		for _, unitName := range isMounted {
			err := systemd.UnitSetActive(unitName, true)
			if err != nil {
				logging.Error().
					Err(err).
					Str("unit", unitName).
					Msg("Failed to restart unit.")
			}
		}

		// all done
		config.CacheDir = path
		err := config.WriteConfig(configPath)
		if err != nil {
			logging.Error().Err(err).Msg("Failed to write config.")
			return
		}
		button.SetLabel(path)
	})
	settingsRowCacheDir.PackEnd(cacheDirPicker, false, false, 0)

	// assemble rows
	settingsDialogBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, offset)
	settingsDialogBox.SetBorderWidth(offset)
	settingsDialogBox.PackStart(settingsRowLog, true, true, 0)
	settingsDialogBox.PackStart(settingsRowLogOutput, true, true, 0)
	settingsDialogBox.PackStart(settingsRowCacheDir, true, true, 0)

	contentArea, err := settingsDialog.GetContentArea()
	if err != nil {
		logging.Error().Err(err).Msg("Failed to get settings dialog content area.")
		return
	}

	contentArea.Add(settingsDialogBox)

	settingsDialog.SetModal(true)
	settingsDialog.SetTransientFor(parent)
	settingsDialog.ShowAll()
}
