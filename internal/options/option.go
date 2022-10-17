package options

import (
	"flag"
	"log"
	"strconv"
	"strings"
)

func init() {
	flag.IntVar(&port, "p", 0, "Server port. Required")
	flag.StringVar(&adsIpPort, "d", "", "List ip: port of advertising services. Required. Format (comma, without spaces) - ip:port,ip:port...")
}

var (
	port      int
	adsIpPort string
)

const minPort, maxPort = 1, 65535

var requiredFlags = []string{"p", "d"}

func BuildOptions() (int, []string) {
	flag.Parse()
	currentFlags := make(map[string]struct{})
	flag.Visit(func(f *flag.Flag) { currentFlags[f.Name] = struct{}{} })
	adIpPorts := make([]string, 1)
	for _, req := range requiredFlags {
		if _, ok := currentFlags[req]; !ok {
			log.Fatalf("missing required -%s argument/flag\n", req)
		}
		if req == "p" {
			portValidation(port)
		}
		if req == "d" {
			adIpPorts = strings.Split(adsIpPort, ",")
			for _, v := range adIpPorts {
				splittedAddress := strings.Split(v, ":")
				if len(splittedAddress) != 2 {
					log.Fatalf("incorrect ip:port - %s\n", v)
				}
				adPort, err := strconv.Atoi(splittedAddress[1])
				if err != nil {
					log.Fatalf("incorrect ip:port - %s\n", v)
				}
				portValidation(adPort)
			}
		}
	}
	return port, adIpPorts
}

func portValidation(port int) {
	if port < minPort || port > maxPort {
		log.Fatalf("incorrect port - %d\n", port)
	}
}
