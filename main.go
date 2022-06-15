package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/semaphore"
)

const (
	SrcDir  = "D:/tmp/Music/"
	DstDir  = "D:/tmp/Converted/"
	ExtList = ".flac .mp3 .ogg .wav .wma"
	DstExt  = ".opus"
)

type Task struct {
	Src string
	Dst string
}

var (
	taskChan  = make(chan Task)
	finishSem *semaphore.Weighted
)

func main() {
	threads := 1 // runtime.NumCPU()
	finishSem = semaphore.NewWeighted(int64(threads))
	fmt.Printf("Using %d threads\n", threads)
	for i := 0; i < threads; i++ {
		go executor()
	}

	filepath.Walk(SrcDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		name := info.Name()
		ext := filepath.Ext(name)
		baseName := strings.TrimSuffix(name, ext)

		if ext == "" || !strings.Contains(ExtList, ext) {
			return nil
		}

		err = os.MkdirAll(filepath.Dir(DstDir+strings.TrimPrefix(strings.ReplaceAll(path, "\\", "/"), SrcDir)), 0755)
		if err != nil {
			fmt.Println("Failed to create dir:", err)
			return err
		}

		// Starts converting
		taskChan <- Task{path, filepath.Dir(DstDir+path) + "/" + baseName + DstExt}

		return nil
	})

	// Close taskChan to signal all goroutines to exit
	close(taskChan)

	// Wait for all tasks to finish
	finishSem.Acquire(context.Background(), int64(threads))
}

func executor() {
	finishSem.Acquire(context.Background(), 1)
	for task := range taskChan {
		fmt.Printf("Converting: %s -> %s\n", task.Src, task.Dst)
	}
	finishSem.Release(1)
}
