package integrationtests

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/CS-SI/SafeScale/integrationtests/enums/providers"
	"github.com/CS-SI/SafeScale/lib/utils/temporal"
)

func keyFromProvider(provider providers.Enum) string {
	switch provider {
	case providers.LOCAL:
		return "TEST_LOCAL"
	case providers.OVH:
		return "TEST_OVH"
	case providers.CLOUDFERRO:
		return "TEST_CLOUDFERRO"
	case providers.FLEXIBLEENGINE:
		return "TEST_FLEXIBLE"
	case providers.AWS:
		return "TEST_AWS"
	case providers.GCP:
		return "TEST_GCP"
	case providers.OUTSCALE:
		return "TEST_OUTSCALE"
	}
	return ""
}

func nameFromProvider(provider providers.Enum) string {
	switch provider {
	case providers.LOCAL:
		return "local"
	case providers.OVH:
		return "ovh"
	case providers.CLOUDFERRO:
		return "cloudferro"
	case providers.FLEXIBLEENGINE:
		return "flexibleengine"
	case providers.AWS:
		return "aws"
	case providers.GCP:
		return "gcp"
	case providers.OUTSCALE:
		return "outscale"
	}
	return ""
}

func EnvSetup(t *testing.T, provider providers.Enum) {
	key := keyFromProvider(provider)
	require.NotEmpty(t, key)

	err := RunOnlyInIntegrationTest(key)
	if err != nil {
		t.Skip(err)
	}

	safescaledLaunched, err := IsSafescaledLaunched()
	if !safescaledLaunched {
		fmt.Println("This requires that you launch safescaled in background and set the tenant")
	}
	require.True(t, safescaledLaunched)
	require.Nil(t, err)

	inPath, err := CanBeRun("safescale")
	require.Nil(t, err)

	require.True(t, safescaledLaunched)
	require.True(t, inPath)
}

func Setup(t *testing.T, provider providers.Enum) {
	EnvSetup(t, provider)

	name := nameFromProvider(provider)
	require.NotEmpty(t, name)

	listStr, err := GetOutput("safescale tenant list")
	require.Nil(t, err)
	require.True(t, len(listStr) > 0)

	getStr, err := GetOutput("safescale tenant get")
	if err != nil {
		fmt.Println("This requires that you set the right tenant before launching the tests")
	}
	require.Nil(t, err)
	require.True(t, len(getStr) > 0)

	providerName := os.Getenv(keyFromProvider(provider))
	require.NotEmpty(t, providerName)

	require.True(t, strings.Contains(getStr, providerName))
}

func Basic(t *testing.T, provider providers.Enum) {
	Setup(t, provider)

	names := GetNames("BasicTest", 0, 1, 1, 2, 1, 0)
	names.TearDown()
	defer names.TearDown()

	out, err := GetOutput("safescale network list")
	fmt.Println(out)
	require.Nil(t, err)

	fmt.Println("Creating network " + names.Networks[0])

	out, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.40.0/24")
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.40.0/24")
	fmt.Println(out)
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist"))

	fmt.Println("Creating VM " + names.Hosts[0])

	out, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	fmt.Println(out)
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist") || strings.Contains(out, "already used"))

	out, err = GetOutput("safescale host inspect " + names.Hosts[0])
	fmt.Println(out)
	require.Nil(t, err)

	host0 := HostInfo{}
	_ = json.Unmarshal([]byte(out), &host0)

	fmt.Println("Creating VM ", names.Hosts[1])

	out, err = GetOutput("safescale host create " + names.Hosts[1] + " --public --net " + names.Networks[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale host create " + names.Hosts[1] + " --public --net " + names.Networks[0])
	fmt.Println(out)
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist") || strings.Contains(out, "already used"))

	out, err = GetOutput("safescale share list")
	require.False(t, strings.Contains(out, names.Shares[0]))

	fmt.Println("Creating Share " + names.Shares[0])

	out, err = GetOutput("safescale share create " + names.Shares[0] + " " + names.Hosts[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale share mount " + names.Shares[0] + " " + names.Hosts[1])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale share list")
	require.True(t, strings.Contains(out, names.Shares[0]))

	out, err = GetOutput("safescale share inspect " + names.Shares[0])
	fmt.Println(out)
	require.Nil(t, err)

	require.True(t, strings.Contains(out, names.Shares[0]))
	require.True(t, strings.Contains(out, names.Hosts[0]))
	require.True(t, strings.Contains(out, names.Hosts[1]))

	out, err = GetOutput("safescale share umount " + names.Shares[0] + " " + names.Hosts[1])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale share inspect " + names.Shares[0])
	fmt.Println(out)
	require.Nil(t, err)

	require.True(t, strings.Contains(out, names.Shares[0]))
	require.True(t, strings.Contains(out, names.Hosts[0]))
	require.False(t, strings.Contains(out, names.Hosts[1]))

	out, err = GetOutput("safescale share delete " + names.Shares[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale share list")
	require.False(t, strings.Contains(out, names.Shares[0]))

	out, err = GetOutput("safescale volume list")
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "null"))

	fmt.Println("Creating Volume " + names.Volumes[0])

	out, err = GetOutput("safescale volume create " + names.Volumes[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale volume list")
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, names.Volumes[0]))

	out, err = GetOutput("safescale volume attach " + names.Volumes[0] + " " + names.Hosts[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale volume delete " + names.Volumes[0])
	fmt.Println(out)
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "still attached") || strings.Contains(err.Error(), "still attached"))

	out, err = GetOutput("safescale volume inspect " + names.Volumes[0])
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, names.Hosts[0]))

	out, err = GetOutput("safescale volume  detach " + names.Volumes[0] + " " + names.Hosts[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale volume inspect " + names.Volumes[0])
	fmt.Println(out)
	require.Nil(t, err)
	require.False(t, strings.Contains(out, names.Hosts[0]))

	out, err = GetOutput("safescale volume delete " + names.Volumes[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale volume list")
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "null"))

	out, err = GetOutput("safescale ssh run " + names.Hosts[0] + " -c \"uptime\"")
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, " user"))

	out, err = GetOutput("safescale host delete " + names.Hosts[0])
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "success"))

	out, err = GetOutput("safescale host delete " + names.Hosts[1])
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "success"))

	out, err = GetOutput("safescale host delete gw-" + names.Networks[0])
	fmt.Println(out)
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "gateway"))

	out, err = GetOutput("safescale network delete " + names.Networks[0])
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "success"))

	fmt.Println("Test OK")
}

func BasicPrivate(t *testing.T, provider providers.Enum) {
	Setup(t, provider)

	names := GetNames("BasicTest", 0, 1, 1, 2, 1, 0)
	names.TearDown()
	defer names.TearDown()

	out, err := GetOutput("safescale network list")
	fmt.Println(out)
	require.Nil(t, err)

	fmt.Println("Creating network " + names.Networks[0])

	out, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.70.0/24")
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.70.0/24")
	fmt.Println(out)
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist"))

	fmt.Println("Creating VM " + names.Hosts[0])

	out, err = GetOutput("safescale host create " + names.Hosts[0] + " --net " + names.Networks[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale host create " + names.Hosts[0] + " --net " + names.Networks[0])
	fmt.Println(out)
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist") || strings.Contains(out, "already used"))

	out, err = GetOutput("safescale host inspect " + names.Hosts[0])
	fmt.Println(out)
	require.Nil(t, err)

	host0 := HostInfo{}
	_ = json.Unmarshal([]byte(out), &host0)

	fmt.Println("Creating VM ", names.Hosts[1])

	out, err = GetOutput("safescale host create " + names.Hosts[1] + " --net " + names.Networks[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale host create " + names.Hosts[1] + " --net " + names.Networks[0])
	fmt.Println(out)
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist") || strings.Contains(out, "already used"))

	out, err = GetOutput("safescale share list")
	require.False(t, strings.Contains(out, names.Shares[0]))

	fmt.Println("Creating Share " + names.Shares[0])

	out, err = GetOutput("safescale share create " + names.Shares[0] + " " + names.Hosts[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale share mount " + names.Shares[0] + " " + names.Hosts[1])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale share list")
	require.True(t, strings.Contains(out, names.Shares[0]))

	out, err = GetOutput("safescale share inspect " + names.Shares[0])
	fmt.Println(out)
	require.Nil(t, err)

	require.True(t, strings.Contains(out, names.Shares[0]))
	require.True(t, strings.Contains(out, names.Hosts[0]))
	require.True(t, strings.Contains(out, names.Hosts[1]))

	out, err = GetOutput("safescale share umount " + names.Shares[0] + " " + names.Hosts[1])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale share inspect " + names.Shares[0])
	fmt.Println(out)
	require.Nil(t, err)

	require.True(t, strings.Contains(out, names.Shares[0]))
	require.True(t, strings.Contains(out, names.Hosts[0]))
	require.False(t, strings.Contains(out, names.Hosts[1]))

	out, err = GetOutput("safescale share delete " + names.Shares[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale share list")
	require.False(t, strings.Contains(out, names.Shares[0]))

	out, err = GetOutput("safescale volume list")
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "null"))

	fmt.Println("Creating Volume " + names.Volumes[0])

	out, err = GetOutput("safescale volume create " + names.Volumes[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale volume list")
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, names.Volumes[0]))

	out, err = GetOutput("safescale volume attach " + names.Volumes[0] + " " + names.Hosts[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale volume delete " + names.Volumes[0])
	fmt.Println(out)
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "still attached"))

	out, err = GetOutput("safescale volume inspect " + names.Volumes[0])
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, host0.ID) || strings.Contains(out, names.Hosts[0]))

	out, err = GetOutput("safescale volume detach " + names.Volumes[0] + " " + names.Hosts[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale volume inspect " + names.Volumes[0])
	fmt.Println(out)
	require.Nil(t, err)
	require.False(t, strings.Contains(out, names.Hosts[0]))

	out, err = GetOutput("safescale volume delete " + names.Volumes[0])
	fmt.Println(out)
	require.Nil(t, err)

	out, err = GetOutput("safescale volume list")
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "null"))

	out, err = GetOutput("safescale ssh run " + names.Hosts[0] + " -c \"uptime\"")
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, " user"))

	out, err = GetOutput("safescale host delete " + names.Hosts[0])
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "success"))

	out, err = GetOutput("safescale host delete " + names.Hosts[1])
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "success"))

	out, err = GetOutput("safescale host delete gw-" + names.Networks[0])
	fmt.Println(out)
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "gateway"))

	out, err = GetOutput("safescale network delete " + names.Networks[0])
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "success"))

	fmt.Println("Test OK")
}

func ReadyToSSH(t *testing.T, provider providers.Enum) {
	Setup(t, provider)

	names := GetNames("ReadyToSSH", 0, 0, 0, 1, 1, 0)
	names.TearDown()
	defer names.TearDown()

	_, err := GetOutput("safescale network list")
	require.Nil(t, err)

	fmt.Println("Creating network " + names.Networks[0] + " --cidr 192.168.41.0/24")

	_, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.41.0/24")
	require.Nil(t, err)

	fmt.Println("Creating VM " + names.Hosts[0])

	_, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	require.Nil(t, err)

	out, err := GetOutput("safescale host inspect " + names.Hosts[0])
	require.Nil(t, err)

	fmt.Println(out)
}

func SharePartialError(t *testing.T, provider providers.Enum) {
	Setup(t, provider)

	names := GetNames("SharePartialError", 0, 1, 1, 1, 1, 0)
	names.TearDown()
	defer names.TearDown()

	out, err := GetOutput("safescale network list")
	_ = out
	require.Nil(t, err)

	fmt.Println("Creating network " + names.Networks[0])

	_, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.49.0/24")
	require.Nil(t, err)

	fmt.Println("Creating VM " + names.Hosts[0])

	_, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	require.Nil(t, err)

	_, err = GetOutput("safescale host inspect " + names.Hosts[0])
	require.Nil(t, err)

	fmt.Println("Creating Share " + names.Shares[0])

	_, err = GetOutput("safescale share create " + names.Shares[0] + " " + names.Hosts[0])
	require.Nil(t, err)

	_, err = GetOutput("safescale share delete " + names.Shares[0])
	require.Nil(t, err)

	out, err = GetOutput("safescale host delete " + names.Hosts[0])
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(out)
	}
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "success"))

	out, err = GetOutput("safescale network delete " + names.Networks[0])
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "success"))
}

func ShareError(t *testing.T, provider providers.Enum) {
	Setup(t, provider)

	names := GetNames("ShareError", 0, 1, 1, 1, 1, 0)
	names.TearDown()
	defer names.TearDown()

	_, err := GetOutput("safescale network list")
	require.Nil(t, err)

	fmt.Println("Creating network " + names.Networks[0])

	_, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.42.0/24")
	require.Nil(t, err)

	fmt.Println("Creating VM " + names.Hosts[0])

	_, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	require.Nil(t, err)

	_, err = GetOutput("safescale host inspect " + names.Hosts[0])
	require.Nil(t, err)

	fmt.Println("Creating Share " + names.Shares[0])

	_, err = GetOutput("safescale share create " + names.Shares[0] + " " + names.Hosts[0])
	require.Nil(t, err)

	fmt.Println("Creating Volume " + names.Volumes[0])

	_, err = GetOutput("safescale volume create --speed SSD " + names.Volumes[0])
	require.Nil(t, err)

	out, err := GetOutput("safescale volume list")
	require.Nil(t, err)
	require.True(t, strings.Contains(out, names.Volumes[0]))

	_, err = GetOutput("safescale volume attach " + names.Volumes[0] + " " + names.Hosts[0])
	require.Nil(t, err)

	time.Sleep(temporal.GetDefaultDelay())

	out, err = GetOutput("safescale volume delete " + names.Volumes[0])
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "still attached"))

	time.Sleep(temporal.GetDefaultDelay())

	_, err = GetOutput("safescale volume detach " + names.Volumes[0] + " " + names.Hosts[0])
	require.Nil(t, err)

	time.Sleep(temporal.GetDefaultDelay())

	_, err = GetOutput("safescale volume delete " + names.Volumes[0])
	require.Nil(t, err)

	time.Sleep(temporal.GetDefaultDelay())

	out, err = GetOutput("safescale volume list")
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "null"))

	out, err = GetOutput("safescale ssh run " + names.Hosts[0] + " -c \"uptime\"")
	require.NoError(t, err)
	require.True(t, strings.Contains(out, " user"))

	_, err = GetOutput("safescale share delete " + names.Shares[0])
	require.NoError(t, err)

	time.Sleep(temporal.GetDefaultDelay())

	out, err = GetOutput("safescale host delete " + names.Hosts[0])
	require.NoError(t, err)
	require.True(t, strings.Contains(out, "success"))

	out, err = GetOutput("safescale network delete " + names.Networks[0])
	require.NoError(t, err)
	require.True(t, strings.Contains(out, "success"))
}

func VolumeError(t *testing.T, provider providers.Enum) {
	Setup(t, provider)

	names := GetNames("VolumeError", 0, 1, 1, 1, 1, 0)
	names.TearDown()
	defer names.TearDown()

	_, err := GetOutput("safescale network list")
	require.NoError(t, err)

	fmt.Println("Creating network " + names.Networks[0])

	_, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.43.0/24")
	require.NoError(t, err)

	fmt.Println("Creating VM " + names.Hosts[0])

	_, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	require.NoError(t, err)

	_, err = GetOutput("safescale host inspect " + names.Hosts[0])
	require.NoError(t, err)

	fmt.Println("Creating Share " + names.Shares[0])

	_, err = GetOutput("safescale share create " + names.Shares[0] + " " + names.Hosts[0])
	require.NoError(t, err)

	fmt.Println("Creating Volume " + names.Volumes[0])

	_, err = GetOutput("safescale volume create " + names.Volumes[0])
	require.NoError(t, err)

	time.Sleep(temporal.GetDefaultDelay())

	_, err = GetOutput("safescale volume  attach " + names.Volumes[0] + " " + names.Hosts[0])
	require.NoError(t, err)

	time.Sleep(temporal.GetDefaultDelay())

	out, err := GetOutput("safescale volume delete " + names.Volumes[0])
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "still attached"))
}

func StopStart(t *testing.T, provider providers.Enum) {
	Setup(t, provider)

	names := GetNames("StopStart", 0, 1, 1, 1, 1, 0)
	names.TearDown()
	defer names.TearDown()

	_, err := GetOutput("safescale network list")
	require.NoError(t, err)

	fmt.Println("Creating network " + names.Networks[0])

	_, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.44.0/24")
	require.Nil(t, err)

	out, err := GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.44.0/24")
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist"))

	fmt.Println("Creating VM " + names.Hosts[0])

	out, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	require.NoError(t, err)
	require.True(t, strings.Contains(out, "success"))
	require.True(t, strings.Contains(out, names.Hosts[0]))

	out, err = GetOutput("safescale host stop " + names.Hosts[0])
	require.True(t, strings.Contains(out, "success"))
	require.NoError(t, err)

	time.Sleep(2 * temporal.GetDefaultDelay())

	out, err = GetOutput("safescale host status " + names.Hosts[0])
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "success"))
	require.True(t, strings.Contains(out, "STOPPED"))

	_, err = GetOutput("safescale host start " + names.Hosts[0])
	require.Nil(t, err)

	time.Sleep(temporal.GetDefaultDelay())

	out, err = GetOutput("safescale ssh run " + names.Hosts[0] + " -c \"uptime\"")
	require.Nil(t, err)
	require.True(t, strings.Contains(out, " user"))

	_, err = GetOutput("safescale host reboot " + names.Hosts[0])
	require.Nil(t, err)

	time.Sleep(temporal.GetDefaultDelay())

	out, err = GetOutput("safescale ssh run " + names.Hosts[0] + " -c \"uptime\"")
	require.Nil(t, err)
	require.True(t, strings.Contains(out, " user"))
}

func DeleteVolumeMounted(t *testing.T, provider providers.Enum) {
	Setup(t, provider)

	names := GetNames("DeleteVolumeMounted", 0, 1, 1, 2, 1, 0)
	names.TearDown()
	defer names.TearDown()

	_, err := GetOutput("safescale network list")
	require.Nil(t, err)

	fmt.Println("Creating network " + names.Networks[0])

	_, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.45.0/24")
	require.Nil(t, err)

	out, err := GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.45.0/24")
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist"))

	fmt.Println("Creating VM " + names.Hosts[0])

	_, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	require.Nil(t, err)

	out, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist") || strings.Contains(out, "already used"))

	_, err = GetOutput("safescale host inspect " + names.Hosts[0])
	require.Nil(t, err)

	fmt.Println("Creating VM " + names.Hosts[1])

	_, err = GetOutput("safescale host create " + names.Hosts[1] + " --public --net " + names.Networks[0])
	require.Nil(t, err)

	out, err = GetOutput("safescale host create " + names.Hosts[1] + " --public --net " + names.Networks[0])
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist") || strings.Contains(out, "already used"))

	fmt.Println("Creating Share " + names.Shares[0])

	_, err = GetOutput("safescale share create " + names.Shares[0] + " " + names.Hosts[0])
	require.Nil(t, err)

	_, err = GetOutput("safescale share mount " + names.Shares[0] + " " + names.Hosts[1])
	require.Nil(t, err)

	out, _ = GetOutput("safescale share list")
	require.True(t, strings.Contains(out, names.Shares[0]))

	out, err = GetOutput("safescale share inspect " + names.Shares[0])
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, names.Shares[0]))
	require.True(t, strings.Contains(out, names.Hosts[0]))
	require.True(t, strings.Contains(out, names.Hosts[1]))

	_, err = GetOutput("safescale share umount " + names.Shares[0] + " " + names.Hosts[1])
	require.Nil(t, err)

	out, err = GetOutput("safescale share inspect " + names.Shares[0])
	fmt.Println(out)
	require.Nil(t, err)
	require.True(t, strings.Contains(out, names.Shares[0]))
	require.True(t, strings.Contains(out, names.Hosts[0]))
	require.False(t, strings.Contains(out, names.Hosts[1]))

	_, err = GetOutput("safescale share delete " + names.Shares[0])
	require.Nil(t, err)

	out, err = GetOutput("safescale share list")
	require.False(t, strings.Contains(out, names.Shares[0]))

	out, err = GetOutput("safescale volume list")
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "null"))

	fmt.Println("Creating Volume " + names.Volumes[0])

	_, err = GetOutput("safescale volume create " + names.Volumes[0])
	require.Nil(t, err)

	out, err = GetOutput("safescale volume list")
	require.Nil(t, err)
	require.True(t, strings.Contains(out, names.Volumes[0]))

	_, err = GetOutput("safescale volume attach " + names.Volumes[0] + " " + names.Hosts[0])
	require.Nil(t, err)

	out, err = GetOutput("safescale volume delete " + names.Volumes[0])
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "still attached to ") || strings.Contains(err.Error(), "still attached to"))
}

func UntilShare(t *testing.T, provider providers.Enum) {
	Setup(t, provider)

	names := GetNames("UntilShare", 0, 1, 1, 2, 1, 0)
	names.TearDown()
	defer names.TearDown()

	_, err := GetOutput("safescale network list")
	require.Nil(t, err)

	fmt.Println("Creating network " + names.Networks[0])

	_, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.46.0/24")
	require.Nil(t, err)

	out, err := GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.46.0/24")
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist"))

	fmt.Println("Creating VM " + names.Hosts[0])

	_, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	require.Nil(t, err)

	out, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist") || strings.Contains(out, "already used"))

	_, err = GetOutput("safescale host inspect " + names.Hosts[0])
	require.Nil(t, err)

	fmt.Println("Creating VM " + names.Hosts[1])

	out, err = GetOutput("safescale host create " + names.Hosts[1] + " --public --net " + names.Networks[0])
	if err != nil {
		fmt.Println(err)
	}
	require.Nil(t, err)

	out, err = GetOutput("safescale host create " + names.Hosts[1] + " --public --net " + names.Networks[0])
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist") || strings.Contains(out, "already used"))

	out, err = GetOutput("safescale share list")
	require.False(t, strings.Contains(out, names.Shares[0]))

	fmt.Println("Creating Share " + names.Shares[0])

	out, err = GetOutput("safescale share create " + names.Shares[0] + " " + names.Hosts[0])
	require.NoError(t, err)
}

func UntilVolume(t *testing.T, provider providers.Enum) {
	Setup(t, provider)

	names := GetNames("UntilVolume", 0, 1, 1, 2, 1, 0)
	names.TearDown()
	defer names.TearDown()

	_, err := GetOutput("safescale network list")
	require.Nil(t, err)

	fmt.Println("Creating network " + names.Networks[0])

	_, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.47.0/24")
	require.Nil(t, err)

	out, err := GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.47.0/24")
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist"))

	fmt.Println("Creating VM " + names.Hosts[0])

	_, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	require.Nil(t, err)

	out, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist") || strings.Contains(out, "already used"))

	_, err = GetOutput("safescale host inspect " + names.Hosts[0])
	require.Nil(t, err)

	fmt.Println("Creating VM " + names.Hosts[1])

	_, err = GetOutput("safescale host create " + names.Hosts[1] + " --public --net " + names.Networks[0])
	require.Nil(t, err)

	out, err = GetOutput("safescale host create " + names.Hosts[1] + " --public --net " + names.Networks[0])
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist") || strings.Contains(out, "already used"))

	out, err = GetOutput("safescale volume list")
	require.Nil(t, err)
	require.True(t, strings.Contains(out, "null"))

	fmt.Println("Creating Volume " + names.Volumes[0])

	_, err = GetOutput("safescale volume create " + names.Volumes[0])
	require.Nil(t, err)

	time.Sleep(temporal.GetDefaultDelay())

	out, err = GetOutput("safescale volume list")
	require.Nil(t, err)
	require.True(t, strings.Contains(out, names.Volumes[0]))
}

func ShareVolumeMounted(t *testing.T, provider providers.Enum) {
	Setup(t, provider)

	names := GetNames("ShareVolumeMounted", 0, 1, 1, 2, 1, 0)
	names.TearDown()
	// defer names.TearDown()

	_, err := GetOutput("safescale network list")
	require.Nil(t, err)

	fmt.Println("Creating network " + names.Networks[0])

	_, err = GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.38.0/24")
	require.Nil(t, err)

	out, err := GetOutput("safescale network create " + names.Networks[0] + " --cidr 192.168.38.0/24")
	require.NotNil(t, err)
	require.True(t, strings.Contains(out, "already exist"))

	fmt.Println("Creating VM " + names.Hosts[0])

	_, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
	require.Nil(t, err)

	/*
		out, err = GetOutput("safescale host create " + names.Hosts[0] + " --public --net " + names.Networks[0])
		require.NotNil(t, err)
		require.True(t, strings.Contains(out, "already exist") || strings.Contains(out, "already used"))

		_, err = GetOutput("safescale host inspect " + names.Hosts[0])
		require.Nil(t, err)

		fmt.Println("Creating VM " + names.Hosts[1])

		_, err = GetOutput("safescale host create " + names.Hosts[1] + " --public --net " + names.Networks[0])
		require.Nil(t, err)

		out, err = GetOutput("safescale host create " + names.Hosts[1] + " --public --net " + names.Networks[0])
		require.NotNil(t, err)
		require.True(t, strings.Contains(out, "already exist") || strings.Contains(out, "already used"))

		_, err = GetOutput("safescale share list")
		require.Nil(t, err)

		fmt.Println("Creating Share " + names.Shares[0])

		_, err = GetOutput("safescale share create " + names.Shares[0] + " " + names.Hosts[0])
		require.Nil(t, err)

		_, err = GetOutput("safescale share mount " + names.Shares[0] + " " + names.Hosts[1])
		require.Nil(t, err)
	*/
}
