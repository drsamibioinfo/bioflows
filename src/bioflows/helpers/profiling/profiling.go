package profiling

import (
	"github.com/bioflows/src/bioflows/models"
	"log"
	"net"
	"runtime"
)

func GetLocalAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func GetCPU() int {
	return runtime.NumCPU()
}

func GetCPUProfile() models.CPUProfile {
	profile := models.CPUProfile{}
	profile.Memstats = &runtime.MemStats{}
	runtime.ReadMemStats(profile.Memstats)
	profile.Addr = GetLocalAddress()
	profile.CPU = GetCPU()
	return profile
}