1. 整体程序结构重新调整 单一职责
2. consumer masscan拆分 ✅
3. masscan用白名单机制提取✅ 扫描结果通过channel传递出去 不用slice ✅ 对应internal/scanner/masscan.go
4. 回调函数 internal里面基础组件不要有交叉 ✅
5. 先写伪代码

masscan 47.93.190.244/24 -p8080 --rate 100 -oJ - 1>stdout.log 2>stderr.log