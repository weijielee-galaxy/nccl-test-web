package handlers

import (
	"regexp"
	"strconv"
	"strings"
)

// 正则表达式模式常量
const (
	// 匹配空白字符
	whitespacePattern = `\s+`
	// 匹配 NCCL 输出的表头行，包含所有关键字段
	tableHeaderPattern = `^\s*#.*size.*count.*type.*time.*algbw.*busbw`
	// 匹配表格结束标记
	tableEndPattern = `^\s*#\s*(Out of bounds|Avg bus bandwidth)`
	// 匹配数据行：数字开头，后面跟着至少11个空白分隔的字段（总共12个字段）
	// 格式：size count type redop root time algbw busbw #wrong time algbw busbw #wrong
	dataLinePattern = `^\s*\d+\s+\d+\s+\S+\s+\S+\s+\S+\s+[\d.]+\s+[\d.]+\s+[\d.]+\s+\d+\s+[\d.]+\s+[\d.]+\s+[\d.]+`
)

// ChartDataPoint 表示图表的一个数据点
type ChartDataPoint struct {
	Size     int     `json:"size"`
	Count    int     `json:"count"`
	Type     string  `json:"type"`
	OutAlgbw float64 `json:"outAlgbw"`
	OutBusbw float64 `json:"outBusbw"`
	InAlgbw  float64 `json:"inAlgbw"`
	InBusbw  float64 `json:"inBusbw"`
}

// NCCLOutputParser NCCL 输出解析器
type NCCLOutputParser struct {
	whitespaceRegex  *regexp.Regexp
	tableHeaderRegex *regexp.Regexp
	tableEndRegex    *regexp.Regexp
	dataLineRegex    *regexp.Regexp
}

// NewNCCLOutputParser 创建新的 NCCL 输出解析器
func NewNCCLOutputParser() *NCCLOutputParser {
	return &NCCLOutputParser{
		whitespaceRegex:  regexp.MustCompile(whitespacePattern),
		tableHeaderRegex: regexp.MustCompile(tableHeaderPattern),
		tableEndRegex:    regexp.MustCompile(tableEndPattern),
		dataLineRegex:    regexp.MustCompile(dataLinePattern),
	}
}

// Parse 解析 NCCL 测试输出，提取图表数据
func (p *NCCLOutputParser) Parse(output string) []ChartDataPoint {
	if output == "" {
		return []ChartDataPoint{}
	}

	lines := strings.Split(output, "\n")
	data := []ChartDataPoint{}
	parsing := false

	for _, line := range lines {
		if p.isTableHeader(line) {
			parsing = true
			continue
		}

		if p.isTableEnd(line) {
			break
		}

		if parsing {
			if dataPoint, ok := p.parseDataLine(line); ok {
				data = append(data, dataPoint)
			}
		}
	}

	return data
}

// isTableHeader 检查是否为表头行
// 使用正则表达式匹配，比多次 strings.Contains 更高效
func (p *NCCLOutputParser) isTableHeader(line string) bool {
	return p.tableHeaderRegex.MatchString(line)
}

// isTableEnd 检查是否为表格结束行
// 使用正则表达式匹配结束标记
func (p *NCCLOutputParser) isTableEnd(line string) bool {
	return p.tableEndRegex.MatchString(line)
}

// isDataLine 检查是否为有效的数据行
// 使用正则表达式判断是否以数字开头（size 字段），排除 DEBUG/INFO 日志
func (p *NCCLOutputParser) isDataLine(line string) bool {
	return p.dataLineRegex.MatchString(line)
}

// parseDataLine 解析单行数据
func (p *NCCLOutputParser) parseDataLine(line string) (ChartDataPoint, bool) {
	if !p.isDataLine(line) {
		return ChartDataPoint{}, false
	}

	fields := p.splitFields(line)
	if len(fields) < 12 {
		return ChartDataPoint{}, false
	}

	size, err := strconv.Atoi(fields[0])
	if err != nil || size <= 0 {
		return ChartDataPoint{}, false
	}

	count, _ := strconv.Atoi(fields[1])
	outAlgbw, _ := strconv.ParseFloat(fields[6], 64)
	outBusbw, _ := strconv.ParseFloat(fields[7], 64)
	inAlgbw, _ := strconv.ParseFloat(fields[10], 64)
	inBusbw, _ := strconv.ParseFloat(fields[11], 64)

	dataPoint := ChartDataPoint{
		Size:     size,
		Count:    count,
		Type:     fields[2],
		OutAlgbw: outAlgbw,
		OutBusbw: outBusbw,
		InAlgbw:  inAlgbw,
		InBusbw:  inBusbw,
	}

	return dataPoint, true
}

// splitFields 分割字段并过滤空字符串
func (p *NCCLOutputParser) splitFields(line string) []string {
	parts := p.whitespaceRegex.Split(strings.TrimSpace(line), -1)

	fields := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			fields = append(fields, part)
		}
	}

	return fields
}

// ParseNCCLOutput 解析 NCCL 测试输出（便捷函数）
func ParseNCCLOutput(output string) []ChartDataPoint {
	parser := NewNCCLOutputParser()
	return parser.Parse(output)
}

// ExtractRawDataLines 提取原始数据行，用于调试和验证
// 返回所有被识别为数据行的原始文本，方便人工检查解析是否正确
func ExtractRawDataLines(output string) []string {
	if output == "" {
		return []string{}
	}

	parser := NewNCCLOutputParser()
	lines := strings.Split(output, "\n")
	rawDataLines := []string{}
	parsing := false

	for _, line := range lines {
		if parser.isTableHeader(line) {
			parsing = true
			continue
		}

		if parser.isTableEnd(line) {
			break
		}

		if parsing && parser.isDataLine(line) {
			rawDataLines = append(rawDataLines, line)
		}
	}

	return rawDataLines
}
