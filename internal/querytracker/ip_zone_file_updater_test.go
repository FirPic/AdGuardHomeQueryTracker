package querytracker

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetRemoteChecksum(t *testing.T) {
	// Create an HTTP test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("d41d8cd98f00b204e9800998ecf8427e  all-zones.tar.gz\n"))
	}))
	defer server.Close()

	checksum, err := get_remote_checksum(server.URL, `^(.*all-zones\.tar\.gz.*)$`)
	if err != nil {
		t.Fatalf("Error while getting remote checksum: %v", err)
	}

	expected := "d41d8cd98f00b204e9800998ecf8427e"
	if checksum != expected {
		t.Errorf("Expected checksum %s, got %s", expected, checksum)
	}
}

func TestGetLocalChecksum(t *testing.T) {
	// Create a temporary file
	file, err := os.CreateTemp("", "MD5SUM")
	if err != nil {
		t.Fatalf("Error while creating temporary file: %v", err)
	}
	defer os.Remove(file.Name())

	_, err = file.WriteString("d41d8cd98f00b204e9800998ecf8427e  all-zones.tar.gz\n")
	if err != nil {
		t.Fatalf("Error while writing to temporary file: %v", err)
	}
	file.Close()

	checksum, err := get_local_checksum(file.Name(), `^(.*all-zones\.tar\.gz.*)$`)
	if err != nil {
		t.Fatalf("Error while getting local checksum: %v", err)
	}

	expected := "d41d8cd98f00b204e9800998ecf8427e"
	if checksum != expected {
		t.Errorf("Expected checksum %s, got %s", expected, checksum)
	}
}

func TestComparChecksum(t *testing.T) {
	// Create an HTTP test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("d41d8cd98f00b204e9800998ecf8427e  all-zones.tar.gz\n"))
	}))
	defer server.Close()

	// Create a temporary file
	file, err := os.CreateTemp("", "MD5SUM")
	if err != nil {
		t.Fatalf("Error while creating temporary file: %v", err)
	}
	defer os.Remove(file.Name())

	_, err = file.WriteString("d41d8cd98f00b204e9800998ecf8427e  all-zones.tar.gz\n")
	if err != nil {
		t.Fatalf("Error while writing to temporary file: %v", err)
	}
	file.Close()

	isSame, err := compar_checksum("ipv4", server.URL, file.Name(), `^(.*all-zones\.tar\.gz.*)$`)
	if err != nil {
		t.Fatalf("Error while comparing checksums: %v", err)
	}

	if !isSame {
		t.Errorf("Checksums should be identical")
	}
}

func TestComparChecksumDifference(t *testing.T) {
	// Create an HTTP test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("d41d8cd98f00b204e9800998ecf8427e  all-zones.tar.gz\n"))
	}))
	defer server.Close()

	// Create a temporary file with a different checksum
	file, err := os.CreateTemp("", "MD5SUM")
	if err != nil {
		t.Fatalf("Error while creating temporary file: %v", err)
	}
	defer os.Remove(file.Name())

	_, err = file.WriteString("e99a18c428cb38d5f260853678922e03  all-zones.tar.gz\n")
	if err != nil {
		t.Fatalf("Error while writing to temporary file: %v", err)
	}
	file.Close()

	isSame, err := compar_checksum("ipv4", server.URL, file.Name(), `^(.*all-zones\.tar\.gz.*)$`)
	if err != nil {
		t.Fatalf("Error while comparing checksums: %v", err)
	}

	if isSame {
		t.Errorf("Checksums should not be identical")
	}
}
