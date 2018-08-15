// +build !windows

package riot

import (
	"log"
	"testing"

	"github.com/vcaesar/tt"
)

// TestSysInfo
func TestMem(t *testing.T) {
	log.Println("SYS info test...")
	var engine Engine

	log.Println("Mem info test...")
	tt.Equal(t, true, InitMemUsed != 0)
	tt.Equal(t, true, InitMemUsed != 0)

	memPercent, err := MemPercent()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, memPercent != "")

	mem, err := MemUsed()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, mem != 0)

	useMem, err := engine.UsedMem()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, useMem != 0)

	memT, err := MemTotal()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, memT != 0)

	memFree, err := MemFree()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, memFree != 0)

	tt.Equal(t, 1, ToKB(1024))
	tt.Equal(t, 1, ToMB(1024*1024))
	tt.Equal(t, 1, ToGB(1024*1024*1024))
}

func TestDisk(t *testing.T) {
	log.Println("Disk info test...")
	var engine Engine

	diskPercent, err := DiskPercent()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, diskPercent != "")

	disk, err := DiskUsed()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, disk != 0)

	useDisk, err := engine.UsedDisk()
	tt.Equal(t, nil, err)
	log.Println("useDisk: ", useDisk)
	// tt.Equal(t, true, useDisk != 0)

	diskTotal, err := DiskTotal()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, diskTotal != 0)

	diskFree, err := DiskFree()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, diskFree != 0)
}

func TestCPU(t *testing.T) {
	log.Println("CPU info test...")

	cpuInfo, err := CPUInfo()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, cpuInfo != "")

	cpuPct, err := CPUPercent()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, cpuPct != nil)
}

func TestPlatform(t *testing.T) {
	log.Println("Platform info test...")

	uptime, err := Uptime()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, uptime != 0)

	platform, family, osVersion, err := PlatformInfo()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, platform != "" && osVersion != "")
	log.Println(family)

	palt, err := Platform()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, palt != "")

	kver, err := KernelVer()
	tt.Equal(t, nil, err)
	tt.Equal(t, true, kver != "")
}
