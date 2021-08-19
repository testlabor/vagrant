package core

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/ptypes"
	"github.com/mitchellh/mapstructure"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vagrant-plugin-sdk/component"
	"github.com/hashicorp/vagrant-plugin-sdk/core"
	"github.com/hashicorp/vagrant-plugin-sdk/proto/vagrant_plugin_sdk"
	"github.com/hashicorp/vagrant/internal/server/proto/vagrant_server"
)

type Machine struct {
	*Target
	machine *vagrant_server.Target_Machine
	logger  hclog.Logger
}

// Close implements core.Machine
func (m *Machine) Close() (err error) {
	return
}

// ID implements core.Machine
func (m *Machine) ID() (id string, err error) {
	return m.machine.Id, nil
}

// SetID implements core.Machine
func (m *Machine) SetID(value string) (err error) {
	m.machine.Id = value
	return m.SaveMachine()
}

func (m *Machine) Box() (b *core.Box, err error) {
	var result core.Box
	return &result, mapstructure.Decode(m.machine.Box, &result)
}

// Guest implements core.Machine
func (m *Machine) Guest() (g core.Guest, err error) {
	guests, err := m.project.basis.typeComponents(m.ctx, component.GuestType)
	if err != nil {
		return
	}

	var result core.Guest
	var result_name string

	for name, g := range guests {
		guest := g.Value.(core.Guest)
		detected, err := guest.Detect(m.toTarget())
		if err != nil {
			m.logger.Error("guest error on detection check",
				"plugin", name,
				"type", "Guest",
				"error", err)

			continue
		}
		if result == nil {
			if detected {
				result = guest
				result_name = name
			}
			continue
		}

		gp, err := guest.Parents()
		if err != nil {
			m.logger.Error("failed to get parents from guest",
				"plugin", name,
				"type", "Guest",
				"error", err)

			continue
		}

		rp, err := result.Parents()
		if err != nil {
			m.logger.Error("failed to get parents from guest",
				"plugin", result_name,
				"type", "Guest",
				"error", err)

			continue
		}

		if len(gp) > len(rp) {
			result = guest
			result_name = name
		}
	}

	if result == nil {
		return nil, fmt.Errorf("failed to detect guest plugin for current platform")
	}

	m.logger.Info("guest detection complete",
		"name", result_name)

	return result, nil
}

func (m *Machine) Inspect() (printable string, err error) {
	name, err := m.Name()
	provider, err := m.Provider()
	printable = "#<" + reflect.TypeOf(m).String() + ": " + name + " (" + reflect.TypeOf(provider).String() + ")>"
	return
}

// Reload implements core.Machine
func (m *Machine) Reload() (err error) {
	// TODO
	return
}

// ConnectionInfo implements core.Machine
func (m *Machine) ConnectionInfo() (info *core.ConnectionInfo, err error) {
	// TODO: need Vagrantfile
	return
}

// MachineState implements core.Machine
func (m *Machine) MachineState() (state *core.MachineState, err error) {
	var result core.MachineState
	return &result, mapstructure.Decode(m.machine.State, &result)
}

// SetMachineState implements core.Machine
func (m *Machine) SetMachineState(state *core.MachineState) (err error) {
	var st *vagrant_plugin_sdk.Args_Target_Machine_State
	mapstructure.Decode(state, &st)
	m.machine.State = st
	return m.SaveMachine()
}

func (m *Machine) UID() (userId string, err error) {
	return m.machine.Uid, nil
}

// SyncedFolders implements core.Machine
func (m *Machine) SyncedFolders() (folders []core.SyncedFolder, err error) {
	// TODO: need Vagrantfile
	return
}

func (m *Machine) SaveMachine() (err error) {
	m.logger.Debug("saving machine to db", "machine", m.machine.Id)
	m.target.Record, err = ptypes.MarshalAny(m.machine)
	if err != nil {
		return nil
	}
	return m.Save()
}

func (m *Machine) toTarget() core.Target {
	return m
}

var _ core.Machine = (*Machine)(nil)
var _ core.Target = (*Machine)(nil)
