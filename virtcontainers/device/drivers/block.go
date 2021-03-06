// Copyright (c) 2017-2018 Intel Corporation
// Copyright (c) 2018 Huawei Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package drivers

import (
	"path/filepath"

	"github.com/kata-containers/runtime/virtcontainers/device/api"
	"github.com/kata-containers/runtime/virtcontainers/device/config"
	"github.com/kata-containers/runtime/virtcontainers/utils"
)

const maxDevIDSize = 31

// BlockDevice refers to a block storage device implementation.
type BlockDevice struct {
	*GenericDevice
	BlockDrive *config.BlockDrive
}

// NewBlockDevice creates a new block device based on DeviceInfo
func NewBlockDevice(devInfo *config.DeviceInfo) *BlockDevice {
	return &BlockDevice{
		GenericDevice: &GenericDevice{
			ID:         devInfo.ID,
			DeviceInfo: devInfo,
		},
	}
}

// Attach is standard interface of api.Device, it's used to add device to some
// DeviceReceiver
func (device *BlockDevice) Attach(devReceiver api.DeviceReceiver) (err error) {
	skip, err := device.bumpAttachCount(true)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	// Increment the block index for the sandbox. This is used to determine the name
	// for the block device in the case where the block device is used as container
	// rootfs and the predicted block device name needs to be provided to the agent.
	index, err := devReceiver.GetAndSetSandboxBlockIndex()

	defer func() {
		if err != nil {
			devReceiver.DecrementSandboxBlockIndex()
		} else {
			device.AttachCount = 1
		}
	}()

	if err != nil {
		return err
	}

	drive := &config.BlockDrive{
		File:   device.DeviceInfo.HostPath,
		Format: "raw",
		ID:     utils.MakeNameID("drive", device.DeviceInfo.ID, maxDevIDSize),
		Index:  index,
	}

	driveName, err := utils.GetVirtDriveName(index)
	if err != nil {
		return err
	}

	customOptions := device.DeviceInfo.DriverOptions
	if customOptions != nil && customOptions["block-driver"] == "virtio-blk" {
		drive.VirtPath = filepath.Join("/dev", driveName)
	} else {
		scsiAddr, err := utils.GetSCSIAddress(index)
		if err != nil {
			return err
		}

		drive.SCSIAddr = scsiAddr
	}

	deviceLogger().WithField("device", device.DeviceInfo.HostPath).Info("Attaching block device")
	device.BlockDrive = drive
	if err = devReceiver.HotplugAddDevice(device, config.DeviceBlock); err != nil {
		return err
	}

	return nil
}

// Detach is standard interface of api.Device, it's used to remove device from some
// DeviceReceiver
func (device *BlockDevice) Detach(devReceiver api.DeviceReceiver) error {
	skip, err := device.bumpAttachCount(false)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	deviceLogger().WithField("device", device.DeviceInfo.HostPath).Info("Unplugging block device")

	if err := devReceiver.HotplugRemoveDevice(device, config.DeviceBlock); err != nil {
		deviceLogger().WithError(err).Error("Failed to unplug block device")
		return err
	}
	device.AttachCount = 0
	return nil
}

// DeviceType is standard interface of api.Device, it returns device type
func (device *BlockDevice) DeviceType() config.DeviceType {
	return config.DeviceBlock
}

// GetDeviceInfo returns device information used for creating
func (device *BlockDevice) GetDeviceInfo() interface{} {
	return device.BlockDrive
}

// It should implement GetAttachCount() and DeviceID() as api.Device implementation
// here it shares function from *GenericDevice so we don't need duplicate codes
