// Copyright (c) 2018 HyperHQ Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// direct implements base vm factory without vm templating.

package direct

import (
	"context"

	vc "github.com/kata-containers/runtime/virtcontainers"
	"github.com/kata-containers/runtime/virtcontainers/factory/base"
)

type direct struct {
	config vc.VMConfig
}

// New returns a new direct vm factory.
func New(ctx context.Context, config vc.VMConfig) base.FactoryBase {
	return &direct{config}
}

// Config returns the direct factory's configuration.
func (d *direct) Config() vc.VMConfig {
	return d.config
}

// GetBaseVM create a new VM directly.
func (d *direct) GetBaseVM(ctx context.Context) (*vc.VM, error) {
	vm, err := vc.NewVM(ctx, d.config)
	if err != nil {
		return nil, err
	}

	err = vm.Pause()
	if err != nil {
		vm.Stop()
		return nil, err
	}

	return vm, nil
}

// CloseFactory closes the direct vm factory.
func (d *direct) CloseFactory(ctx context.Context) {
}
