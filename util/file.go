package util

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strings"
)

// ReadDir 读取文件夹文件
func ReadDir(fi os.FileInfo, src string, cmpFunc func(i, j string) bool) ([][]byte, error) {
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
			filename := strings.Split(file.Name(), "_")[0]
			file_index = append(file_index, filename)
			filemap[filename] = file.Name()
		}
	}
	rand.Shuffle(len(file_index), func(i, j int) {
		file_index[i], file_index[j] = file_index[j], file_index[i]
	})
	sort.Slice(file_index, func(i, j int) bool {
		return cmpFunc(file_index[i], file_index[j])
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
