package handlers

import (
	"fmt"
	"testing"
)

const sampleNCCLOutput = `[cetus-g88-094] running nccl test all_reduce -b 1 -e 1G, world_size=16
# nGpus(perProc) 1 minBytes 1 maxBytes 1073741824 step: 2(factor) warmup iters: 5 iters: 20 agg iters: 1 validation: 1 graph: 0
#
# Using devices
#  Rank  7 Group  0 Pid 2562583 on cetus-g88-094 device  7 [0xd7] NVIDIA H200
#  Rank 15 Group  0 Pid 1924610 on cetus-g88-061 device  7 [0xd7] NVIDIA H200
NCCL version 2.27.7+cuda12.4
#
#                                                              out-of-place                       in-place          
#       size         count      type   redop    root     time   algbw   busbw #wrong     time   algbw   busbw #wrong
#        (B)    (elements)                               (us)  (GB/s)  (GB/s)            (us)  (GB/s)  (GB/s)       
cetus-g88-061:3259226:3260615 [7] NCCL INFO Connected all trees
cetus-g88-061:3259225:3260619 [6] NCCL INFO Connected all trees
cetus-g88-061:3259220:3260623 [5] NCCL INFO Connected all trees
cetus-g88-061:3259203:3260616 [0] NCCL INFO Connected all trees
cetus-g88-061:3259206:3260617 [1] NCCL INFO Connected all trees
    579.3    0.00    0.00      0    137.0    0.00    0.00      0
           0             0  bfloat16     sum      -1     0.36    0.00    0.00      0     0.33    0.00    0.00      0
           2             1  bfloat16     sum      -1   2500.6    0.00    0.00      0   3291.0    0.00    0.00      0
           4             2  bfloat16     sum      -1    142.6    0.00    0.00      0    142.6    0.00    0.00      0
           8             4  bfloat16     sum      -1    142.5    0.00    0.00      0    142.5    0.00    0.00      0
          16             8  bfloat16     sum      -1    142.4    0.00    0.00      0    141.4    0.00    0.00      0
          32            16  bfloat16     sum      -1    142.6    0.00    0.00      0   3541.4    0.00    0.00      0
          64            32  bfloat16     sum      -1    143.6    0.00    0.00      0    144.5    0.00    0.00      0
         128            64  bfloat16     sum      -1    144.1    0.00    0.00      0    143.7    0.00    0.00      0
         256           128  bfloat16     sum      -1    177.5    0.00    0.00      0    575.5    0.00    0.00      0
         512           256  bfloat16     sum      -1    342.4    0.00    0.00      0    351.8    0.00    0.00      0
        1024           512  bfloat16     sum      -1    144.4    0.01    0.01      0    145.3    0.01    0.01      0
        2048          1024  bfloat16     sum      -1    146.2    0.01    0.03      0    146.6    0.01    0.03      0
        4096          2048  bfloat16     sum      -1    148.4    0.03    0.05      0    147.9    0.03    0.05      0
        8192          4096  bfloat16     sum      -1    153.7    0.05    0.10      0    150.5    0.05    0.10      0
       16384          8192  bfloat16     sum      -1    154.1    0.11    0.20      0    150.8    0.11    0.20      0
       32768         16384  bfloat16     sum      -1    153.5    0.21    0.40      0    152.0    0.22    0.40      0
       65536         32768  bfloat16     sum      -1    153.6    0.43    0.80      0    151.6    0.43    0.81      0
      131072         65536  bfloat16     sum      -1    157.6    0.83    1.56      0    153.3    0.85    1.60      0
      262144        131072  bfloat16     sum      -1   4068.8    0.06    0.12      0    166.1    1.58    2.96      0
      524288        262144  bfloat16     sum      -1    213.0    2.46    4.62      0    195.0    2.69    5.04      0
     1048576        524288  bfloat16     sum      -1    173.0    6.06   11.36      0   1037.1    1.01    1.90      0
     2097152       1048576  bfloat16     sum      -1    188.3   11.14   20.88      0    183.1   11.45   21.48      0
     4194304       2097152  bfloat16     sum      -1    369.8   11.34   21.26      0    455.2    9.21   17.28      0
     8388608       4194304  bfloat16     sum      -1   3365.2    2.49    4.67      0   1942.5    4.32    8.10      0
    16777216       8388608  bfloat16     sum      -1    442.6   37.90   71.07      0    444.0   37.79   70.85      0
    33554432      16777216  bfloat16     sum      -1   4582.2    7.32   13.73      0   4568.9    7.34   13.77      0
    67108864      33554432  bfloat16     sum      -1   2090.7   32.10   60.19      0   4596.7   14.60   27.37      0
   134217728      67108864  bfloat16     sum      -1   4708.6   28.50   53.45      0   4215.9   31.84   59.69      0
   268435456     134217728  bfloat16     sum      -1   8721.5   30.78   57.71      0   9455.5   28.39   53.23      0
   536870912     268435456  bfloat16     sum      -1   8299.2   64.69  121.29      0    11804   45.48   85.28      0
  1073741824     536870912  bfloat16     sum      -1    22645   47.42   88.91      0    23322   46.04   86.33      0
# Out of bounds values : 0 OK
# Avg bus bandwidth    : 15.9501 
#`

func TestTableHeaderRecognition(t *testing.T) {
	parser := NewNCCLOutputParser()

	testCases := []struct {
		line     string
		expected bool
		desc     string
	}{
		{
			line:     "#       size         count      type   redop    root     time   algbw   busbw #wrong     time   algbw   busbw #wrong",
			expected: true,
			desc:     "标准表头",
		},
		{
			line:     "# Out of bounds values : 0 OK",
			expected: false,
			desc:     "结束标记",
		},
		{
			line:     "           0             0  bfloat16     sum      -1     0.36    0.00    0.00      0     0.33    0.00    0.00      0",
			expected: false,
			desc:     "数据行",
		},
		{
			line:     "cetus-g88-061:1263256:1263256 [0] NCCL INFO Bootstrap",
			expected: false,
			desc:     "DEBUG日志",
		},
	}

	fmt.Println("\n表头识别测试：")
	for _, tc := range testCases {
		result := parser.isTableHeader(tc.line)
		status := "✓"
		if result != tc.expected {
			status = "✗"
			t.Errorf("失败")
		}
		preview := tc.line
		if len(preview) > 50 {
			preview = preview[:50]
		}
		fmt.Printf("  %s [%s] 预期: %v, 实际: %v | %s", status, tc.desc, tc.expected, result, preview)
	}
}

func TestExtractRawDataLines(t *testing.T) {
	rawLines := ExtractRawDataLines(sampleNCCLOutput)

	fmt.Printf("提取到 %d 条原始数据行\n", len(rawLines))

	fmt.Println("\n所有原始数据行：")
	for i, line := range rawLines {
		fmt.Printf("[%2d] %s\n", i+1, line)
	}
}

func TestParseNCCLOutput(t *testing.T) {
	data := ParseNCCLOutput(sampleNCCLOutput)

	fmt.Printf("解析到 %d 条数据行\n", len(data))

	fmt.Println("\n解析结果：")
	for i, d := range data {
		fmt.Printf("  [%d] Size: %d, Count: %d, Type: %s, OutBusbw: %.2f, InBusbw: %.2f",
			i, d.Size, d.Count, d.Type, d.OutBusbw, d.InBusbw)
	}
}
