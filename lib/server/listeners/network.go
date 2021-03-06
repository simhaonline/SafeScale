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

package listeners

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	googleprotobuf "github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/CS-SI/SafeScale/lib"
	"github.com/CS-SI/SafeScale/lib/server/handlers"
	"github.com/CS-SI/SafeScale/lib/server/iaas/resources"
	"github.com/CS-SI/SafeScale/lib/server/iaas/resources/enums/ipversion"
	srvutils "github.com/CS-SI/SafeScale/lib/server/utils"
	"github.com/CS-SI/SafeScale/lib/utils/concurrency"
	"github.com/CS-SI/SafeScale/lib/utils/scerr"
)

// NetworkHandler ...
var NetworkHandler = handlers.NewNetworkHandler

// safescale network create net1 --cidr="192.145.0.0/16" --cpu=2 --ram=7 --disk=100 --os="Ubuntu 16.04" (par défault "192.168.0.0/24", on crée une gateway sur chaque réseau: gw-net1)
// safescale network list
// safescale network delete net1
// safescale network inspect net1

// NetworkListener network service server grpc
type NetworkListener struct{}

// Create a new network
func (s *NetworkListener) Create(ctx context.Context, in *pb.NetworkDefinition) (net *pb.Network, err error) {
	if s == nil {
		return nil, status.Errorf(codes.FailedPrecondition, scerr.InvalidInstanceError().Message())
	}
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, scerr.InvalidParameterError("in", "cannot be nil").Message())
	}
	networkName := in.GetName()

	tracer := concurrency.NewTracer(nil, fmt.Sprintf("('%s')", networkName), true).WithStopwatch().GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()

	ctx, cancelFunc := context.WithCancel(ctx)
	if err := srvutils.JobRegister(ctx, cancelFunc, "Create network "+networkName); err == nil {
		defer srvutils.JobDeregister(ctx)
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		// log.Info("Can't create network: no tenant set")
		return nil, status.Errorf(codes.FailedPrecondition, "cannot create network: no tenant set")
	}

	var (
		sizing    *resources.SizingRequirements
		gwImageID string
		gwName    string
	)
	if in.Gateway == nil || in.Gateway.Sizing == nil {
		sizing = &resources.SizingRequirements{
			MinCores:    int(in.Gateway.Sizing.MinCpuCount),
			MaxCores:    int(in.Gateway.Sizing.MaxCpuCount),
			MinRAMSize:  in.Gateway.Sizing.MinRamSize,
			MaxRAMSize:  in.Gateway.Sizing.MaxRamSize,
			MinDiskSize: int(in.Gateway.Sizing.MinDiskSize),
			MinGPU:      int(in.Gateway.Sizing.GpuCount),
			MinFreq:     in.Gateway.Sizing.MinCpuFreq,
		}
	} else {
		s := srvutils.FromPBHostSizing(in.Gateway.Sizing)
		sizing = &s
	}
	if in.Gateway != nil {
		gwImageID = in.GetGateway().GetImageId()
		gwName = in.GetGateway().GetName()
	}

	handler := NetworkHandler(tenant.Service)
	network, err := handler.Create(ctx,
		networkName,
		in.GetCidr(),
		ipversion.IPv4,
		*sizing,
		gwImageID,
		gwName,
		in.FailOver,
		in.Domain,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, getUserMessage(err))
	}

	log.Infof("Network '%s' successfully created.", networkName)
	return srvutils.ToPBNetwork(network), nil
}

// List existing networks
func (s *NetworkListener) List(ctx context.Context, in *pb.NetworkListRequest) (rv *pb.NetworkList, err error) {
	if s == nil {
		return nil, status.Errorf(codes.FailedPrecondition, scerr.InvalidInstanceError().Message())
	}
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, scerr.InvalidParameterError("in", "cannot be nil").Message())
	}

	tracer := concurrency.NewTracer(nil, "", true).WithStopwatch().GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()

	log.Infof("Listeners: network list")

	ctx, cancelFunc := context.WithCancel(ctx)
	if err := srvutils.JobRegister(ctx, cancelFunc, "List networks"); err == nil {
		defer srvutils.JobDeregister(ctx)
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		// log.Info("Can't list network: no tenant set")
		return nil, status.Errorf(codes.FailedPrecondition, "cannot list networks: no tenant set")
	}

	handler := NetworkHandler(tenant.Service)
	networks, err := handler.List(ctx, in.GetAll())
	if err != nil {
		return nil, status.Errorf(codes.Internal, getUserMessage(err))
	}

	// Map resources.Network to pb.Network
	var pbnetworks []*pb.Network
	for _, network := range networks {
		pbnetworks = append(pbnetworks, srvutils.ToPBNetwork(network))
	}
	rv = &pb.NetworkList{Networks: pbnetworks}
	return rv, nil
}

// Inspect returns infos on a network
func (s *NetworkListener) Inspect(ctx context.Context, in *pb.Reference) (net *pb.Network, err error) {
	if s == nil {
		return nil, status.Errorf(codes.FailedPrecondition, scerr.InvalidInstanceError().Message())
	}
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, scerr.InvalidParameterError("in", "cannot be nil").Message())
	}
	ref := srvutils.GetReference(in)
	if ref == "" {
		return nil, status.Errorf(codes.FailedPrecondition, "cannot inspect network: neither name nor id given as reference")
	}

	tracer := concurrency.NewTracer(nil, fmt.Sprintf("('%s')", ref), true).WithStopwatch().GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()

	ctx, cancelFunc := context.WithCancel(ctx)
	if err := srvutils.JobRegister(ctx, cancelFunc, "Inspect network "+in.GetName()); err == nil {
		defer srvutils.JobDeregister(ctx)
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		log.Info("Can't inspect network: no tenant set")
		return nil, status.Errorf(codes.FailedPrecondition, "cannot inspect network: no tenant set")
	}

	handler := NetworkHandler(currentTenant.Service)
	network, err := handler.Inspect(ctx, ref)
	if err != nil {
		return nil, status.Errorf(codes.Internal, getUserMessage(err))
	}
	if network == nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("cannot inspect network '%s': not found", ref))
	}

	return srvutils.ToPBNetwork(network), nil
}

// Delete a network
func (s *NetworkListener) Delete(ctx context.Context, in *pb.Reference) (buf *googleprotobuf.Empty, err error) {
	if s == nil {
		return nil, status.Errorf(codes.FailedPrecondition, scerr.InvalidInstanceError().Message())
	}
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, scerr.InvalidParameterError("in", "cannot be nil").Message())
	}
	ref := srvutils.GetReference(in)
	if ref == "" {
		return nil, status.Errorf(codes.FailedPrecondition, "cannot inspect network: neither name nor id given as reference")
	}

	tracer := concurrency.NewTracer(nil, fmt.Sprintf("('%s')", ref), true).WithStopwatch().GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()

	ctx, cancelFunc := context.WithCancel(ctx)
	if err := srvutils.JobRegister(ctx, cancelFunc, "Delete network "+in.GetName()); err == nil {
		defer srvutils.JobDeregister(ctx)
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		// log.Info("Can't delete network: no tenant set")
		return nil, status.Errorf(codes.FailedPrecondition, "cannot delete network: no tenant set")
	}

	handler := NetworkHandler(currentTenant.Service)
	err = handler.Delete(ctx, ref)
	if err != nil {
		return nil, status.Errorf(codes.Internal, getUserMessage(err))
	}

	log.Infof("Network '%s' successfully deleted.", ref)
	return &googleprotobuf.Empty{}, nil
}

// Destroy a network
func (s *NetworkListener) Destroy(ctx context.Context, in *pb.Reference) (buf *googleprotobuf.Empty, err error) {
	if s == nil {
		return nil, status.Errorf(codes.FailedPrecondition, scerr.InvalidInstanceError().Message())
	}
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, scerr.InvalidParameterError("in", "cannot be nil").Message())
	}
	ref := srvutils.GetReference(in)
	if ref == "" {
		return nil, status.Errorf(codes.FailedPrecondition, "cannot inspect network: neither name nor id given as reference")
	}

	tracer := concurrency.NewTracer(nil, fmt.Sprintf("('%s')", ref), true).WithStopwatch().GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()

	ctx, cancelFunc := context.WithCancel(ctx)
	if err := srvutils.JobRegister(ctx, cancelFunc, "Delete network "+in.GetName()); err == nil {
		defer srvutils.JobDeregister(ctx)
	}

	tenant := GetCurrentTenant()
	if tenant == nil {
		// log.Info("Can't delete network: no tenant set")
		return nil, status.Errorf(codes.FailedPrecondition, "cannot delete network: no tenant set")
	}

	handler := NetworkHandler(currentTenant.Service)
	err = handler.Destroy(ctx, ref)
	if err != nil {
		return nil, status.Errorf(codes.Internal, getUserMessage(err))
	}

	log.Infof("Network '%s' successfully deleted.", ref)
	return &googleprotobuf.Empty{}, nil
}
