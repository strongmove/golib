package fileutils

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

func GetAllFiles(targetDir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := strings.ToLower(d.Name())
		if !d.IsDir() && (strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".rar") || strings.HasSuffix(name, ".mp4")) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func GetSizeOnDisk(path string) (uint64, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetCompressedFileSizeW")
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var high uint32
	r1, _, err := proc.Call(uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(&high)))
	if r1 == 0xFFFFFFFF {
		if err != syscall.Errno(0) {
			return 0, err
		}
	}
	return (uint64(high) << 32) + uint64(r1), nil
}

func testDiskSizesMatch(filePath string) (bool, map[string]uint64) {
	info, err := os.Stat(filePath)
	if err != nil {
		return false, map[string]uint64{}
	}
	size := uint64(info.Size())
	sizeOnDisk, err := GetSizeOnDisk(filePath)
	if err != nil {
		return false, map[string]uint64{}
	}
	sizeDifference := size - sizeOnDisk
	match := size == sizeOnDisk
	data := map[string]uint64{
		"size":           size,
		"sizeOnDisk":     sizeOnDisk,
		"sizeDifference": sizeDifference,
	}
	return match, data
}

func GetMismatchedFiles(files []string) []string {
	var mismatches []string
	for _, file := range files {
		match, _ := testDiskSizesMatch(file)
		if !match {
			mismatches = append(mismatches, file)
		}
	}
	return mismatches
}
