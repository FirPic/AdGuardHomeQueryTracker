package querytracker

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func detect_ip_type(ip string) (string, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", fmt.Errorf("Invalid IP address: %s", ip)
	}

	if parsedIP.To4() != nil {
		return "IPv4", nil
	}
	return "IPv6", nil
}

func is_ip_in_range(ip string, cidr string) (bool, error) {
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false, fmt.Errorf("Invalid IP address: %s", ip)
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("Invalid CIDR range: %s", cidr)
	}

	if ipNet.Contains(ipAddr) {
		return true, nil
	}
	return false, nil
}

func get_ip() string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Enter IP: ")
	var input string

	if scanner.Scan() {
		input = scanner.Text()
	} else {
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Reading error:", err)
		}
	}
	return input
}

func processCheckIp(ip string) error {
	files_tested := 0

	ip_type, err := detect_ip_type(ip)
	if err != nil {
		return fmt.Errorf("Error: %w", err)
	}

	var directory string
	if ip_type == "IPv4" {
		directory = "./ipv4/"
	} else if ip_type == "IPv6" {
		directory = "./ipv6/"
	} else {
		return fmt.Errorf("Unknown IP type: %s", ip_type)
	}

	files, err := filepath.Glob(filepath.Join(directory, "*.zone"))
	if err != nil {
		return fmt.Errorf("Error searching files: %w", err)
	}

	for _, file := range files {
		if file != filepath.Join("ipv4", "zz.zone") && file != filepath.Join("ipv6", "zz.zone") {
			files_tested++
			country_code := file[(strings.Index(file, "/") + 1):]
			country_code = country_code[:strings.Index(country_code, ".")]

			country_file, err := os.Open(file)
			if err != nil {
				fmt.Printf("Error opening file %s: %v\n", file, err)
				continue
			}
			defer country_file.Close()

			scanner_file := bufio.NewScanner(country_file)
			for scanner_file.Scan() {
				cidr := scanner_file.Text()
				rlt, err := is_ip_in_range(ip, cidr)
				if err != nil {
					fmt.Printf("Error checking IP in file %s: %v\n", file, err)
					continue
				}

				if rlt {
					fmt.Printf("%s - %s\n", country_code, cidr)
					break
				}
			}
			if err := scanner_file.Err(); err != nil {
				fmt.Printf("Error reading file %s: %v\n", file, err)
			}
		}
	}

	fmt.Printf("\nNumber of files tested: %d\n", files_tested)
	return nil
}

// func main() {
// 	ip := get_ip()
// 	err := processCheckIp(ip)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// }
