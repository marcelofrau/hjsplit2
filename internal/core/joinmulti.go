package core

import (
	"context"
	"fmt"
	"io"
	"os"
)

func JoinMulti(ctx context.Context, files []string, outputPath string, progress func(current, total int64)) error {
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}

	var totalSize int64
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", f, err)
		}
		totalSize += info.Size()
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	var totalCopied int64
	buf := make([]byte, 32*1024)

	for _, part := range files {
		select {
		case <-ctx.Done():
			outFile.Close()
			os.Remove(outputPath)
			return ctx.Err()
		default:
		}

		partFile, err := os.Open(part)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", part, err)
		}

		for {
			select {
			case <-ctx.Done():
				partFile.Close()
				outFile.Close()
				os.Remove(outputPath)
				return ctx.Err()
			default:
			}

			n, readErr := partFile.Read(buf)
			if n > 0 {
				if _, werr := outFile.Write(buf[:n]); werr != nil {
					partFile.Close()
					return fmt.Errorf("failed to write output: %w", werr)
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
				return fmt.Errorf("failed to read %s: %w", part, readErr)
			}
		}
		partFile.Close()
	}

	return nil
}
