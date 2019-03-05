/*
 * Copyright 2018-2019, CS Systemes d'Information, http://www.c-s.fr
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

package listeners

import (
	"context"
	"fmt"

	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	log "github.com/sirupsen/logrus"

	pb "github.com/CS-SI/SafeScale/safescale"
	"github.com/CS-SI/SafeScale/safescale/server/handlers"
	"github.com/CS-SI/SafeScale/safescale/utils"
	conv "github.com/CS-SI/SafeScale/safescale/utils"
)

// HostHandler ...
var HostHandler = handlers.NewHostHandler

// safescale host create host1 --net="net1" --cpu=2 --ram=7 --disk=100 --os="Ubuntu 16.04" --public=true
// safescale host list --all=false
// safescale host inspect host1
// safescale host create host2 --net="net1" --cpu=2 --ram=7 --disk=100 --os="Ubuntu 16.04" --public=false

// HostListener host service server grpc
type HostListener struct{}

// StoredCPUInfo ...
type StoredCPUInfo struct {
	ID           string `bow:"key"`
	TenantName   string `json:"tenant_name,omitempty"`
	TemplateID   string `json:"template_id,omitempty"`
	TemplateName string `json:"template_name,omitempty"`
	ImageID      string `json:"image_id,omitempty"`
	ImageName    string `json:"image_name,omitempty"`
	LastUpdated  string `json:"last_updated,omitempty"`

	NumberOfCPU    int     `json:"number_of_cpu,omitempty"`
	NumberOfCore   int     `json:"number_of_core,omitempty"`
	NumberOfSocket int     `json:"number_of_socket,omitempty"`
	CPUFrequency   float64 `json:"cpu_frequency,omitempty"`
	CPUArch        string  `json:"cpu_arch,omitempty"`
	Hypervisor     string  `json:"hypervisor,omitempty"`
	CPUModel       string  `json:"cpu_model,omitempty"`
	RAMSize        float64 `json:"ram_size,omitempty"`
	RAMFreq        float64 `json:"ram_freq,omitempty"`
	GPU            int     `json:"gpu,omitempty"`
	GPUModel       string  `json:"gpu_model,omitempty"`
}

// Start ...
func (s *HostListener) Start(ctx context.Context, in *pb.Reference) (*google_protobuf.Empty, error) {
	log.Infof("Listeners: host start '%s' called", in.Name)
	defer log.Debugf("Listeners: host start '%s' done", in.Name)

	ctx, cancelFunc := context.WithCancel(ctx)
	if err := utils.ProcessRegister(ctx, cancelFunc, "Start Host "+in.GetName()); err != nil {
		return nil, fmt.Errorf("Failed to register the process : %s", err.Error())
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		log.Info("Can't start host: no tenant set")
		return nil, grpc.Errorf(codes.FailedPrecondition, "Can't start host: no tenant set")
	}

	handler := HostHandler(tenant.Service)
	ref := utils.GetReference(in)
	err := handler.Start(ctx, ref)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	log.Printf("Host '%s' successfully started", ref)
	return &google_protobuf.Empty{}, nil
}

// Stop ...
func (s *HostListener) Stop(ctx context.Context, in *pb.Reference) (*google_protobuf.Empty, error) {
	log.Infof("Listeners: host stop '%s' called", in.Name)
	defer log.Debugf("Listeners: host stop '%s' done", in.Name)

	ctx, cancelFunc := context.WithCancel(ctx)

	if err := utils.ProcessRegister(ctx, cancelFunc, "Stop Host "+in.GetName()); err == nil {
		defer utils.ProcessDeregister(ctx)
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		log.Info("Can't stop host: no tenant set")
		return nil, grpc.Errorf(codes.FailedPrecondition, "can't stop host: no tenant set")
	}

	handler := HostHandler(tenant.Service)
	ref := utils.GetReference(in)
	err := handler.Stop(ctx, ref)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	log.Printf("Host '%s' stopped", ref)
	return &google_protobuf.Empty{}, nil
}

// Reboot ...
func (s *HostListener) Reboot(ctx context.Context, in *pb.Reference) (*google_protobuf.Empty, error) {
	log.Infof("Listeners: host reboot '%s' called", in.Name)
	defer log.Debugf("Listeners: host reboot '%s' done", in.Name)

	ctx, cancelFunc := context.WithCancel(ctx)

	if err := utils.ProcessRegister(ctx, cancelFunc, "Reboot Host "+in.GetName()); err == nil {
		defer utils.ProcessDeregister(ctx)
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		log.Info("Can't reboot host: no tenant set")
		return nil, grpc.Errorf(codes.FailedPrecondition, "can't reboot host: no tenant set")
	}

	handler := HostHandler(tenant.Service)
	ref := utils.GetReference(in)
	err := handler.Reboot(ctx, ref)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	log.Printf("Host '%s' successfully rebooted.", ref)
	return &google_protobuf.Empty{}, nil
}

// List available hosts
func (s *HostListener) List(ctx context.Context, in *pb.HostListRequest) (*pb.HostList, error) {
	log.Infoln("Listeners: host list called")
	defer log.Debugln("Listeners: host list done")

	ctx, cancelFunc := context.WithCancel(ctx)

	if err := utils.ProcessRegister(ctx, cancelFunc, "List Hosts"); err == nil {
		defer utils.ProcessDeregister(ctx)
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		log.Info("Can't list host: no tenant set")
		return nil, grpc.Errorf(codes.FailedPrecondition, "can't list hosts: no tenant set")
	}

	handler := HostHandler(tenant.Service)
	hosts, err := handler.List(ctx, in.GetAll())
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	// Map resources.Host to pb.Host
	var pbhost []*pb.Host
	for _, host := range hosts {
		pbhost = append(pbhost, conv.ToPBHost(host))
	}
	rv := &pb.HostList{Hosts: pbhost}
	return rv, nil
}

// Create a new host
func (s *HostListener) Create(ctx context.Context, in *pb.HostDefinition) (*pb.Host, error) {
	log.Infof("Listeners: host create '%s' done", in.Name)
	defer log.Debugf("Listeners: host create '%s' done", in.Name)

	ctx, cancelFunc := context.WithCancel(ctx)

	if err := utils.ProcessRegister(ctx, cancelFunc, "Create Host "+in.GetName()); err == nil {
		defer utils.ProcessDeregister(ctx)
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		log.Info("Can't create host: no tenant set")
		return nil, grpc.Errorf(codes.FailedPrecondition, "can't create host: no tenant set")
	}

	handler := HostHandler(tenant.Service)
	host, err := handler.Create(ctx,
		in.GetName(),
		in.GetNetwork(),
		int(in.GetCpuCount()),
		in.GetRam(),
		int(in.GetDisk()),
		in.GetImageId(),
		in.GetPublic(),
		int(in.GetGpuCount()),
		float32(in.GetCpuFreq()),
		in.Force,
	)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	log.Infof("Host '%s' created", in.GetName())
	return conv.ToPBHost(host), nil

}

// Resize an host
func (s *HostListener) Resize(ctx context.Context, in *pb.HostDefinition) (*pb.Host, error) {
	log.Infof("Listeners: host resize '%s' done", in.Name)
	defer log.Debugf("Listeners: host resize '%s' done", in.Name)

	ctx, cancelFunc := context.WithCancel(ctx)

	if err := utils.ProcessRegister(ctx, cancelFunc, "Resize Host "+in.GetName()); err == nil {
		defer utils.ProcessDeregister(ctx)
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		log.Info("Can't resize host: no tenant set")
		return nil, grpc.Errorf(codes.FailedPrecondition, "can't resize host: no tenant set")
	}

	handler := HostHandler(tenant.Service)
	host, err := handler.Resize(ctx,
		in.GetName(),
		int(in.GetCpuCount()),
		in.GetRam(),
		int(in.GetDisk()),
		int(in.GetGpuCount()),
		float32(in.GetCpuFreq()),
	)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	log.Infof("Host '%s' resized", in.GetName())
	return conv.ToPBHost(host), nil
}

// Status of a host
func (s *HostListener) Status(ctx context.Context, in *pb.Reference) (*pb.HostStatus, error) {
	log.Infof("Listeners: host status '%s' called", in.Name)
	defer log.Debugf("Listeners: host status '%s' done", in.Name)

	ctx, cancelFunc := context.WithCancel(ctx)

	if err := utils.ProcessRegister(ctx, cancelFunc, "Status of Host "+in.GetName()); err == nil {
		defer utils.ProcessDeregister(ctx)
	}

	ref := utils.GetReference(in)
	if ref == "" {
		return nil, fmt.Errorf("Can't get host status: neither name nor id given as reference")
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		log.Info("Can't get host status: no tenant set")
		return nil, grpc.Errorf(codes.FailedPrecondition, "can't get host status: no tenant set")
	}

	handler := HostHandler(tenant.Service)
	host, err := handler.ForceInspect(ctx, ref)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	return conv.ToHostStatus(host), nil
}

// Inspect an host
func (s *HostListener) Inspect(ctx context.Context, in *pb.Reference) (*pb.Host, error) {
	log.Infof("Receiving 'host inspect %s'", in.Name)
	log.Debugf(">>> safescale.server.listeners.HostListener::Inspect(%s)", in.Name)
	defer log.Debugf("<<< safescale.server.listeners.HostListener::Inspect(%s)", in.Name)

	ctx, cancelFunc := context.WithCancel(ctx)

	if err := utils.ProcessRegister(ctx, cancelFunc, "Inspect Host "+in.GetName()); err == nil {
		defer utils.ProcessDeregister(ctx)
	}

	ref := utils.GetReference(in)
	if ref == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "can't inspect host: neither name nor id given as reference")
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		log.Info("Can't inspect host: no tenant set")
		return nil, grpc.Errorf(codes.FailedPrecondition, "can't inspect host: no tenant set")
	}

	handler := HostHandler(tenant.Service)
	host, err := handler.ForceInspect(ctx, ref)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, fmt.Sprintf("can't inspect host: %v", err))
	}
	return conv.ToPBHost(host), nil
}

// Delete an host
func (s *HostListener) Delete(ctx context.Context, in *pb.Reference) (*google_protobuf.Empty, error) {
	log.Infof("Receiving 'host delete %s'", in.Name)
	log.Debugf(">>> safescale.server.listeners.HostListener::Delete(%s)", in.Name)
	defer log.Debugf("<<< safescale.server.Listeners.HostListener::Delete(%s)", in.Name)

	ctx, cancelFunc := context.WithCancel(ctx)

	if err := utils.ProcessRegister(ctx, cancelFunc, "Delete Host "+in.GetName()); err == nil {
		defer utils.ProcessDeregister(ctx)
	}

	ref := utils.GetReference(in)
	if ref == "" {
		return nil, fmt.Errorf("Can't delete host: neither name nor id given as reference")
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		log.Info("Can't delete host: no tenant set")
		return nil, grpc.Errorf(codes.FailedPrecondition, "can't delete host: no tenant set")
	}

	handler := HostHandler(tenant.Service)
	err := handler.Delete(ctx, ref)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	log.Printf("Host '%s' successfully deleted.", ref)
	return &google_protobuf.Empty{}, nil
}

// SSH returns ssh parameters to access an host
func (s *HostListener) SSH(ctx context.Context, in *pb.Reference) (*pb.SshConfig, error) {
	log.Debugf(">>> safescale.server.listeners.HostListener::SSH(%s)", in.Name)
	defer log.Debugf("<<< safescale.server.listeners.HostListener::SSH(%s)", in.Name)

	ctx, cancelFunc := context.WithCancel(ctx)

	if err := utils.ProcessRegister(ctx, cancelFunc, "SSH config of Host "+in.GetName()); err == nil {
		defer utils.ProcessDeregister(ctx)
	}

	ref := utils.GetReference(in)
	if ref == "" {
		return nil, fmt.Errorf("Can't ssh to host: neither name nor id given as reference")
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		log.Info("Can't delete host: no tenant set")
		return nil, grpc.Errorf(codes.FailedPrecondition, "can't ssh host: no tenant set")
	}

	handler := HostHandler(currentTenant.Service)
	sshConfig, err := handler.SSH(ctx, ref)
	if err != nil {
		return nil, err
	}
	return conv.ToPBSshConfig(sshConfig), nil
}
