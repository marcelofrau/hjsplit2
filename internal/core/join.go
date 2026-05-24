package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Join(ctx context.Context, firstPartPath string, progress func(current, total int64)) (string, error) {
	basePath := stripPartSuffix(firstPartPath)
	if basePath == firstPartPath {
		return "", fmt.Errorf("file must have a .001 extension")
	}

	dir := filepath.Dir(basePath)
	base := filepath.Base(basePath)
	ext := filepath.Ext(base)
	stem := base[:len(base)-len(ext)]

	var joinedName string
	if ext == "" {
		joinedName = stem + "_restored"
	} else {
		joinedName = stem + "_restored" + ext
	}
	joinedPath := filepath.Join(dir, joinedName)

	// Calculate total size of all parts
	var totalSize int64
	partNum := 1
	for {
		partPath := fmt.Sprintf("%s.%03d", basePath, partNum)
		info, err := os.Stat(partPath)
		if err != nil {
			break
		}
		totalSize += info.Size()
		partNum++
	}
	if partNum == 1 {
		return "", fmt.Errorf("no part files found")
	}

	joinedFile, err := os.Create(joinedPath)
	if err != nil {
		return "", fmt.Errorf("failed to create joined file: %w", err)
	}
	defer joinedFile.Close()

	var totalCopied int64
	buf := make([]byte, 32*1024)

	partNum = 1
	for {
		select {
		case <-ctx.Done():
			joinedFile.Close()
			os.Remove(joinedPath)
			return "", ctx.Err()
		default:
		}

		partPath := fmt.Sprintf("%s.%03d", basePath, partNum)
		partFile, err := os.Open(partPath)
		if err != nil {
			if os.IsNotExist(err) {
				break
			}
			return "", fmt.Errorf("failed to open part %s: %w", partPath, err)
		}

		for {
			select {
			case <-ctx.Done():
				partFile.Close()
				joinedFile.Close()
				os.Remove(joinedPath)
				return "", ctx.Err()
			default:
			}

			n, readErr := partFile.Read(buf)
			if n > 0 {
				if _, werr := joinedFile.Write(buf[:n]); werr != nil {
					partFile.Close()
					joinedFile.Close()
					os.Remove(joinedPath)
					return "", fmt.Errorf("failed to write joined file: %w", werr)
				}
				totalCopied += int64(n)
				if progress != nil {
					progress(totalCopied, totalSize)
				}
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				partFile.Close()
				joinedFile.Close()
				os.Remove(joinedPath)
				return "", fmt.Errorf("failed to read %s: %w", partPath, readErr)
			}
		}
		partFile.Close()
		partNum++
	}

	return joinedPath, nil
}

func stripPartSuffix(path string) string {
	ext := filepath.Ext(path)
	if ext == ".001" {
		return path[:len(path)-len(ext)]
	}
	return path
}
