package utils

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	vdpa "github.com/k8snetworkplumbingwg/govdpa/pkg/kvdpa"
)

// VdpaProvider is a wrapper type over go-vdpa library
type VdpaProvider interface {
	GetVdpaDeviceByPci(pciAddr string) (vdpa.VdpaDevice, error)
}

type defaultVdpaProvider struct {
}

var vdpaProvider VdpaProvider = &defaultVdpaProvider{}

// SetVdpaProviderInst method would be used by unit tests in other packages
func SetVdpaProviderInst(inst VdpaProvider) {
	vdpaProvider = inst
}

// GetVdpaProvider will be invoked by functions in other packages that would need access to the vdpa library methods.
func GetVdpaProvider() VdpaProvider {
	return vdpaProvider
}

type mgmtDev struct {
	busName string
	devName string
}

// BusName returns the MgmtDev's bus name
func (m *mgmtDev) BusName() string {
	return m.busName
}

// BusName returns the MgmtDev's device name
func (m *mgmtDev) DevName() string {
	return m.devName
}

func (m *mgmtDev) Name() string {
	if m.busName != "" {
		return strings.Join([]string{m.busName, m.devName}, "/")
	}
	return m.devName
}

type vhostVdpa struct {
	name string
	path string
}

// Name returns the vhost device's name
func (v *vhostVdpa) Name() string {
	return v.name
}

// Name returns the vhost device's path
func (v *vhostVdpa) Path() string {
	return v.path
}

type vdpaDeviceHack struct {
	name      string // vdpa:0000:65:00.2
	busName   string // optional
	devName   string // 0000:65:00.2
	vHostName string // vhost-vdpa-0
	vHostPath string // /dev/vhost-vdpa-0
}

func (v *vdpaDeviceHack) Driver() string {
	return "vhost_vdpa"
}
func (v *vdpaDeviceHack) Name() string {
	return v.name
}

func (v *vdpaDeviceHack) MgmtDev() vdpa.MgmtDev {
	return &mgmtDev{busName: v.busName, devName: v.devName}
}
func (v *vdpaDeviceHack) VirtioNet() vdpa.VirtioNet {
	return nil
}
func (v *vdpaDeviceHack) VhostVdpa() vdpa.VhostVdpa {
	return &vhostVdpa{name: v.vHostName, path: v.vHostPath}
}
func (v *vdpaDeviceHack) ParentDevicePath() (string, error) {
	return v.devName, nil
}

func (defaultVdpaProvider) GetVdpaDeviceByPci(pciAddr string) (vdpa.VdpaDevice, error) {
	if pciAddr == "0000:65:00.2" {
		return &vdpaDeviceHack{name: "vdpa:0000:65:00.2", busName: "", devName: pciAddr, vHostName: "vhost-vdpa-0", vHostPath: "/dev/vhost-vdpa-0"}, nil
	} else if pciAddr == "0000:65:00.3" {
		return &vdpaDeviceHack{name: "vdpa:0000:65:00.3", busName: "", devName: pciAddr, vHostName: "vhost-vdpa-1", vHostPath: "/dev/vhost-vdpa-1"}, nil
	}

	// the govdpa library requires the pci address to include the "pci/" prefix
	fullPciAddr := "pci/" + pciAddr
	vdpaDevices, err := vdpa.GetVdpaDevicesByPciAddress(fullPciAddr)
	if err != nil {
		return nil, err
	}
	numVdpaDevices := len(vdpaDevices)
	if numVdpaDevices == 0 {
		return nil, fmt.Errorf("no vdpa device associated to pciAddress %s", pciAddr)
	}
	if numVdpaDevices > 1 {
		glog.Infof("More than one vDPA device found for pciAddress %s, returning the first one", pciAddr)
	}
	return vdpaDevices[0], nil
}
