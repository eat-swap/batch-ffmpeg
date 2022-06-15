package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/semaphore"
)

const (
	SrcDir  = "D:/tmp/Musik/"
	DstDir  = "D:/tmp/Converted/"
	ExtList = ".flac .mp3 .ogg .wav .wma"
	DstExt  = ".opus"
)

type Task struct {
	Src    string
	Dst    string
	TaskId int
}

var (
	taskChan  = make(chan Task)
	finishSem *semaphore.Weighted
)

func main() {
	threads := runtime.NumCPU()
	finishSem = semaphore.NewWeighted(int64(threads))
	fmt.Printf("Using %d threads\n", threads)
	for i := 0; i < threads; i++ {
		go executor()
	}

	taskId := 0
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

		dstDir := filepath.Dir(DstDir + strings.TrimPrefix(strings.ReplaceAll(path, "\\", "/"), SrcDir))
		err = os.MkdirAll(dstDir, 0755)
		if err != nil {
			fmt.Println("Failed to create dir:", err)
			return err
		}

		// Starts converting
		taskId++
		taskChan <- Task{path, dstDir + "/" + baseName + DstExt, taskId}

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
		fmt.Printf("[%d] Converting: %s\n", task.TaskId, task.Src)
		cmd := exec.Command("ffmpeg",
			"-i",
			task.Src,
			"-c:a",
			"libopus",
			"-b:a",
			"64k",
			task.Dst,
		)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Failed to convert %s: %s\n", task.Src, err.Error())
		}
	}
	finishSem.Release(1)
}
