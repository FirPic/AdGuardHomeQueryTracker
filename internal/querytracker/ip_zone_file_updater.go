package querytracker

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func get_remote_checksum(urlStr, pattern string) (string, error) {
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return "", fmt.Errorf("error during GET request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = fmt.Errorf("error closing response body: %w", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	md5sum := string(body)
	checksums_regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("error: regex compilation error: %w", err)
	}

	var checksum string
	for _, line := range strings.Split(md5sum, "\n") {
		if checksums_regex.MatchString(line) {
			checksum_regex, err := regexp.Compile(`^([a-fA-F0-9]+)\s+.*\.tar\.gz$`)
			if err != nil {
				return "", fmt.Errorf("error: regex compilation error: %w", err)
			}

			match := checksum_regex.FindStringSubmatch(line)
			if len(match) > 1 {
				checksum = match[1]
			}
		}
	}

	return checksum, nil
}

func get_local_checksum(path, pattern string) (string, error) {
	md5sum_file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer func() {
		if cerr := md5sum_file.Close(); cerr != nil {
			err = fmt.Errorf("error closing file: %w", cerr)
		}
	}()

	md5sum_scanner := bufio.NewScanner(md5sum_file)
	checksums_regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("error: regex compilation error: %w", err)
	}

	var checksum string
	for md5sum_scanner.Scan() {
		if checksums_regex.MatchString(md5sum_scanner.Text()) {
			checksum_regex, err := regexp.Compile(`^([a-fA-F0-9]+)\s+.*\.tar\.gz$`)
			if err != nil {
				return "", fmt.Errorf("error: regex compilation error: %w", err)
			}

			match := checksum_regex.FindStringSubmatch(md5sum_scanner.Text())
			if len(match) > 1 {
				checksum = match[1]
			}
		}
	}

	if err := md5sum_scanner.Err(); err != nil {
		return "", fmt.Errorf("error scanning file: %w", err)
	}

	return checksum, nil
}

func compar_checksum(ip_protocol, remote_url, local_path, pattern string) (bool, error) {
	rchecksum, err := get_remote_checksum(remote_url, pattern)
	if err != nil {
		return false, fmt.Errorf("error getting remote checksum: %w", err)
	}

	lchecksum, err := get_local_checksum(local_path, pattern)
	if err != nil {
		return false, fmt.Errorf("error getting local checksum: %w", err)
	}

	return rchecksum == lchecksum, nil
}

func downloadArchive(urlStr, destPath string) error {
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return fmt.Errorf("error during GET request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = fmt.Errorf("Error closing response body: %w", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error: Status code %d", resp.StatusCode)
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("Error creating destination file: %w", err)
	}
	defer func() {
		if cerr := outFile.Close(); cerr != nil {
			err = fmt.Errorf("Error closing file: %w", cerr)
		}
	}()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("Error copying data: %w", err)
	}

	return nil
}

func extractArchive(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("Error opening archive file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			err = fmt.Errorf("Error closing file: %w", cerr)
		}
	}()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("Error creating destination directory: %w", err)
	}

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("Error creating gzip reader: %w", err)
	}
	defer func() {
		if cerr := gzr.Close(); cerr != nil {
			err = fmt.Errorf("Error closing gzip reader: %w", cerr)
		}
	}()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("Error reading tar header: %w", err)
		}

		if header.Name == "./" {
			continue
		}

		destPath := filepath.Join(destDir, header.Name)
		destPath = filepath.Clean(destPath)
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("Invalid file path: %s", destPath)
		}

		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("Error creating directory: %w", err)
			}
			continue
		}

		outFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("Error creating file: %w", err)
		}
		defer func() {
			if cerr := outFile.Close(); cerr != nil {
				err = fmt.Errorf("Error closing file: %w", cerr)
			}
		}()

		if _, err := io.Copy(outFile, tr); err != nil {
			return fmt.Errorf("Error copying data: %w", err)
		}
	}

	return nil
}

func downloadFile(urlStr, destPath string) error {
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return fmt.Errorf("Invalid URL: %w", err)
	}

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return fmt.Errorf("Error during GET request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = fmt.Errorf("Error closing response body: %w", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error: Status code %d", resp.StatusCode)
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("Error creating destination file: %w", err)
	}
	defer func() {
		if cerr := outFile.Close(); cerr != nil {
			err = fmt.Errorf("Error closing file: %w", cerr)
		}
	}()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("Error copying data: %w", err)
	}

	return nil
}

func processChecksum(ip_protocol, remote_url, local_path, archive_url, archive_path, dest_dir, pattern string) (bool, error) {
	if _, err := os.Stat(dest_dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dest_dir, 0755); err != nil {
			return false, fmt.Errorf("Error creating directory: %w", err)
		}
	}

	if _, err := os.Stat(local_path); os.IsNotExist(err) {
		err = downloadArchive(archive_url, archive_path)
		if err != nil {
			return false, fmt.Errorf("Error downloading archive: %w", err)
		}

		err = extractArchive(archive_path, dest_dir)
		if err != nil {
			return false, fmt.Errorf("Error extracting archive: %w", err)
		}

		if ip_protocol == "ipv4" {
			err = downloadFile(remote_url, local_path)
			if err != nil {
				return false, fmt.Errorf("Error downloading MD5SUM file: %w", err)
			}
		}

		return true, nil
	}

	isSame, err := compar_checksum(ip_protocol, remote_url, local_path, pattern)
	if err != nil {
		return false, fmt.Errorf("Error comparing checksums: %w", err)
	}
	if isSame {
		return false, nil
	}

	err = downloadArchive(archive_url, archive_path)
	if err != nil {
		return false, fmt.Errorf("Error downloading archive: %w", err)
	}

	err = extractArchive(archive_path, dest_dir)
	if err != nil {
		return false, fmt.Errorf("Error extracting archive: %w", err)
	}

	if ip_protocol == "ipv4" {
		err = downloadFile(remote_url, local_path)
		if err != nil {
			return false, fmt.Errorf("Error downloading MD5SUM file: %w", err)
		}
	}

	return true, nil
}

// func main() {
// 	updated, err := processChecksum("ipv4", "https://www.ipdeny.com/ipblocks/data/countries/MD5SUM", "ipv4/MD5SUM", "https://www.ipdeny.com/ipblocks/data/countries/all-zones.tar.gz", "all-zones.tar.gz", "ipv4/", `^(.*all-zones\.tar\.gz.*)$`)
// 	if err != nil {
// 		fmt.Println(err)
// 	} else if updated {
// 		fmt.Println("ipv4 updated")
// 	}

// 	updated, err = processChecksum("ipv6", "https://www.ipdeny.com/ipv6/ipaddresses/blocks/MD5SUM", "ipv6/MD5SUM", "https://www.ipdeny.com/ipv6/ipaddresses/blocks/ipv6-all-zones.tar.gz", "ipv6-all-zones.tar.gz", "ipv6/", `^(.*ipv6-all-zones\.tar\.gz.*)$`)
// 	if err != nil {
// 		fmt.Println(err)
// 	} else if updated {
// 		fmt.Println("ipv6 updated")
// 	}
// }
