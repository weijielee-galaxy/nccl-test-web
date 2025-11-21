package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// DataDir 数据存储目录
	DataDir = "./data"
	// IPListDir IP列表存储目录
	IPListDir = "iplist"
	// IPListFileName IP列表文件名（已弃用，保留用于向后兼容）
	IPListFileName = "iplist"
	// DefaultIPListFile 默认IP列表文件名
	DefaultIPListFile = "default"
)

// IPListRequest 接收IP列表的请求结构
type IPListRequest struct {
	IPList []string `json:"iplist" binding:"required"`
}

// IPListResponse 返回IP列表的响应结构
type IPListResponse struct {
	IPList []string `json:"iplist"`
	Count  int      `json:"count"`
}

// IPListFileInfo IP列表文件信息
type IPListFileInfo struct {
	Filename string    `json:"filename"`
	Modified time.Time `json:"modified"`
	Size     int64     `json:"size"`
}

// GetIPListFiles 获取所有IP列表文件
func GetIPListFiles(c *gin.Context) {
	iplistDir := filepath.Join(DataDir, IPListDir)

	// 检查目录是否存在，不存在则创建并初始化默认文件
	if _, err := os.Stat(iplistDir); os.IsNotExist(err) {
		// 创建目录
		if err := os.MkdirAll(iplistDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create IP list directory",
			})
			return
		}

		// 创建默认文件（空文件）
		defaultFile := filepath.Join(iplistDir, DefaultIPListFile)
		if err := os.WriteFile(defaultFile, []byte(""), 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create default IP list file",
			})
			return
		}
	}

	// 读取目录
	entries, err := os.ReadDir(iplistDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read IP list directory",
		})
		return
	}

	// 收集所有文件（不过滤文件扩展名）
	var files []IPListFileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, IPListFileInfo{
			Filename: entry.Name(),
			Modified: info.ModTime(),
			Size:     info.Size(),
		})
	}

	// 按修改时间倒序排列
	sort.Slice(files, func(i, j int) bool {
		return files[i].Modified.After(files[j].Modified)
	})

	c.JSON(http.StatusOK, gin.H{
		"count": len(files),
		"files": files,
	})
}

// SaveIPList 保存IP列表到指定文件
func SaveIPList(c *gin.Context) {
	filename := c.Param("filename")

	// 如果没有指定文件名，使用默认值
	if filename == "" {
		filename = DefaultIPListFile
	}

	// 安全检查
	if filename == "" || filepath.Dir(filename) != "." {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}

	var req IPListRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// 确保目录存在
	iplistDir := filepath.Join(DataDir, IPListDir)
	if err := os.MkdirAll(iplistDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create IP list directory",
		})
		return
	}

	// 构建完整路径
	filePath := filepath.Join(iplistDir, filename)

	// 安全检查：防止路径遍历
	absPath, _ := filepath.Abs(filePath)
	absIPListDir, _ := filepath.Abs(iplistDir)
	if !filepath.HasPrefix(absPath, absIPListDir) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file path",
		})
		return
	}

	// 将IP列表转换为字符串，每行一个IP
	content := strings.Join(req.IPList, "\n")
	if len(req.IPList) > 0 {
		content += "\n" // 最后添加换行符
	}

	// 写入文件
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save IP list",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "IP list saved successfully",
		"filename": filename,
		"count":    len(req.IPList),
	})
}

// GetIPList 获取IP列表（查询）
func GetIPList(c *gin.Context) {
	filename := c.Param("filename")

	// 如果没有指定文件名，使用默认值
	if filename == "" {
		filename = DefaultIPListFile
	}

	// 安全检查
	if filename == "" || filepath.Dir(filename) != "." {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}

	iplistDir := filepath.Join(DataDir, IPListDir)
	filePath := filepath.Join(iplistDir, filename)

	// 安全检查：防止路径遍历
	absPath, _ := filepath.Abs(filePath)
	absIPListDir, _ := filepath.Abs(iplistDir)
	if !filepath.HasPrefix(absPath, absIPListDir) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file path",
		})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusOK, IPListResponse{
			IPList: []string{},
			Count:  0,
		})
		return
	}

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read IP list",
		})
		return
	}

	// 解析IP列表
	content := strings.TrimSpace(string(data))
	var ipList []string
	if content != "" {
		ipList = strings.Split(content, "\n")
		// 过滤空行
		filteredList := make([]string, 0, len(ipList))
		for _, ip := range ipList {
			if trimmed := strings.TrimSpace(ip); trimmed != "" {
				filteredList = append(filteredList, trimmed)
			}
		}
		ipList = filteredList
	}

	c.JSON(http.StatusOK, IPListResponse{
		IPList: ipList,
		Count:  len(ipList),
	})
}

// UpdateIPList 更新IP列表（修改）
func UpdateIPList(c *gin.Context) {
	filename := c.Param("filename")

	// 如果没有指定文件名，使用默认值
	if filename == "" {
		filename = DefaultIPListFile
	}

	// 安全检查
	if filename == "" || filepath.Dir(filename) != "." {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}

	var req IPListRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// 确保目录存在
	iplistDir := filepath.Join(DataDir, IPListDir)
	if err := os.MkdirAll(iplistDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create IP list directory",
		})
		return
	}

	// 构建完整路径
	filePath := filepath.Join(iplistDir, filename)

	// 安全检查：防止路径遍历
	absPath, _ := filepath.Abs(filePath)
	absIPListDir, _ := filepath.Abs(iplistDir)
	if !filepath.HasPrefix(absPath, absIPListDir) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file path",
		})
		return
	}

	// 将IP列表转换为字符串，每行一个IP
	content := strings.Join(req.IPList, "\n")
	if len(req.IPList) > 0 {
		content += "\n"
	}

	// 写入文件
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update IP list",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "IP list updated successfully",
		"filename": filename,
		"count":    len(req.IPList),
	})
}

// DeleteIPList 删除IP列表（删除）
func DeleteIPList(c *gin.Context) {
	filename := c.Param("filename")

	// 如果没有指定文件名，使用默认值
	if filename == "" {
		filename = DefaultIPListFile
	}

	// 安全检查
	if filename == "" || filepath.Dir(filename) != "." {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}

	iplistDir := filepath.Join(DataDir, IPListDir)
	filePath := filepath.Join(iplistDir, filename)

	// 安全检查：防止路径遍历
	absPath, _ := filepath.Abs(filePath)
	absIPListDir, _ := filepath.Abs(iplistDir)
	if !filepath.HasPrefix(absPath, absIPListDir) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file path",
		})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{
			"message": "IP list file does not exist",
		})
		return
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete IP list",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "IP list deleted successfully",
	})
}
