package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/faceair/clash-speedtest/speedtester"
	"github.com/metacubex/mihomo/log"
	"github.com/olekukonko/tablewriter"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v3"
)

var (
	configPathsConfig = flag.String("c", "", "config file path, also support http(s) url")
	filterRegexConfig = flag.String("f", ".+", "filter proxies by name, use regexp")
	serverURL         = flag.String("server-url", "https://speed.cloudflare.com", "server url")
	downloadSize      = flag.Int("download-size", 50*1024*1024, "download size for testing proxies")
	uploadSize        = flag.Int("upload-size", 20*1024*1024, "upload size for testing proxies")
	timeout           = flag.Duration("timeout", time.Second*5, "timeout for testing proxies")
	concurrent        = flag.Int("concurrent", 4, "download concurrent size")
	outputPath        = flag.String("output", "", "output config file path")
	stashCompatible   = flag.Bool("stash-compatible", false, "enable stash compatible mode")
	maxLatency        = flag.Duration("max-latency", 800*time.Millisecond, "filter latency greater than this value")
	minDownloadSpeed  = flag.Float64("min-download-speed", 5, "filter download speed less than this value(unit: MB/s)")
	minUploadSpeed    = flag.Float64("min-upload-speed", 2, "filter upload speed less than this value(unit: MB/s)")
	renameNodes       = flag.Bool("rename", false, "rename nodes with IP location and speed")
	interactive       = flag.Bool("interactive", false, "enable interactive debug mode")
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorReset  = "\033[0m"
)

// 全局变量用于交互式模式
var (
	globalSpeedTester *speedtester.SpeedTester
	globalProxies     map[string]*speedtester.CProxy
	globalResults     []*speedtester.Result
)

func main() {
	flag.Parse()
	log.SetLevel(log.SILENT)

	if *interactive {
		runInteractiveMode()
		return
	}

	// 原有的非交互式模式
	if *configPathsConfig == "" {
		log.Fatalln("please specify the configuration file")
	}

	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths:      *configPathsConfig,
		FilterRegex:      *filterRegexConfig,
		ServerURL:        *serverURL,
		DownloadSize:     *downloadSize,
		UploadSize:       *uploadSize,
		Timeout:          *timeout,
		Concurrent:       *concurrent,
		MaxLatency:       *maxLatency,
		MinDownloadSpeed: *minDownloadSpeed * 1024 * 1024,
		MinUploadSpeed:   *minUploadSpeed * 1024 * 1024,
	})

	allProxies, err := speedTester.LoadProxies(*stashCompatible)
	if err != nil {
		log.Fatalln("load proxies failed: %v", err)
	}

	bar := progressbar.Default(int64(len(allProxies)), "测试中...")
	results := make([]*speedtester.Result, 0)
	speedTester.TestProxies(allProxies, func(result *speedtester.Result) {
		bar.Add(1)
		bar.Describe(result.ProxyName)
		results = append(results, result)
	})

	sort.Slice(results, func(i, j int) bool {
		return results[i].DownloadSpeed > results[j].DownloadSpeed
	})

	printResults(results)

	if *outputPath != "" {
		err = saveConfig(results)
		if err != nil {
			log.Fatalln("save config file failed: %v", err)
		}
		fmt.Printf("\nsave config file to: %s\n", *outputPath)
	}
}

func runInteractiveMode() {
	printWelcome()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		printMenu()
		fmt.Print(colorCyan + "请选择操作: " + colorReset)
		
		if !scanner.Scan() {
			break
		}
		
		choice := strings.TrimSpace(scanner.Text())
		
		switch choice {
		case "1":
			handleLoadConfig(scanner)
		case "2":
			handleLoadProxies(scanner)
		case "3":
			handleViewProxies()
		case "4":
			handleTestSingleProxy(scanner)
		case "5":
			handleTestAllProxies(scanner)
		case "6":
			handleViewResults()
		case "7":
			handleFilterResults(scanner)
		case "8":
			handleSaveResults(scanner)
		case "9":
			handleSettings(scanner)
		case "10":
			handleDebugConfigFile(scanner)
		case "0":
			fmt.Println(colorGreen + "再见！" + colorReset)
			return
		default:
			fmt.Println(colorRed + "无效选择，请重试" + colorReset)
		}
		
		fmt.Print("\n按 Enter 继续...")
		scanner.Scan()
	}
}

func printWelcome() {
	fmt.Println(colorBlue + "=" + strings.Repeat("=", 60) + colorReset)
	fmt.Println(colorBlue + "           Clash Speed Test 交互式调试工具" + colorReset)
	fmt.Println(colorBlue + "=" + strings.Repeat("=", 60) + colorReset)
	fmt.Println()
}

func printMenu() {
	fmt.Println(colorYellow + "\n========== 主菜单 ==========" + colorReset)
	fmt.Println("1. 设置配置文件路径")
	fmt.Println("2. 加载代理列表")
	fmt.Println("3. 查看代理列表")
	fmt.Println("4. 测试单个代理")
	fmt.Println("5. 测试所有代理")
	fmt.Println("6. 查看测试结果")
	fmt.Println("7. 过滤测试结果")
	fmt.Println("8. 保存配置文件")
	fmt.Println("9. 调试设置")
	fmt.Println("10. 调试配置文件")
	fmt.Println("0. 退出")
	fmt.Println(colorYellow + "===========================" + colorReset)
}

func handleLoadConfig(scanner *bufio.Scanner) {
	fmt.Print("请输入配置文件路径(支持本地文件和HTTP URL): ")
	if !scanner.Scan() {
		return
	}
	path := strings.TrimSpace(scanner.Text())
	if path == "" {
		fmt.Println(colorRed + "配置文件路径不能为空" + colorReset)
		return
	}

	// 移除可能的引号
	if (strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"")) ||
		(strings.HasPrefix(path, "'") && strings.HasSuffix(path, "'")) {
		path = path[1 : len(path)-1]
	}

	// 验证路径
	if !strings.HasPrefix(path, "http") {
		// 本地文件路径验证
		if _, err := os.Stat(path); err != nil {
			fmt.Printf(colorRed+"本地文件路径无效: %v\n"+colorReset, err)
			fmt.Println(colorYellow + "提示:" + colorReset)
			fmt.Println("- 确保文件路径正确")
			fmt.Println("- 如果路径包含空格，可以用引号包围")
			fmt.Println("- 例如: \"/path/with spaces/config.yaml\"")
			return
		}
	}

	*configPathsConfig = path
	fmt.Println(colorGreen + "配置文件路径已设置: " + path + colorReset)
}

func handleLoadProxies(scanner *bufio.Scanner) {
	if *configPathsConfig == "" {
		fmt.Println(colorRed + "请先设置配置文件路径" + colorReset)
		return
	}

	// 创建SpeedTester实例
	globalSpeedTester = speedtester.New(&speedtester.Config{
		ConfigPaths:      *configPathsConfig,
		FilterRegex:      *filterRegexConfig,
		ServerURL:        *serverURL,
		DownloadSize:     *downloadSize,
		UploadSize:       *uploadSize,
		Timeout:          *timeout,
		Concurrent:       *concurrent,
		MaxLatency:       *maxLatency,
		MinDownloadSpeed: *minDownloadSpeed * 1024 * 1024,
		MinUploadSpeed:   *minUploadSpeed * 1024 * 1024,
	})

	fmt.Println("正在加载代理列表...")
	fmt.Printf(colorBlue+"配置路径: %s\n"+colorReset, *configPathsConfig)
	fmt.Printf(colorBlue+"过滤正则: %s\n"+colorReset, *filterRegexConfig)
	
	// 添加调试信息
	fmt.Println(colorYellow + "开始下载配置文件..." + colorReset)
	
	proxies, err := globalSpeedTester.LoadProxies(*stashCompatible)
	if err != nil {
		fmt.Printf(colorRed+"加载代理失败: %v\n"+colorReset, err)
		return
	}

	globalProxies = proxies
	fmt.Printf(colorGreen+"成功加载 %d 个代理\n"+colorReset, len(proxies))
	
	// 如果加载的代理数为0，提供调试选项
	if len(proxies) == 0 {
		fmt.Println(colorYellow + "代理数量为0，是否需要查看调试信息？(y/N): " + colorReset)
		if scanner.Scan() && strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
			showDebugInfo(*configPathsConfig)
		}
	}
}

func handleViewProxies() {
	if globalProxies == nil {
		fmt.Println(colorRed + "请先加载代理列表" + colorReset)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"序号", "代理名称", "类型", "服务器", "端口"})

	i := 1
	for name, proxy := range globalProxies {
		server := "N/A"
		port := "N/A"
		if s, ok := proxy.Config["server"]; ok {
			server = fmt.Sprintf("%v", s)
		}
		if p, ok := proxy.Config["port"]; ok {
			port = fmt.Sprintf("%v", p)
		}

		table.Append([]string{
			fmt.Sprintf("%d", i),
			name,
			proxy.Type().String(),
			server,
			port,
		})
		i++
	}

	fmt.Printf(colorBlue+"代理列表 (共 %d 个):\n"+colorReset, len(globalProxies))
	table.Render()
}

func handleTestSingleProxy(scanner *bufio.Scanner) {
	if globalProxies == nil {
		fmt.Println(colorRed + "请先加载代理列表" + colorReset)
		return
	}

	fmt.Print("请输入代理名称或序号: ")
	if !scanner.Scan() {
		return
	}
	
	input := strings.TrimSpace(scanner.Text())
	var targetProxy *speedtester.CProxy
	var targetName string

	// 尝试按序号查找
	if num, err := strconv.Atoi(input); err == nil {
		i := 1
		for name, proxy := range globalProxies {
			if i == num {
				targetProxy = proxy
				targetName = name
				break
			}
			i++
		}
	} else {
		// 按名称查找
		for name, proxy := range globalProxies {
			if strings.Contains(strings.ToLower(name), strings.ToLower(input)) {
				targetProxy = proxy
				targetName = name
				break
			}
		}
	}

	if targetProxy == nil {
		fmt.Println(colorRed + "未找到匹配的代理" + colorReset)
		return
	}

	fmt.Printf("正在测试代理: %s\n", targetName)
	
	// 创建一个临时的代理映射进行测试
	tempProxies := map[string]*speedtester.CProxy{targetName: targetProxy}
	
	globalSpeedTester.TestProxies(tempProxies, func(result *speedtester.Result) {
		fmt.Printf(colorGreen+"测试完成!\n"+colorReset)
		printSingleResult(result)
		
		// 添加到全局结果中
		if globalResults == nil {
			globalResults = make([]*speedtester.Result, 0)
		}
		// 检查是否已存在，如果存在则更新
		found := false
		for i, r := range globalResults {
			if r.ProxyName == result.ProxyName {
				globalResults[i] = result
				found = true
				break
			}
		}
		if !found {
			globalResults = append(globalResults, result)
		}
	})
}

func handleTestAllProxies(scanner *bufio.Scanner) {
	if globalProxies == nil {
		fmt.Println(colorRed + "请先加载代理列表" + colorReset)
		return
	}

	fmt.Printf("将测试 %d 个代理，确认吗？(y/N): ", len(globalProxies))
	if !scanner.Scan() {
		return
	}
	
	if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
		fmt.Println("已取消测试")
		return
	}

	globalResults = make([]*speedtester.Result, 0)
	bar := progressbar.Default(int64(len(globalProxies)), "测试中...")
	
	globalSpeedTester.TestProxies(globalProxies, func(result *speedtester.Result) {
		bar.Add(1)
		bar.Describe(result.ProxyName)
		globalResults = append(globalResults, result)
	})

	// 按下载速度排序
	sort.Slice(globalResults, func(i, j int) bool {
		return globalResults[i].DownloadSpeed > globalResults[j].DownloadSpeed
	})

	fmt.Printf(colorGreen+"\n所有代理测试完成！共测试 %d 个代理\n"+colorReset, len(globalResults))
}

func handleViewResults() {
	if globalResults == nil || len(globalResults) == 0 {
		fmt.Println(colorRed + "暂无测试结果" + colorReset)
		return
	}

	printResults(globalResults)
}

func handleFilterResults(scanner *bufio.Scanner) {
	if globalResults == nil || len(globalResults) == 0 {
		fmt.Println(colorRed + "暂无测试结果" + colorReset)
		return
	}

	fmt.Println(colorYellow + "过滤选项:" + colorReset)
	fmt.Println("1. 按延迟过滤")
	fmt.Println("2. 按下载速度过滤")
	fmt.Println("3. 按上传速度过滤")
	fmt.Println("4. 按丢包率过滤")
	fmt.Print("请选择过滤类型: ")

	if !scanner.Scan() {
		return
	}

	choice := strings.TrimSpace(scanner.Text())
	filtered := make([]*speedtester.Result, 0)

	switch choice {
	case "1":
		fmt.Print("请输入最大延迟(ms): ")
		if !scanner.Scan() {
			return
		}
		if maxMs, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil {
			maxLatency := time.Duration(maxMs) * time.Millisecond
			for _, r := range globalResults {
				if r.Latency <= maxLatency && r.Latency > 0 {
					filtered = append(filtered, r)
				}
			}
		}
	case "2":
		fmt.Print("请输入最小下载速度(MB/s): ")
		if !scanner.Scan() {
			return
		}
		if minSpeed, err := strconv.ParseFloat(strings.TrimSpace(scanner.Text()), 64); err == nil {
			minSpeedBytes := minSpeed * 1024 * 1024
			for _, r := range globalResults {
				if r.DownloadSpeed >= minSpeedBytes {
					filtered = append(filtered, r)
				}
			}
		}
	case "3":
		fmt.Print("请输入最小上传速度(MB/s): ")
		if !scanner.Scan() {
			return
		}
		if minSpeed, err := strconv.ParseFloat(strings.TrimSpace(scanner.Text()), 64); err == nil {
			minSpeedBytes := minSpeed * 1024 * 1024
			for _, r := range globalResults {
				if r.UploadSpeed >= minSpeedBytes {
					filtered = append(filtered, r)
				}
			}
		}
	case "4":
		fmt.Print("请输入最大丢包率(%): ")
		if !scanner.Scan() {
			return
		}
		if maxLoss, err := strconv.ParseFloat(strings.TrimSpace(scanner.Text()), 64); err == nil {
			for _, r := range globalResults {
				if r.PacketLoss <= maxLoss {
					filtered = append(filtered, r)
				}
			}
		}
	default:
		fmt.Println(colorRed + "无效选择" + colorReset)
		return
	}

	if len(filtered) == 0 {
		fmt.Println(colorRed + "没有符合条件的结果" + colorReset)
		return
	}

	fmt.Printf(colorGreen+"过滤后的结果 (共 %d 个):\n"+colorReset, len(filtered))
	printResults(filtered)
}

func handleSaveResults(scanner *bufio.Scanner) {
	if globalResults == nil || len(globalResults) == 0 {
		fmt.Println(colorRed + "暂无测试结果" + colorReset)
		return
	}

	fmt.Print("请输入保存路径: ")
	if !scanner.Scan() {
		return
	}
	
	path := strings.TrimSpace(scanner.Text())
	if path == "" {
		fmt.Println(colorRed + "路径不能为空" + colorReset)
		return
	}

	*outputPath = path
	err := saveConfig(globalResults)
	if err != nil {
		fmt.Printf(colorRed+"保存失败: %v\n"+colorReset, err)
		return
	}

	fmt.Printf(colorGreen+"配置文件已保存到: %s\n"+colorReset, path)
}

func handleSettings(scanner *bufio.Scanner) {
	for {
		fmt.Println(colorYellow + "\n========== 调试设置 ==========" + colorReset)
		fmt.Printf("1. 服务器URL: %s\n", *serverURL)
		fmt.Printf("2. 下载大小: %d MB\n", *downloadSize/(1024*1024))
		fmt.Printf("3. 上传大小: %d MB\n", *uploadSize/(1024*1024))
		fmt.Printf("4. 超时时间: %v\n", *timeout)
		fmt.Printf("5. 并发数: %d\n", *concurrent)
		fmt.Printf("6. 最大延迟: %v\n", *maxLatency)
		fmt.Printf("7. 最小下载速度: %.3f MB/s (%.0f KB/s)\n", *minDownloadSpeed, *minDownloadSpeed*1024)
		fmt.Printf("8. 最小上传速度: %.3f MB/s (%.0f KB/s)\n", *minUploadSpeed, *minUploadSpeed*1024)
		fmt.Printf("9. 过滤正则: %s\n", *filterRegexConfig)
		fmt.Println("0. 返回主菜单")
		fmt.Println(colorYellow + "==============================" + colorReset)
		
		fmt.Print("请选择要修改的设置: ")
		if !scanner.Scan() {
			return
		}
		
		choice := strings.TrimSpace(scanner.Text())
		
		switch choice {
		case "1":
			fmt.Print("请输入新的服务器URL: ")
			if scanner.Scan() {
				*serverURL = strings.TrimSpace(scanner.Text())
				fmt.Println(colorGreen + "服务器URL已更新" + colorReset)
			}
		case "2":
			fmt.Print("请输入下载大小(MB): ")
			if scanner.Scan() {
				if size, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil {
					*downloadSize = size * 1024 * 1024
					fmt.Println(colorGreen + "下载大小已更新" + colorReset)
				}
			}
		case "3":
			fmt.Print("请输入上传大小(MB): ")
			if scanner.Scan() {
				if size, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil {
					*uploadSize = size * 1024 * 1024
					fmt.Println(colorGreen + "上传大小已更新" + colorReset)
				}
			}
		case "4":
			fmt.Print("请输入超时时间(秒): ")
			if scanner.Scan() {
				if t, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil {
					*timeout = time.Duration(t) * time.Second
					fmt.Println(colorGreen + "超时时间已更新" + colorReset)
				}
			}
		case "5":
			fmt.Print("请输入并发数: ")
			if scanner.Scan() {
				if c, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil {
					*concurrent = c
					fmt.Println(colorGreen + "并发数已更新" + colorReset)
				}
			}
		case "6":
			fmt.Print("请输入最大延迟(毫秒): ")
			if scanner.Scan() {
				if ms, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil {
					*maxLatency = time.Duration(ms) * time.Millisecond
					fmt.Println(colorGreen + "最大延迟已更新" + colorReset)
				}
			}
		case "7":
			fmt.Printf("当前最小下载速度: %.3f MB/s (%.0f KB/s)\n", *minDownloadSpeed, *minDownloadSpeed*1024)
			fmt.Print("请输入新的最小下载速度(KB/s): ")
			if scanner.Scan() {
				if speedKB, err := strconv.ParseFloat(strings.TrimSpace(scanner.Text()), 64); err == nil {
					*minDownloadSpeed = speedKB / 1024  // 转换为MB/s
					fmt.Printf(colorGreen+"最小下载速度已更新为: %.3f MB/s (%.0f KB/s)\n"+colorReset, 
						*minDownloadSpeed, *minDownloadSpeed*1024)
				}
			}
		case "8":
			fmt.Printf("当前最小上传速度: %.3f MB/s (%.0f KB/s)\n", *minUploadSpeed, *minUploadSpeed*1024)
			fmt.Print("请输入新的最小上传速度(KB/s): ")
			if scanner.Scan() {
				if speedKB, err := strconv.ParseFloat(strings.TrimSpace(scanner.Text()), 64); err == nil {
					*minUploadSpeed = speedKB / 1024  // 转换为MB/s
					fmt.Printf(colorGreen+"最小上传速度已更新为: %.3f MB/s (%.0f KB/s)\n"+colorReset, 
						*minUploadSpeed, *minUploadSpeed*1024)
				}
			}
		case "9":
			fmt.Print("请输入过滤正则表达式: ")
			if scanner.Scan() {
				*filterRegexConfig = strings.TrimSpace(scanner.Text())
				fmt.Println(colorGreen + "过滤正则已更新" + colorReset)
			}
		case "0":
			return
		default:
			fmt.Println(colorRed + "无效选择" + colorReset)
		}
	}
}

func printSingleResult(result *speedtester.Result) {
	fmt.Printf("代理名称: %s\n", result.ProxyName)
	fmt.Printf("代理类型: %s\n", result.ProxyType)
	fmt.Printf("延迟: %s\n", result.FormatLatency())
	fmt.Printf("抖动: %s\n", result.FormatJitter())
	fmt.Printf("丢包率: %s\n", result.FormatPacketLoss())
	fmt.Printf("下载速度: %s\n", result.FormatDownloadSpeed())
	fmt.Printf("上传速度: %s\n", result.FormatUploadSpeed())
}

func printResults(results []*speedtester.Result) {
	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{
		"序号",
		"节点名称",
		"类型",
		"延迟",
		"抖动",
		"丢包率",
		"下载速度",
		"上传速度",
	})

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	for i, result := range results {
		idStr := fmt.Sprintf("%d.", i+1)

		// 延迟颜色
		latencyStr := result.FormatLatency()
		if result.Latency > 0 {
			if result.Latency < 800*time.Millisecond {
				latencyStr = colorGreen + latencyStr + colorReset
			} else if result.Latency < 1500*time.Millisecond {
				latencyStr = colorYellow + latencyStr + colorReset
			} else {
				latencyStr = colorRed + latencyStr + colorReset
			}
		} else {
			latencyStr = colorRed + latencyStr + colorReset
		}

		jitterStr := result.FormatJitter()
		if result.Jitter > 0 {
			if result.Jitter < 800*time.Millisecond {
				jitterStr = colorGreen + jitterStr + colorReset
			} else if result.Jitter < 1500*time.Millisecond {
				jitterStr = colorYellow + jitterStr + colorReset
			} else {
				jitterStr = colorRed + jitterStr + colorReset
			}
		} else {
			jitterStr = colorRed + jitterStr + colorReset
		}

		// 丢包率颜色
		packetLossStr := result.FormatPacketLoss()
		if result.PacketLoss < 10 {
			packetLossStr = colorGreen + packetLossStr + colorReset
		} else if result.PacketLoss < 20 {
			packetLossStr = colorYellow + packetLossStr + colorReset
		} else {
			packetLossStr = colorRed + packetLossStr + colorReset
		}

		// 下载速度颜色 (以MB/s为单位判断)
		downloadSpeed := result.DownloadSpeed / (1024 * 1024)
		downloadSpeedStr := result.FormatDownloadSpeed()
		if downloadSpeed >= 10 {
			downloadSpeedStr = colorGreen + downloadSpeedStr + colorReset
		} else if downloadSpeed >= 5 {
			downloadSpeedStr = colorYellow + downloadSpeedStr + colorReset
		} else {
			downloadSpeedStr = colorRed + downloadSpeedStr + colorReset
		}

		// 上传速度颜色
		uploadSpeed := result.UploadSpeed / (1024 * 1024)
		uploadSpeedStr := result.FormatUploadSpeed()
		if uploadSpeed >= 5 {
			uploadSpeedStr = colorGreen + uploadSpeedStr + colorReset
		} else if uploadSpeed >= 2 {
			uploadSpeedStr = colorYellow + uploadSpeedStr + colorReset
		} else {
			uploadSpeedStr = colorRed + uploadSpeedStr + colorReset
		}

		row := []string{
			idStr,
			result.ProxyName,
			result.ProxyType,
			latencyStr,
			jitterStr,
			packetLossStr,
			downloadSpeedStr,
			uploadSpeedStr,
		}

		table.Append(row)
	}

	fmt.Println()
	table.Render()
	fmt.Println()
}

func saveConfig(results []*speedtester.Result) error {
	proxies := make([]map[string]any, 0)
	for _, result := range results {
		if *maxLatency > 0 && result.Latency > *maxLatency {
			continue
		}
		if *downloadSize > 0 && *minDownloadSpeed > 0 && result.DownloadSpeed < *minDownloadSpeed*1024*1024 {
			continue
		}
		if *uploadSize > 0 && *minUploadSpeed > 0 && result.UploadSpeed < *minUploadSpeed*1024*1024 {
			continue
		}

		proxyConfig := result.ProxyConfig
		if *renameNodes {
			location, err := getIPLocation(proxyConfig["server"].(string))
			if err != nil || location.CountryCode == "" {
				proxies = append(proxies, proxyConfig)
				continue
			}
			proxyConfig["name"] = generateNodeName(location.CountryCode, result.DownloadSpeed)
		}
		proxies = append(proxies, proxyConfig)
	}

	config := &speedtester.RawConfig{
		Proxies: proxies,
	}
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(*outputPath, yamlData, 0o644)
}

type IPLocation struct {
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
}

var countryFlags = map[string]string{
	"US": "🇺🇸", "CN": "🇨🇳", "GB": "🇬🇧", "UK": "🇬🇧", "JP": "🇯🇵", "DE": "🇩🇪", "FR": "🇫🇷", "RU": "🇷🇺",
	"SG": "🇸🇬", "HK": "🇭🇰", "TW": "🇹🇼", "KR": "🇰🇷", "CA": "🇨🇦", "AU": "🇦🇺", "NL": "🇳🇱", "IT": "🇮🇹",
	"ES": "🇪🇸", "SE": "🇸🇪", "NO": "🇳🇴", "DK": "🇩🇰", "FI": "🇫🇮", "CH": "🇨🇭", "AT": "🇦🇹", "BE": "🇧🇪",
	"BR": "🇧🇷", "IN": "🇮🇳", "TH": "🇹🇭", "MY": "🇲🇾", "VN": "🇻🇳", "PH": "🇵🇭", "ID": "🇮🇩", "UA": "🇺🇦",
	"TR": "🇹🇷", "IL": "🇮🇱", "AE": "🇦🇪", "SA": "🇸🇦", "EG": "🇪🇬", "ZA": "🇿🇦", "NG": "🇳🇬", "KE": "🇰🇪",
	"RO": "🇷🇴", "PL": "🇵🇱", "CZ": "🇨🇿", "HU": "🇭🇺", "BG": "🇧🇬", "HR": "🇭🇷", "SI": "🇸🇮", "SK": "🇸🇰",
	"LT": "🇱🇹", "LV": "🇱🇻", "EE": "🇪🇪", "PT": "🇵🇹", "GR": "🇬🇷", "IE": "🇮🇪", "LU": "🇱🇺", "MT": "🇲🇹",
	"CY": "🇨🇾", "IS": "🇮🇸", "MX": "🇲🇽", "AR": "🇦🇷", "CL": "🇨🇱", "CO": "🇨🇴", "PE": "🇵🇪", "VE": "🇻🇪",
	"EC": "🇪🇨", "UY": "🇺🇾", "PY": "🇵🇾", "BO": "🇧🇴", "CR": "🇨🇷", "PA": "🇵🇦", "GT": "🇬🇹", "HN": "🇭🇳",
	"SV": "🇸🇻", "NI": "🇳🇮", "BZ": "🇧🇿", "JM": "🇯🇲", "TT": "🇹🇹", "BB": "🇧🇧", "GD": "🇬🇩", "LC": "🇱🇨",
	"VC": "🇻🇨", "AG": "🇦🇬", "DM": "🇩🇲", "KN": "🇰🇳", "BS": "🇧🇸", "CU": "🇨🇺", "DO": "🇩🇴", "HT": "🇭🇹",
	"PR": "🇵🇷", "VI": "🇻🇮", "GU": "🇬🇺", "AS": "🇦🇸", "MP": "🇲🇵", "PW": "🇵🇼", "FM": "🇫🇲", "MH": "🇲🇭",
	"KI": "🇰🇮", "TV": "🇹🇻", "NR": "🇳🇷", "WS": "🇼🇸", "TO": "🇹🇴", "FJ": "🇫🇯", "VU": "🇻🇺", "SB": "🇸🇧",
	"PG": "🇵🇬", "NC": "🇳🇨", "PF": "🇵🇫", "WF": "🇼🇫", "CK": "🇨🇰", "NU": "🇳🇺", "TK": "🇹🇰", "SC": "🇸🇨",
}

func getIPLocation(ip string) (*IPLocation, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=country,countryCode", ip))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get location for IP %s", ip)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var location IPLocation
	if err := json.Unmarshal(body, &location); err != nil {
		return nil, err
	}
	return &location, nil
}

func generateNodeName(countryCode string, downloadSpeed float64) string {
	flag, exists := countryFlags[strings.ToUpper(countryCode)]
	if !exists {
		flag = "🏳️"
	}

	speedMBps := downloadSpeed / (1024 * 1024)
	return fmt.Sprintf("%s %s | ⬇️ %.2f MB/s", flag, strings.ToUpper(countryCode), speedMBps)
}

func showDebugInfo(configPath string) {
	// 处理路径
	configPath = strings.TrimSpace(configPath)
	if (strings.HasPrefix(configPath, "\"") && strings.HasSuffix(configPath, "\"")) ||
		(strings.HasPrefix(configPath, "'") && strings.HasSuffix(configPath, "'")) {
		configPath = configPath[1 : len(configPath)-1]
	}

	fmt.Println(colorYellow + "========== 调试信息 ==========" + colorReset)
	fmt.Printf(colorBlue+"正在调试配置路径: %s\n"+colorReset, configPath)
	
	// 1. 下载配置文件内容
	fmt.Println("1. 下载配置文件...")
	var body []byte
	var err error
	
	if strings.HasPrefix(configPath, "http") {
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(configPath)
		if err != nil {
			fmt.Printf(colorRed+"下载失败: %v\n"+colorReset, err)
			return
		}
		defer resp.Body.Close()
		
		fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)
		fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
		fmt.Printf("Content-Length: %s\n", resp.Header.Get("Content-Length"))
		
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf(colorRed+"读取响应失败: %v\n"+colorReset, err)
			return
		}
	} else {
		body, err = os.ReadFile(configPath)
		if err != nil {
			fmt.Printf(colorRed+"读取文件失败: %v\n"+colorReset, err)
			return
		}
	}
	
	fmt.Printf("配置文件大小: %d bytes\n", len(body))
	
	// 2. 显示配置文件前500字符
	fmt.Println("\n2. 配置文件内容预览:")
	contentPreview := string(body)
	if len(contentPreview) > 500 {
		contentPreview = contentPreview[:500] + "..."
	}
	fmt.Printf(colorCyan+"%s\n"+colorReset, contentPreview)
	
	// 3. 尝试解析YAML
	fmt.Println("\n3. 解析YAML配置:")
	rawCfg := &speedtester.RawConfig{
		Proxies: []map[string]any{},
	}
	
	if err := yaml.Unmarshal(body, rawCfg); err != nil {
		fmt.Printf(colorRed+"YAML解析失败: %v\n"+colorReset, err)
		
		// 尝试查看是否是JSON格式
		fmt.Println("\n尝试解析为JSON:")
		var jsonData map[string]any
		if jsonErr := json.Unmarshal(body, &jsonData); jsonErr == nil {
			fmt.Println(colorGreen+"内容似乎是JSON格式"+colorReset)
			if proxies, ok := jsonData["proxies"]; ok {
				fmt.Printf("找到 proxies 字段，类型: %T\n", proxies)
			}
		} else {
			fmt.Printf(colorRed+"JSON解析也失败: %v\n"+colorReset, jsonErr)
		}
		return
	}
	
	// 4. 显示解析结果
	fmt.Printf(colorGreen+"YAML解析成功\n"+colorReset)
	fmt.Printf("原始代理数量: %d\n", len(rawCfg.Proxies))
	fmt.Printf("代理提供者数量: %d\n", len(rawCfg.Providers))
	
	// 5. 显示前几个代理的信息
	if len(rawCfg.Proxies) > 0 {
		fmt.Println("\n5. 前3个代理信息:")
		for i, proxy := range rawCfg.Proxies {
			if i >= 3 {
				break
			}
			fmt.Printf("代理 %d:\n", i+1)
			fmt.Printf("  名称: %v\n", proxy["name"])
			fmt.Printf("  类型: %v\n", proxy["type"])
			fmt.Printf("  服务器: %v\n", proxy["server"])
			fmt.Printf("  端口: %v\n", proxy["port"])
			fmt.Println()
		}
	}
	
	// 6. 检查过滤正则
	fmt.Println("6. 检查过滤正则:")
	fmt.Printf("当前过滤正则: %s\n", *filterRegexConfig)
	
	if len(rawCfg.Proxies) > 0 {
		filterRegexp := regexp.MustCompile(*filterRegexConfig)
		matchedCount := 0
		
		for _, proxy := range rawCfg.Proxies {
			if name, ok := proxy["name"].(string); ok {
				if filterRegexp.MatchString(name) {
					matchedCount++
				}
			}
		}
		
		fmt.Printf("通过正则过滤的代理数量: %d\n", matchedCount)
		
		if matchedCount == 0 {
			fmt.Println(colorRed+"没有代理通过正则过滤！"+colorReset)
			fmt.Println("建议将过滤正则改为 \".+\" 来包含所有代理")
		}
	}
	
	// 7. 检查Stash兼容性
	if *stashCompatible {
		fmt.Println("\n7. Stash兼容性检查:")
		fmt.Println("当前启用了Stash兼容模式，某些代理类型可能被过滤")
		fmt.Println("建议暂时关闭Stash兼容模式进行测试")
	}
	
	fmt.Println(colorYellow + "=============================" + colorReset)
}

func handleDebugConfigFile(scanner *bufio.Scanner) {
	fmt.Print("请输入配置文件路径(支持本地文件和HTTP URL): ")
	if !scanner.Scan() {
		return
	}
	path := strings.TrimSpace(scanner.Text())
	if path == "" {
		fmt.Println(colorRed + "配置文件路径不能为空" + colorReset)
		return
	}

	// 移除可能的引号
	if (strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"")) ||
		(strings.HasPrefix(path, "'") && strings.HasSuffix(path, "'")) {
		path = path[1 : len(path)-1]
	}

	// 验证路径
	if !strings.HasPrefix(path, "http") {
		// 本地文件路径验证
		if _, err := os.Stat(path); err != nil {
			fmt.Printf(colorRed+"本地文件路径无效: %v\n"+colorReset, err)
			fmt.Println(colorYellow + "提示:" + colorReset)
			fmt.Println("- 确保文件路径正确")
			fmt.Println("- 如果路径包含空格，可以用引号包围")
			fmt.Println("- 例如: \"/path/with spaces/config.yaml\"")
			return
		}
	}

	showDebugInfo(path)
}
