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

package client

import (
	"context"
	"strings"
	"sync"
	"time"

	pb "github.com/CS-SI/SafeScale/lib"
	"github.com/CS-SI/SafeScale/lib/server/utils"
	clitools "github.com/CS-SI/SafeScale/lib/utils/cli"
	googleprotobuf "github.com/golang/protobuf/ptypes/empty"
)

// bucket is the part of the safescale client handling buckets
type bucket struct {
	// session is not used currently.
	session *Session
}

// List ...
func (c *bucket) List(timeout time.Duration) (_ *pb.BucketList, err error) {
	c.session.Connect()
	defer c.session.Disconnect()
	service := pb.NewBucketServiceClient(c.session.connection)
	ctx, err := utils.GetContext(true)
	if err != nil {
		return nil, err
	}

	var ctxTo context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		ctxTo, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	} else {
		ctxTo = ctx
	}

	bl, err := service.List(ctxTo, &googleprotobuf.Empty{})

	return bl, err
}

// Create ...
func (c *bucket) Create(name string, timeout time.Duration) error {
	c.session.Connect()
	defer c.session.Disconnect()

	service := pb.NewBucketServiceClient(c.session.connection)
	ctx, err := utils.GetContext(true)
	if err != nil {
		return err
	}

	var ctxTo context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		ctxTo, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	} else {
		ctxTo = ctx
	}

	_, err = service.Create(ctxTo, &pb.Bucket{Name: name})

	return err
}

// WaitGroup with timeout, returns true when it's a timeout
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c: // OK
		return false
	case <-time.After(timeout): // timeout
		return true
	}
}

// Destroy ...
func (c *bucket) Destroy(names []string, timeout time.Duration) error {
	c.session.Connect()
	defer c.session.Disconnect()
	service := pb.NewBucketServiceClient(c.session.connection)
	ctx, err := utils.GetContext(true)
	if err != nil {
		return err
	}

	var ctxTo context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		ctxTo, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	} else {
		ctxTo = ctx
	}

	var (
		mutex sync.Mutex
		wg    sync.WaitGroup
		errs  []string
	)

	bucketDeleter := func(aname string) {
		defer wg.Done()
		_, err := service.Destroy(ctxTo, &pb.Bucket{Name: aname})
		if err != nil {
			mutex.Lock()
			errs = append(errs, err.Error())
			mutex.Unlock()
		}
	}

	wg.Add(len(names))
	for _, target := range names {
		go bucketDeleter(target)
	}

	isTimeout := waitTimeout(&wg, timeout)

	if len(errs) > 0 {
		return clitools.ExitOnRPC(strings.Join(errs, ", "))
	}

	if isTimeout {
		return clitools.ExitOnRPC("timeout destroying buckets")
	}

	return nil
}

// Delete ...
func (c *bucket) Delete(names []string, timeout time.Duration) error {
	c.session.Connect()
	defer c.session.Disconnect()
	service := pb.NewBucketServiceClient(c.session.connection)
	ctx, err := utils.GetContext(true)
	if err != nil {
		return err
	}

	var ctxTo context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		ctxTo, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	} else {
		ctxTo = ctx
	}

	var (
		mutex sync.Mutex
		wg    sync.WaitGroup
		errs  []string
	)

	bucketDeleter := func(aname string) {
		defer wg.Done()
		_, err := service.Delete(ctxTo, &pb.Bucket{Name: aname})
		if err != nil {
			mutex.Lock()
			defer mutex.Unlock()
			errs = append(errs, err.Error())
			mutex.Unlock()
		}
	}

	wg.Add(len(names))
	for _, target := range names {
		go bucketDeleter(target)
	}

	isTimeout := waitTimeout(&wg, timeout)

	if len(errs) > 0 {
		return clitools.ExitOnRPC(strings.Join(errs, ", "))
	}

	if isTimeout {
		return clitools.ExitOnRPC("timeout deleting buckets")
	}

	return nil
}

// Inspect ...
func (c *bucket) Inspect(name string, timeout time.Duration) (*pb.BucketMountingPoint, error) {
	c.session.Connect()
	defer c.session.Disconnect()
	service := pb.NewBucketServiceClient(c.session.connection)
	ctx, err := utils.GetContext(true)
	if err != nil {
		return nil, err
	}

	var ctxTo context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		ctxTo, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	} else {
		ctxTo = ctx
	}

	rv, err := service.Inspect(ctxTo, &pb.Bucket{Name: name})

	return rv, err
}

// Mount ...
func (c *bucket) Mount(bucketName, hostName, mountPoint string, timeout time.Duration) error {
	c.session.Connect()
	defer c.session.Disconnect()
	service := pb.NewBucketServiceClient(c.session.connection)
	ctx, err := utils.GetContext(true)
	if err != nil {
		return err
	}

	var ctxTo context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		ctxTo, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	} else {
		ctxTo = ctx
	}

	_, err = service.Mount(ctxTo, &pb.BucketMountingPoint{
		Bucket: bucketName,
		Host: &pb.Reference{
			Name: hostName,
		},
		Path: mountPoint,
	})

	return err
}

// Unmount ...
func (c *bucket) Unmount(bucketName, hostName string, timeout time.Duration) error {
	c.session.Connect()
	defer c.session.Disconnect()
	service := pb.NewBucketServiceClient(c.session.connection)
	ctx, err := utils.GetContext(true)
	if err != nil {
		return err
	}

	var ctxTo context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		ctxTo, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	} else {
		ctxTo = ctx
	}

	_, err = service.Unmount(ctxTo, &pb.BucketMountingPoint{
		Bucket: bucketName,
		Host:   &pb.Reference{Name: hostName},
	})

	return err
}
