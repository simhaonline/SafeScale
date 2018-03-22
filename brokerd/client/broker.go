package main

import (
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/SafeScale/brokerd"
	cli "github.com/urfave/cli"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// CLI Flags
	app := cli.NewApp()
	app.Name = "broker"
	app.Usage = "broker COMMAND"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "CS-SI",
			Email: "safescale@c-s.fr",
		},
	}
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		{
			Name:  "network",
			Usage: "network COMMAND",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "list",
					Action: func(c *cli.Context) error {
						networkService := pb.NewNetworkServiceClient(conn)
						networks, err := networkService.List(ctx, &pb.Empty{})
						if err != nil {
							return fmt.Errorf("could not get network list: %v", err)
						}
						for i, network := range networks.GetNetworks() {
							// log.Printf("Network %d: %s", i, network)
							fmt.Printf("Network %d: %s", i, network)
						}

						return nil
					},
				},
				{
					Name:      "delete",
					Usage:     "delete NETWORK",
					ArgsUsage: "<network_name>",
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							fmt.Println("Missing mandatory argument <network_name>")
							cli.ShowSubcommandHelp(c)
							return fmt.Errorf("Network name required")
						}

						// Network
						networkService := pb.NewNetworkServiceClient(conn)
						_, err := networkService.Delete(ctx, &pb.Reference{Name: c.Args().First(), TenantID: "TestOvh"})
						if err != nil {
							return fmt.Errorf("could not delete network %s: %v", c.Args().First(), err)
						}
						fmt.Printf("Network %s deleted", c.Args().First())

						return nil
					},
				},
				{
					Name:      "inspect",
					Usage:     "inspect NETWORK",
					ArgsUsage: "<network_name>",
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							fmt.Println("Missing mandatory argument <network_name>")
							cli.ShowSubcommandHelp(c)
							return fmt.Errorf("Network name required")
						}

						// Network
						networkService := pb.NewNetworkServiceClient(conn)
						network, err := networkService.Inspect(ctx, &pb.Reference{Name: c.Args().First(), TenantID: "TestOvh"})
						if err != nil {
							return fmt.Errorf("could not inspect network %s: %v", c.Args().First(), err)
						}
						fmt.Printf("Network infos: %s", network)

						return nil
					},
				},
				{
					Name:      "create",
					Usage:     "create a network",
					ArgsUsage: "<network_name>",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "cidr",
							Value: "192.168.0.0/24",
							Usage: "cidr of the network",
						},
						cli.IntFlag{
							Name:  "cpu",
							Value: 1,
							Usage: "Number of CPU for the gateway",
						},
						cli.Float64Flag{
							Name:  "ram",
							Value: 1,
							Usage: "RAM for the gateway",
						},
						cli.IntFlag{
							Name:  "disk",
							Value: 100,
							Usage: "Disk space for the gateway",
						},
						cli.StringFlag{
							Name:  "os",
							Value: "Ubuntu 16.04",
							Usage: "Image name for the gateway",
						}},
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							fmt.Println("Missing mandatory argument <network_name>")
							cli.ShowSubcommandHelp(c)
							return fmt.Errorf("Network name reqired")
						}
						fmt.Println("create network: ", c.Args().First())
						// Network
						networkService := pb.NewNetworkServiceClient(conn)
						netdef := &pb.NetworkDefinition{
							CIDR:   c.String("cidr"),
							Name:   c.Args().Get(0),
							Tenant: "TestOvh",
							Gateway: &pb.GatewayDefinition{
								CPU:  int32(c.Int("cpu")),
								Disk: int32(c.Int("disk")),
								RAM:  float32(c.Float64("ram")),
								// CPUFrequency: ??,
								ImageID: c.String("os"),
							},
						}
						network, err := networkService.Create(ctx, netdef)
						if err != nil {
							return fmt.Errorf("Could not get network list: %v", err)
						}
						fmt.Printf("Network: %s", network)

						return nil
					},
				},
			},
		},
		{
			Name:  "tenant",
			Usage: "tenant COMMAND",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List available tenants",
					Action: func(c *cli.Context) error {
						tenantService := pb.NewTenantServiceClient(conn)
						tenants, err := tenantService.List(ctx, &pb.Empty{})
						if err != nil {
							return fmt.Errorf("Could not get tenant list: %v", err)
						}
						for i, tenant := range tenants.GetTenants() {
							fmt.Printf("Tenant %d: %s", i, tenant)
						}

						return nil
					},
				},
				{
					Name:  "get",
					Usage: "Get current tenant",
					Action: func(c *cli.Context) error {
						tenantService := pb.NewTenantServiceClient(conn)
						tenant, err := tenantService.Get(ctx, &pb.Empty{})
						if err != nil {
							return fmt.Errorf("Could not get current tenant: %v", err)
						}
						fmt.Println(tenant.GetName())

						return nil
					},
				},
				{
					Name:      "set",
					Usage:     "Set tenant to work with",
					ArgsUsage: "<tenant_name>",
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							fmt.Println("Missing mandatory argument <tenant_name>")
							cli.ShowSubcommandHelp(c)
							return fmt.Errorf("Tenant name required")
						}

						tenantService := pb.NewTenantServiceClient(conn)
						_, err := tenantService.Set(ctx, &pb.TenantName{Name: c.Args().First()})
						if err != nil {
							return fmt.Errorf("Could not get current tenant: %v", err)
						}
						fmt.Printf("Tenant '%s' set", c.Args().First())

						return nil
					},
				},
			}}, {
			Name:  "vm",
			Usage: "vm COMMAND",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List available VMs",
					Action: func(c *cli.Context) error {
						service := pb.NewVMServiceClient(conn)
						resp, err := service.List(ctx, &pb.Empty{})
						if err != nil {
							return fmt.Errorf("Could not get vm list: %v", err)
						}
						for i, vm := range resp.GetVMs() {
							fmt.Println(fmt.Sprintf("VM %d: %s", i, vm))
						}

						return nil
					},
				},
				{
					Name:      "inspect",
					Usage:     "inspect VM",
					ArgsUsage: "<VM_name|VM_ID>",
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							fmt.Println("Missing mandatory argument <VM_name>")
							cli.ShowSubcommandHelp(c)
							return fmt.Errorf("VM name or ID required")
						}
						service := pb.NewVMServiceClient(conn)
						resp, err := service.Inspect(ctx, &pb.Reference{Name: c.Args().First()})
						if err != nil {
							return fmt.Errorf("Could not inspect vm '%s': %v", c.Args().First(), err)
						}

						fmt.Printf("VM infos: %s", resp)

						return nil
					},
				}, {
					Name:      "create",
					Usage:     "create a new VM",
					ArgsUsage: "<VM_name>",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "net",
							Usage: "Name or ID of the network to put the VM on",
						},
						cli.IntFlag{
							Name:  "cpu",
							Value: 1,
							Usage: "Number of CPU for the VM",
						},
						cli.Float64Flag{
							Name:  "ram",
							Value: 1,
							Usage: "RAM for the VM",
						},
						cli.IntFlag{
							Name:  "disk",
							Value: 100,
							Usage: "Disk space for the VM",
						},
						cli.StringFlag{
							Name:  "os",
							Value: "Ubuntu 16.04",
							Usage: "Image name for the VM",
						},
						cli.BoolTFlag{
							Name:  "public",
							Usage: "Public IP",
						},
						cli.BoolFlag{
							Name:   "gpu",
							Usage:  "With GPU",
							Hidden: true,
						},
					},
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							fmt.Println("Missing mandatory argument <VM_name>")
							cli.ShowSubcommandHelp(c)
							return fmt.Errorf("VM name required")
						}

						service := pb.NewVMServiceClient(conn)
						resp, err := service.Create(ctx, &pb.VMDefinition{
							Name:      c.Args().First(),
							CPUNumber: int32(c.Int("cpu")),
							Disk:      float32(c.Float64("disk")),
							GPU:       c.Bool("gpu"),
							ImageID:   c.String("os"),
							Network:   c.String("net"),
							Public:    c.BoolT("public"),
							RAM:       float32(c.Float64("ram")),
						})
						if err != nil {
							return fmt.Errorf("Could not create vm '%s': %v", c.Args().First(), err)
						}

						fmt.Printf("VM infos: %s", resp)

						return nil
					},
				}, {
					Name:      "delete",
					Usage:     "Delete VM",
					ArgsUsage: "<VM_name|VM_ID>",
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							fmt.Println("Missing mandatory argument <VM_name>")
							cli.ShowSubcommandHelp(c)
							return fmt.Errorf("VM name or ID required")
						}
						service := pb.NewVMServiceClient(conn)
						_, err := service.Delete(ctx, &pb.Reference{Name: c.Args().First()})
						if err != nil {
							return fmt.Errorf("Could not delete vm '%s': %v", c.Args().First(), err)
						}
						fmt.Printf("VM '%s' deleted", c.Args().First())
						return nil
					},
				}, {
					Name:      "ssh",
					Usage:     "Get ssh config to connect to VM",
					ArgsUsage: "<VM_name|VM_ID>",
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							fmt.Println("Missing mandatory argument <VM_name>")
							cli.ShowSubcommandHelp(c)
							return fmt.Errorf("VM name or ID required")
						}
						service := pb.NewVMServiceClient(conn)
						resp, err := service.Ssh(ctx, &pb.Reference{Name: c.Args().First()})
						if err != nil {
							return fmt.Errorf("Could not get ssh config for vm '%s': %v", c.Args().First(), err)
						}
						fmt.Printf("Ssh config for VM '%s': %s", c.Args().First(), resp)
						return nil
					},
				}}},
	}
	err = app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
