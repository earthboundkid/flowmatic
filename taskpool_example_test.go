package flowmatic_test

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/earthboundkid/flowmatic"
)

func ExampleTaskPool() {
	// Compare to https://pkg.go.dev/golang.org/x/sync/errgroup#example-Group-Pipeline and https://blog.golang.org/pipelines

	m, err := MD5All(context.Background(), "testdata/md5all")
	if err != nil {
		log.Fatal(err)
	}

	for k, sum := range m {
		fmt.Printf("%s:\t%x\n", k, sum)
	}

	// Output:
	// testdata/md5all/hello.txt:	bea8252ff4e80f41719ea13cdf007273
}

// MD5All reads all the files in the file tree rooted at root
// and returns a map from file path to the MD5 sum of the file's contents.
// If the directory walk fails or any read operation fails,
// MD5All returns an error.
func MD5All(ctx context.Context, root string) (map[string][md5.Size]byte, error) {
	// Make a pool of 20 digesters
	in, out := flowmatic.TaskPool(20, digest)

	m := make(map[string][md5.Size]byte)
	// Open two goroutines:
	// one for reading file names by walking the filesystem
	// one for recording results from the digesters in a map
	err := flowmatic.All(ctx,
		func(ctx context.Context) error {
			return walkFilesystem(ctx, root, in)
		},
		func(ctx context.Context) error {
			for r := range out {
				if r.Err != nil {
					return r.Err
				}
				m[r.In] = *r.Out
			}
			return nil
		},
	)

	return m, err
}

func walkFilesystem(ctx context.Context, root string, in chan<- string) error {
	defer close(in)

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		select {
		case in <- path:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})
}

func digest(path string) (*[md5.Size]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	hash := md5.Sum(data)
	return &hash, nil
}
