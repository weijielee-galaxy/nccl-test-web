package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// DataDir 数据存储目录
	DataDir = "./data"
	// IPListFileName IP列表文件名
	IPListFileName = "iplist"
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

// SaveIPList 保存IP列表到文件
func SaveIPList(c *gin.Context) {
	var req IPListRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// 确保data目录存在
	if err := os.MkdirAll(DataDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create data directory",
		})
		return
	}

	// 将IP列表转换为字符串，每行一个IP
	content := strings.Join(req.IPList, "\n")
	if len(req.IPList) > 0 {
		content += "\n" // 最后添加换行符
	}

	// 写入文件
	filePath := filepath.Join(DataDir, IPListFileName)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save IP list",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "IP list saved successfully",
		"count":   len(req.IPList),
	})
}

// GetIPList 获取IP列表（查询）
func GetIPList(c *gin.Context) {
	filePath := filepath.Join(DataDir, IPListFileName)

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
	var req IPListRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// 确保data目录存在
	if err := os.MkdirAll(DataDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create data directory",
		})
		return
	}

	// 将IP列表转换为字符串，每行一个IP
	content := strings.Join(req.IPList, "\n")
	if len(req.IPList) > 0 {
		content += "\n"
	}

	// 写入文件
	filePath := filepath.Join(DataDir, IPListFileName)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update IP list",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "IP list updated successfully",
		"count":   len(req.IPList),
	})
}

// DeleteIPList 删除IP列表（删除）
func DeleteIPList(c *gin.Context) {
	filePath := filepath.Join(DataDir, IPListFileName)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{
			"message": "IP list does not exist",
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
