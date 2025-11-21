package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// HistoryDir 历史记录存储目录
	HistoryDir = "data/history"
)

// HistoryRecord 历史记录信息
type HistoryRecord struct {
	Filename string    `json:"filename"`
	Modified time.Time `json:"modified"`
}

// HistoryContent 历史记录内容
type HistoryContent struct {
	Output string `json:"output"`
	Status string `json:"status"`
}

// SaveHistoryAsync 异步保存测试历史数据
func SaveHistoryAsync(output string) {
	go func() {
		if err := saveHistoryFile(output); err != nil {
			fmt.Printf("Failed to save history: %v\n", err)
		}
	}()
}

// saveHistoryFile 保存历史文件（同步操作，在 goroutine 中运行）
func saveHistoryFile(output string) error {
	// 创建历史目录
	if err := os.MkdirAll(HistoryDir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %v", err)
	}

	// 生成文件名：20251120_143022.txt
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(HistoryDir, timestamp+".txt")

	// 保存文件
	if err := os.WriteFile(filename, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write history file: %v", err)
	}

	fmt.Printf("History saved to %s\n", filename)
	return nil
}

// GetHistoryList 获取历史记录列表
func GetHistoryList(c *gin.Context) {
	// 读取历史目录
	entries, err := os.ReadDir(HistoryDir)
	if err != nil {
		if os.IsNotExist(err) {
			// 目录不存在，返回空列表
			c.JSON(http.StatusOK, gin.H{
				"count":   0,
				"records": []HistoryRecord{},
				"status":  "success",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read history directory",
		})
		return
	}

	// 收集所有记录
	var records []HistoryRecord
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// 只收集 .txt 文件
		if filepath.Ext(entry.Name()) != ".txt" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		records = append(records, HistoryRecord{
			Filename: entry.Name(),
			Modified: info.ModTime(),
		})
	}

	// 按修改时间倒序排列（最新的在前）
	sort.Slice(records, func(i, j int) bool {
		return records[i].Modified.After(records[j].Modified)
	})

	c.JSON(http.StatusOK, gin.H{
		"count":   len(records),
		"records": records,
		"status":  "success",
	})
}

// GetHistoryContent 获取指定历史记录的内容
func GetHistoryContent(c *gin.Context) {
	filename := c.Param("filename")

	// 安全检查：防止路径遍历攻击
	if filename == "" || filepath.Dir(filename) != "." {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}

	// 构建完整路径
	filePath := filepath.Join(HistoryDir, filename)

	// 验证文件确实在 HistoryDir 下
	absPath, _ := filepath.Abs(filePath)
	absHistoryDir, _ := filepath.Abs(HistoryDir)
	if !filepath.HasPrefix(absPath, absHistoryDir) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file path",
		})
		return
	}

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "History record not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read history file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"output": string(data),
		"status": "success",
	})
}

// DeleteHistory 删除指定的历史记录
func DeleteHistory(c *gin.Context) {
	filename := c.Param("filename")

	// 安全检查：防止路径遍历攻击
	if filename == "" || filepath.Dir(filename) != "." {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}

	// 构建完整路径
	filePath := filepath.Join(HistoryDir, filename)

	// 验证文件确实在 HistoryDir 下
	absPath, _ := filepath.Abs(filePath)
	absHistoryDir, _ := filepath.Abs(HistoryDir)
	if !filepath.HasPrefix(absPath, absHistoryDir) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file path",
		})
		return
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "History record not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete history file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "History record deleted successfully",
		"status":  "success",
	})
}
