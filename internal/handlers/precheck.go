package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	// MaxConcurrency SSH 并发数
	MaxConcurrency = 64
)

// NodeStatus 节点状态信息
type NodeStatus struct {
	IP           string `json:"ip"`
	ProcessCount int    `json:"process_count"`
	Error        string `json:"error,omitempty"`
}

// PrecheckResponse precheck 接口响应
type PrecheckResponse struct {
	TotalNodes int          `json:"total_nodes"`
	BusyNodes  []NodeStatus `json:"busy_nodes"`
	BusyCount  int          `json:"busy_count"`
	ErrorNodes []NodeStatus `json:"error_nodes,omitempty"`
	ErrorCount int          `json:"error_count"`
}

// Precheck 检查所有节点的 GPU 进程状态
func Precheck(c *gin.Context) {
	// 获取 iplist 文件参数，filename 必选
	iplistFile := c.Query("filename")
	if iplistFile == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "filename parameter is required",
		})
		return
	}

	// 验证路径安全性，防止路径遍历攻击
	absIPListDir := filepath.Join(DataDir, "iplist")
	absPath := filepath.Join(absIPListDir, iplistFile)
	if !strings.HasPrefix(absPath, absIPListDir) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}

	// 读取 IP 列表
	filePath := filepath.Join(DataDir, "iplist", iplistFile)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusOK, PrecheckResponse{
			TotalNodes: 0,
			BusyNodes:  []NodeStatus{},
			BusyCount:  0,
			ErrorCount: 0,
		})
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read IP list",
		})
		return
	}

	// 解析 IP 列表
	content := strings.TrimSpace(string(data))
	if content == "" {
		c.JSON(http.StatusOK, PrecheckResponse{
			TotalNodes: 0,
			BusyNodes:  []NodeStatus{},
			BusyCount:  0,
			ErrorCount: 0,
		})
		return
	}

	ipList := strings.Split(content, "\n")
	var validIPs []string
	for _, ip := range ipList {
		if trimmed := strings.TrimSpace(ip); trimmed != "" {
			validIPs = append(validIPs, trimmed)
		}
	}

	if len(validIPs) == 0 {
		c.JSON(http.StatusOK, PrecheckResponse{
			TotalNodes: 0,
			BusyNodes:  []NodeStatus{},
			BusyCount:  0,
			ErrorCount: 0,
		})
		return
	}

	// 并行检查所有节点
	results := checkNodesParallel(validIPs, MaxConcurrency)

	// 收集繁忙节点和错误节点
	var busyNodes []NodeStatus
	var errorNodes []NodeStatus

	for _, result := range results {
		if result.Error != "" {
			errorNodes = append(errorNodes, result)
		} else if result.ProcessCount > 0 {
			busyNodes = append(busyNodes, result)
		}
	}

	response := PrecheckResponse{
		TotalNodes: len(validIPs),
		BusyNodes:  busyNodes,
		BusyCount:  len(busyNodes),
		ErrorNodes: errorNodes,
		ErrorCount: len(errorNodes),
	}

	c.JSON(http.StatusOK, response)
}

// checkNodesParallel 并行检查多个节点
func checkNodesParallel(ips []string, concurrency int) []NodeStatus {
	results := make([]NodeStatus, len(ips))
	var wg sync.WaitGroup

	// 创建信号量来限制并发数
	semaphore := make(chan struct{}, concurrency)

	for i, ip := range ips {
		wg.Add(1)
		go func(index int, nodeIP string) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 检查单个节点
			results[index] = checkSingleNode(nodeIP)
		}(i, ip)
	}

	wg.Wait()
	return results
}

// checkSingleNode 检查单个节点的 GPU 进程数
func checkSingleNode(ip string) NodeStatus {
	// 构建 SSH 命令
	cmd := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "ConnectTimeout=5",
		ip,
		"nvidia-smi --query-compute-apps=pid --format=csv,noheader | wc -l",
	)

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return NodeStatus{
			IP:           ip,
			ProcessCount: 0,
			Error:        fmt.Sprintf("SSH failed: %v", err),
		}
	}

	// 解析进程数
	countStr := strings.TrimSpace(string(output))
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return NodeStatus{
			IP:           ip,
			ProcessCount: 0,
			Error:        fmt.Sprintf("Failed to parse process count: %v", err),
		}
	}

	return NodeStatus{
		IP:           ip,
		ProcessCount: count,
	}
}
