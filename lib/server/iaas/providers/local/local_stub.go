// +build !libvirt

/*
 * Copyright 2018-2020, CS Systemes d'Information, http://www.c-s.fr
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package local

import (
	"fmt"

	"github.com/CS-SI/SafeScale/lib/server/iaas"
	"github.com/CS-SI/SafeScale/lib/server/iaas/providers"
	providerapi "github.com/CS-SI/SafeScale/lib/server/iaas/providers/api"
	"github.com/CS-SI/SafeScale/lib/server/iaas/resources"
	"github.com/CS-SI/SafeScale/lib/server/iaas/resources/enums/hoststate"
	"github.com/CS-SI/SafeScale/lib/server/iaas/resources/userdata"
)

var errorStr = "libvirt Driver is not enabled, use the libvirt option while compiling (make libvirt all)"

// provider is the implementation of the local driver regarding to the api.Provider
type provider struct {
	tenantParameters map[string]interface{}
}

// AuthOptions fields are the union of those recognized by each identity implementation and provider.
type AuthOptions struct {
}

// CfgOptions configuration options
type CfgOptions struct {
}

func (provider *provider) Build(params map[string]interface{}) (providerapi.Provider, error) {
	provider.tenantParameters = map[string]interface{}{}
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetAuthenticationOptions() (providers.Config, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetConfigurationOptions() (providers.Config, error) {
	return nil, fmt.Errorf(errorStr)
}

func (provider *provider) ListAvailabilityZones() (map[string]bool, error) {
	return nil, fmt.Errorf(errorStr)
}

// ListRegions returns a list with the regions available
func (provider *provider) ListRegions() ([]string, error) {
	return nil, fmt.Errorf(errorStr)
}

func (provider *provider) ListImages(all bool) ([]resources.Image, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetImage(id string) (*resources.Image, error) {
	return nil, fmt.Errorf(errorStr)
}

func (provider *provider) GetTemplate(id string) (*resources.HostTemplate, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) ListTemplates(all bool) ([]resources.HostTemplate, error) {
	return nil, fmt.Errorf(errorStr)
}

func (provider *provider) CreateKeyPair(name string) (*resources.KeyPair, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetKeyPair(id string) (*resources.KeyPair, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) ListKeyPairs() ([]resources.KeyPair, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) DeleteKeyPair(id string) error {
	return fmt.Errorf(errorStr)
}

func (provider *provider) CreateNetwork(req resources.NetworkRequest) (*resources.Network, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetNetwork(id string) (*resources.Network, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetNetworkByName(name string) (*resources.Network, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) ListNetworks() ([]*resources.Network, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) DeleteNetwork(id string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) CreateGateway(req resources.GatewayRequest, sizing *resources.SizingRequirements) (*resources.Host, *userdata.Content, error) {
	return nil, nil, fmt.Errorf(errorStr)
}
func (provider *provider) DeleteGateway(string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) CreateVIP(networkID string, description string) (*resources.VirtualIP, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) AddPublicIPToVIP(vip *resources.VirtualIP) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) BindHostToVIP(vip *resources.VirtualIP, hostID string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) UnbindHostFromVIP(vip *resources.VirtualIP, hostID string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) DeleteVIP(vip *resources.VirtualIP) error {
	return fmt.Errorf(errorStr)
}

func (provider *provider) CreateHost(request resources.HostRequest) (*resources.Host, *userdata.Content, error) {
	return nil, nil, fmt.Errorf(errorStr)
}
func (provider *provider) ResizeHost(id string, request resources.SizingRequirements) (*resources.Host, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) InspectHost(interface{}) (*resources.Host, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetHostByName(string) (*resources.Host, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetHostState(interface{}) (hoststate.Enum, error) {
	return hoststate.ERROR, fmt.Errorf(errorStr)
}
func (provider *provider) ListHosts() ([]*resources.Host, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) DeleteHost(id string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) StartHost(id string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) StopHost(id string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) RebootHost(id string) error {
	return fmt.Errorf(errorStr)
}

func (provider *provider) CreateVolume(request resources.VolumeRequest) (*resources.Volume, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetVolume(id string) (*resources.Volume, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) ListVolumes() ([]resources.Volume, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) DeleteVolume(id string) error {
	return fmt.Errorf(errorStr)
}

func (provider *provider) CreateVolumeAttachment(request resources.VolumeAttachmentRequest) (string, error) {
	return "", fmt.Errorf(errorStr)
}
func (provider *provider) GetVolumeAttachment(serverID, id string) (*resources.VolumeAttachment, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) ListVolumeAttachments(serverID string) ([]resources.VolumeAttachment, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) DeleteVolumeAttachment(serverID, id string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) GetName() string {
	return "local_disabled"
}
func (provider *provider) GetTenantParameters() map[string]interface{} {
	return nil
}

// GetCapabilities returns the capabilities of the provider
func (provider *provider) GetCapabilities() providers.Capabilities {
	return providers.Capabilities{}
}

func init() {
	// log.Debug("Registering fake local provider")
	iaas.Register("local", &provider{})
}
