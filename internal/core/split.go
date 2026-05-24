package core

import (
	"context"
	"fmt"
	"io"
	"os"
)

func Split(ctx context.Context, sourcePath string, chunkSize int64, progress func(current, total int64)) ([]string, error) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat source file: %w", err)
	}
	totalSize := sourceInfo.Size()

	var parts []string
	buf := make([]byte, 32*1024)
	partNum := 1
	var totalWritten int64
	done := false

	for !done {
		select {
		case <-ctx.Done():
			cleanupParts(parts)
			return parts, ctx.Err()
		default:
		}

		outPath := fmt.Sprintf("%s.%03d", sourcePath, partNum)
		outFile, err := os.Create(outPath)
		if err != nil {
			return parts, fmt.Errorf("failed to create part %s: %w", outPath, err)
		}

		var written int64
		for written < chunkSize {
			select {
			case <-ctx.Done():
				outFile.Close()
				os.Remove(outPath)
				cleanupParts(parts)
				return parts, ctx.Err()
			default:
			}

			remaining := chunkSize - written
			readSize := int64(len(buf))
			if readSize > remaining {
				readSize = remaining
			}

			n, readErr := sourceFile.Read(buf[:readSize])
			if n > 0 {
				if _, werr := outFile.Write(buf[:n]); werr != nil {
					outFile.Close()
					os.Remove(outPath)
					cleanupParts(parts)
					return parts, fmt.Errorf("failed to write to %s: %w", outPath, werr)
				}
				written += int64(n)
				totalWritten += int64(n)
				if progress != nil {
					progress(totalWritten, totalSize)
				}
			}
			if readErr == io.EOF {
				done = true
				break
			}
			if readErr != nil {
				outFile.Close()
				os.Remove(outPath)
				cleanupParts(parts)
				return parts, fmt.Errorf("error reading source: %w", readErr)
			}
		}

		outFile.Close()

		if written > 0 {
			parts = append(parts, outPath)
		} else {
			os.Remove(outPath)
		}

		if done {
			break
		}
		partNum++
	}

	return parts, nil
}

func cleanupParts(parts []string) {
	for _, p := range parts {
		os.Remove(p)
	}
}
