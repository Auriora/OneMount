package systemd

import (
	"errors"
	"fmt"
	"strings"

	dbus "github.com/godbus/dbus/v5"
)

const (
	OneMountServiceTemplate = "onemount@.service"
	SystemdBusName          = "org.freedesktop.systemd1"
	SystemdObjectPath       = "/org/freedesktop/systemd1"
)

// TemplateUnit templates a unit name as systemd would
func TemplateUnit(template, instance string) string {
	// Replace forward slashes with hyphens to ensure valid systemd unit name
	escapedInstance := strings.ReplaceAll(instance, "/", "-")
	return strings.Replace(template, "@.", fmt.Sprintf("@%s.", escapedInstance), 1)
}

// UntemplateUnit reverses the templating done by SystemdTemplateUnit
func UntemplateUnit(unit string) (string, error) {
	var start, end int
	for i, char := range unit {
		if char == '@' {
			start = i + 1
		}
		if char == '.' {
			break
		}
		end = i + 1
	}
	if start == 0 {
		return "", errors.New("not a systemd templated unit")
	}
	return unit[start:end], nil
}

// UnitIsActive returns true if the unit is currently active or activating
func UnitIsActive(unit string) (bool, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return false, err
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			// Log the close error but don't override the main error
			// This is a best-effort cleanup
		}
	}()

	obj := conn.Object(SystemdBusName, SystemdObjectPath)
	call := obj.Call("org.freedesktop.systemd1.Manager.GetUnit", 0, unit)
	if call.Err != nil {
		return false, call.Err
	}
	var unitPath string
	if err = call.Store(&unitPath); err != nil {
		return false, err
	}

	obj = conn.Object(SystemdBusName, dbus.ObjectPath(unitPath))
	property, err := obj.GetProperty("org.freedesktop.systemd1.Unit.ActiveState")
	if err != nil {
		return false, err
	}
	var active string
	if err = property.Store(&active); err != nil {
		return false, err
	}
	// Consider both "active" and "activating" states as active
	return active == "active" || active == "activating", nil
}

func UnitSetActive(unit string, active bool) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			// Log the close error but don't override the main error
			// This is a best-effort cleanup
		}
	}()

	obj := conn.Object(SystemdBusName, SystemdObjectPath)
	if active {
		return obj.Call("org.freedesktop.systemd1.Manager.StartUnit", 0, unit, "replace").Err
	}
	return obj.Call("org.freedesktop.systemd1.Manager.StopUnit", 0, unit, "replace").Err
}

// UnitIsEnabled returns true if a particular systemd unit is enabled.
func UnitIsEnabled(unit string) (bool, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return false, err
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			// Log the close error but don't override the main error
			// This is a best-effort cleanup
		}
	}()

	var state string
	obj := conn.Object(SystemdBusName, SystemdObjectPath)
	err = obj.Call(
		"org.freedesktop.systemd1.Manager.GetUnitFileState", 0, unit,
	).Store(&state)
	if err != nil {
		return false, err
	}
	return state == "enabled", nil
}

// UnitSetEnabled sets a systemd unit to enabled/disabled.
func UnitSetEnabled(unit string, enabled bool) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			// Log the close error but don't override the main error
			// This is a best-effort cleanup
		}
	}()

	units := []string{unit}
	obj := conn.Object(SystemdBusName, SystemdObjectPath)
	if enabled {
		return obj.Call(
			"org.freedesktop.systemd1.Manager.EnableUnitFiles", 0, units, false, true,
		).Err
	}
	return obj.Call(
		"org.freedesktop.systemd1.Manager.DisableUnitFiles", 0, units, false,
	).Err
}
