package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var infoTemp string
var selectedFileType string
var fileTypeExtensions = map[string][]string{
	"PHP":        {".php"},
	"Python":     {".py"},
	"Bat":        {".bat", ".cmd"},
	"JAVA":       {".java"},
	"CSharp":     {".cs"},
	"NodeJS":     {".js"},
	"JavaScript": {".js"},
	"Vue":        {".vue"},
	"Go":         {".go"},
	"TypeScript": {".ts", ".tsx"},
	"Shell":      {".sh", ".bash"},
	"SQL":        {".sql"},
}

func outPutInfo(info string) {
	fmt.Println(info)
	infoTemp += info + "\n"
}

func generateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

type BrowserOpener struct{}

func (b *BrowserOpener) Open(url string, browser string, newTab bool) bool {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		if strings.ToLower(browser) == "edge" {
			// Windows Edge 浏览器
			cmd = exec.Command("cmd", "/c", "start", "microsoft-edge:"+url)
		} else {
			// 默认浏览器
			cmd = exec.Command("cmd", "/c", "start", "", url)
		}
	case "darwin":
		// macOS
		cmd = exec.Command("open", url)
	case "linux":
		// Linux
		cmd = exec.Command("xdg-open", url)
	default:
		fmt.Printf("不支持的操作系统: %s\n", runtime.GOOS)
		return false
	}

	err := cmd.Run()
	if err != nil {
		fmt.Printf("打开浏览器失败: %v\n", err)
		return false
	}
	return true
}

func selectSingleFileType() string {
	fmt.Println("===============================================================")
	fmt.Println("                   请选择要扫描的文件类型")
	fmt.Println("===============================================================")
	fmt.Println("说明：请输入数字选择一种文件类型")
	fmt.Println("===============================================================")

	// 创建类型列表并排序
	typeList := make([]string, 0, len(fileTypeExtensions))
	for typeName := range fileTypeExtensions {
		typeList = append(typeList, typeName)
	}

	// 按字母顺序排序
	for i := 0; i < len(typeList)-1; i++ {
		for j := i + 1; j < len(typeList); j++ {
			if typeList[i] > typeList[j] {
				typeList[i], typeList[j] = typeList[j], typeList[i]
			}
		}
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// 显示类型列表
		fmt.Println("\n可选择的文件类型:")
		fmt.Println("----------------------------------------")
		for i, typeName := range typeList {
			fmt.Printf("%2d. %s", i+1, typeName)
			// 显示扩展名
			if exts, ok := fileTypeExtensions[typeName]; ok && len(exts) > 0 {
				fmt.Printf(" (%s)", strings.Join(exts, ", "))
			}
			fmt.Println()
		}
		fmt.Println(" 0. 退出程序")
		fmt.Println("----------------------------------------")
		fmt.Print("请输入要选择的编号: ")

		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "0" {
			fmt.Println("程序已退出")
			os.Exit(0)
		}

		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > len(typeList) {
			fmt.Printf("无效的编号，请输入 1-%d 之间的数字\n", len(typeList))
			continue
		}

		selectedType := typeList[num-1]
		
		// 确认选择
		fmt.Printf("\n您选择了: %s\n", selectedType)
		if exts, ok := fileTypeExtensions[selectedType]; ok && len(exts) > 0 {
			fmt.Printf("将扫描扩展名为 %s 的文件\n", strings.Join(exts, ", "))
		}
		fmt.Print("确认选择？(y/n): ")
		
		scanner.Scan()
		confirm := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if confirm == "y" || confirm == "yes" {
			fmt.Println("===============================================================")
			fmt.Printf("已选择: %s\n", selectedType)
			fmt.Println("===============================================================")
			return selectedType
		}
		
		fmt.Println("重新选择...")
	}
}

func findFiles(directory string) [][2]string {
	var filesList [][2]string

	ignoredDirs := map[string]bool{
		"vendor":       true,
		"node_modules": true,
		"cache":        true,
		"temp":        true,
		".git":        true,
		".svn":        true,
		".idea":       true,
		".vscode":     true,
		"__pycache__": true,
		"target":      true,
		"build":       true,
		"dist":        true,
		"bin":         true,
		"obj":         true,
	}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if ignoredDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件大小 (1MB = 1024*1024 bytes)
		if info.Size() > 1024*1024 {
			return nil
		}

		// 获取文件扩展名
		ext := strings.ToLower(filepath.Ext(path))
		
		// 检查扩展名是否匹配选中的文件类型
		if exts, ok := fileTypeExtensions[selectedFileType]; ok {
			for _, allowedExt := range exts {
				if ext == allowedExt {
					filesList = append(filesList, [2]string{path, selectedFileType})
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("遍历目录时出错: %v\n", err)
	}

	return filesList
}

func safeReadFile(filePath string, maxChars int) (string, string) {
	// 检查文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Sprintf("无法获取文件信息: %v", err)
	}

	if fileInfo.Size() > int64(maxChars*3) { // 粗略估计，考虑编码因素
		return "", fmt.Sprintf("跳过 %s (文件可能超过%d字符限制)", filePath, maxChars)
	}

	// 尝试不同的编码
	encodings := []string{"utf-8", "gbk", "gb18030"}

	for _, encoding := range encodings {
		content, err := readFileWithEncoding(filePath, encoding, maxChars)
		if err == nil {
			if utf8.RuneCountInString(content) > maxChars {
				return "", fmt.Sprintf("跳过 %s (文件大小超过%d字符限制)", filePath, maxChars)
			}
			return content, ""
		}
	}

	// 最终尝试使用二进制读取
	content, err := readFileBinary(filePath, maxChars)
	if err != nil {
		return "", fmt.Sprintf("无法解码文件 %s: %v", filePath, err)
	}

	if utf8.RuneCountInString(content) > maxChars {
		return "", fmt.Sprintf("跳过 %s (文件大小超过%d字符限制)", filePath, maxChars)
	}

	return content, ""
}

func readFileWithEncoding(filePath, encoding string, maxChars int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var reader io.Reader = file

	// 处理中文编码
	if encoding == "gbk" || encoding == "gb18030" {
		reader = transform.NewReader(reader, simplifiedchinese.GBK.NewDecoder())
	}

	// 读取内容
	contentBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	content := string(contentBytes)

	// 检查字符数
	if utf8.RuneCountInString(content) > maxChars {
		// 截断内容
		runes := []rune(content)
		if len(runes) > maxChars {
			content = string(runes[:maxChars])
		}
	}

	return content, nil
}

func readFileBinary(filePath string, maxChars int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	contentBytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	content := string(contentBytes)

	// 检查字符数并截断
	if utf8.RuneCountInString(content) > maxChars {
		runes := []rune(content)
		if len(runes) > maxChars {
			content = string(runes[:maxChars])
		}
	}

	return content, nil
}

func postFileContent(url, filePath string) map[string]interface{} {
	content, errorMsg := safeReadFile(filePath, 60000)
	if errorMsg != "" {
		return map[string]interface{}{
			"content": errorMsg,
		}
	}

	// 准备请求数据
	data := map[string]interface{}{
		"file_content": content,
		"file_name":    filePath,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return map[string]interface{}{
			"content": fmt.Sprintf("JSON编码错误: %v", err),
		}
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return map[string]interface{}{
			"content": fmt.Sprintf("请求错误: %v", err),
		}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return map[string]interface{}{
			"content": fmt.Sprintf("响应解析错误: %v", err),
		}
	}

	return result
}

func main() {
	rand.Seed(time.Now().UnixNano())
	taskID := generateRandomString(8)
	
	fmt.Println("===============================================================")
	fmt.Println("       将程序放入代码根目录运行以扫描所有代码文件")
	fmt.Println("           联系客服咨询问题以及购买检查次数")
	fmt.Println("  https://work.weixin.qq.com/kfid/kfc7a6930ede9575277")
	fmt.Println("===============================================================")


	// 先选择文件类型
	selectedFileType = selectSingleFileType()

	// 简单的命令行参数解析
	startFrom := 0
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "--start-from" && i+1 < len(args) {
			_, err := fmt.Sscanf(args[i+1], "%d", &startFrom)
			if err != nil {
				fmt.Printf("解析参数错误: %v\n", err)
			}
			i++ // 跳过下一个参数
		}
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return
	}

	// 显示选择的文件类型
	fmt.Printf("\n正在搜索 %s 代码文件...\n", selectedFileType)
	if exts, ok := fileTypeExtensions[selectedFileType]; ok && len(exts) > 0 {
		fmt.Printf("将扫描扩展名为: %s\n", strings.Join(exts, ", "))
	}

	filesList := findFiles(currentDir)
	totalFiles := len(filesList)

	outPutInfo(fmt.Sprintf("找到 %d 个符合条件的代码文件", totalFiles))

	if totalFiles == 0 {
		fmt.Println("没有找到符合条件的文件。")
		fmt.Print("按任意键结束...")
		bufio.NewScanner(os.Stdin).Scan()
		return
	}

	fmt.Print("请输入授权码: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	authCode := strings.TrimSpace(scanner.Text())

	if authCode == "" {
		fmt.Println("授权码不能为空")
		return
	}
	fmt.Println("===============================================================")
	outPutInfo("查看本次代码检查报告访问: https://code.lamp.run/?id=" + taskID)
	fmt.Println("===============================================================")
	url := "https://user.lamp.run/cdk/manager/useCdkNum/" + authCode + "/" + fmt.Sprintf("%d", totalFiles)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("授权请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var authData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&authData); err != nil {
		fmt.Printf("解析授权响应失败: %v\n", err)
		return
	}

	if msg, ok := authData["msg"].(string); ok {
		fmt.Println(msg)
	}

	errVal, ok := authData["err"].(float64)
	if ok && errVal == 0 {
		// 创建进度条
		bar := progressbar.NewOptions(totalFiles-startFrom,
			progressbar.OptionSetDescription("检查代码中:"),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(30),
			progressbar.OptionClearOnFinish(),
		)

		// 处理文件
		for i := startFrom; i < totalFiles; i++ {
			if i >= len(filesList) {
				break
			}

			fileInfo := filesList[i]
			filePath := fileInfo[0]
			codeType := fileInfo[1]
			targetURL := "https://code.lamp.run/check/" + codeType + "/" + taskID + "/" + fmt.Sprintf("%d", totalFiles)

			postFileContent(targetURL, filePath)
			bar.Add(1)

			// 添加小延迟避免请求过快
			time.Sleep(100 * time.Millisecond)
		}

		// 打开浏览器
		browser := BrowserOpener{}
		browser.Open("https://code.lamp.run/?id="+taskID, "chrome", true)

		fmt.Print("扫描完成")
		fmt.Println("如果没有有自动打开浏览器，请手动访问: https://code.lamp.run/?id=" + taskID)
		fmt.Print("按任意键结束")
		scanner.Scan()
	} else {
		fmt.Print("授权失败,按任意键结束!")
		scanner.Scan()
	}
}