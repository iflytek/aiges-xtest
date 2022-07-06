package util

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ReadDir 读取文件夹文件
func ReadDir(fi os.FileInfo, src string, cmpFunc func(i, j string) bool, sep string) ([][]byte, error) {
	// 遍历目录文件
	ans := [][]byte{}
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return [][]byte{}, err
	}
	filemap := make(map[string]string)
	// 过滤空文件及子目录
	file_index := []string{}
	for _, file := range files {
		if !file.IsDir() && file.Size() != 0 {
			ext := filepath.Ext(file.Name())                             // 获取文件后缀
			filenamePrefix := file.Name()[0 : len(file.Name())-len(ext)] // 去除文件扩展名
			filename := strings.Split(filenamePrefix, sep)[0]            // 分割文件名称
			file_index = append(file_index, filename)
			filemap[filename] = file.Name()
		}
	}
	rand.Shuffle(len(file_index), func(i, j int) {
		file_index[i], file_index[j] = file_index[j], file_index[i]
	})
	sort.Slice(file_index, func(i, j int) bool {
		if len(file_index[i]) == len(file_index[j]) {
			return cmpFunc(file_index[i], file_index[j])
		}
		return len(file_index[i]) < len(file_index[j])
	})

	fmt.Println(file_index)
	for _, k := range file_index {
		data, err := ioutil.ReadFile(src + "/" + filemap[k])
		if err != nil {
			fmt.Printf("read file %s fail, %s ", src+"/"+filemap[k], err.Error())
			return [][]byte{}, err
		}
		ans = append(ans, data)
	}

	return ans, nil
}
