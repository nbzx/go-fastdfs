package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// 获取当前执行文件绝对路径
func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

// 获取当前执行文件绝对路径（go run）
func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

// 获取执行目录-全兼容
func getCurrentAbPath() (string, error) {
	dir := getCurrentAbPathByExecutable()
	tmpDir, err := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return getCurrentAbPathByCaller(), err
	}
	return dir, err
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

var dataPath string

func init() {
	var err error
	dataPath, err = getCurrentAbPath()
	if err != nil {
		panic(err)
	}
}

func main() {
	var err error
	if len(os.Args) > 1 {
		fmt.Println("设置启动目录: ", os.Args[1])
		dataPath, err = filepath.EvalSymlinks(os.Args[1])
		if err != nil {
			panic(err)
		}
	} else {
		env_storage_dir := os.Getenv("GO_FASTDFS_DIR")
		if env_storage_dir == "" {
			fmt.Println("请指定数据存储目录, 用环境变量GO_FASTDFS_DIR=/home/dfs_storage, 或者参数如/home/dfs_storage")
			return
		}
	}
	InitServer(dataPath)
	ctx := context.Background()
	Start(ctx)
}
