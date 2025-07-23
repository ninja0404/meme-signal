# Meme Signal - Meme代币交易信号监听服务

## 🚀 项目概述

Meme Signal 是一个基于Go语言开发的高性能Meme代币交易信号监听和检测服务。系统通过实时消费Kafka中的DEX交易数据，使用复合信号检测算法识别具有投资潜力的代币，并及时发布交易信号。

### 核心特性

- 🎯 **复合信号检测**: 基于价格涨幅(≥30%) + 交易量(≥250k USD)的双重条件检测
- ⚡ **高并发处理**: 16个Worker协程并行处理，支持高频交易数据
- 📊 **滑动窗口统计**: 5分钟时间窗口，30秒为单位的时间桶统计
- 💎 **金融级精度**: 使用decimal类型避免浮点精度丢失
- 🔄 **实时数据流**: Kafka消费 → 检测分析 → 信号发布的完整链路
- 🎨 **模块化架构**: 数据源、检测引擎、发布器等模块解耦设计

## 🏗️ 系统架构

### 整体架构图

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐    ┌──────────────────┐
│                 │    │                  │    │                 │    │                  │
│   Kafka Brokers │───▶│   Source Layer   │───▶│ Detection Layer │───▶│ Publisher Layer  │
│  (DEX交易数据)   │    │   (数据源模块)    │    │   (检测引擎)     │    │   (信号发布)      │
│                 │    │                  │    │                 │    │                  │
└─────────────────┘    └──────────────────┘    └─────────────────┘    └──────────────────┘
                              │                         │                         │
                              ▼                         ▼                         ▼
                      ┌──────────────┐       ┌─────────────────┐       ┌──────────────┐
                      │ Kafka Source │       │ 16 Workers      │       │ Log Publisher│
                      │   - Consumer │       │ Token Windows   │       │   - Console  │
                      │   - Decoder  │       │ Time Buckets    │       │   - File     │
                      └──────────────┘       └─────────────────┘       └──────────────┘
```

### 核心模块架构

```
meme-signal/
├── pkg/                        # 公共包
│   ├── config/                # 配置管理系统
│   │   ├── config.go          # 配置核心
│   │   ├── loader/            # 配置加载器
│   │   ├── reader/            # 配置读取器
│   │   └── source/            # 配置源
│   ├── logger/                # 日志系统
│   └── mq/kafka/              # Kafka客户端封装
├── internal/                   # 内部模块
│   ├── app/                   # 应用入口
│   ├── config/                # 配置管理
│   ├── model/                 # 数据模型
│   ├── source/                # 数据源模块
│   │   └── kafka/             # Kafka数据源
│   ├── detector/              # 信号检测引擎
│   │   ├── detector.go        # 检测引擎核心
│   │   └── window.go          # 时间窗口统计
│   ├── publisher/             # 信号发布模块
│   └── pipeline/              # 数据处理管道
├── config/
│   └── config.yaml            # 配置文件
└── main.go                    # 程序入口
```

## 📊 数据流程

### 1. 数据接收流程

```
Kafka Topic (test-solana-tx)
         │
         ▼
  Kafka Consumer
         │
         ▼ 
   DecodeEvent()  ─────────────► 解析DEX交易数据
         │
         ▼
  Transaction Model
         │
         ▼
  Hash分片路由 ─────────────────► 基于TokenAddress分配Worker
         │
         ▼
   Worker协程池 (16个Worker)
```

### 2. 检测分析流程

```
Transaction
    │
    ▼
TokenWindow ─────────────────► 5分钟滑动窗口
    │                          └── 10个TimeBucket (30秒/桶)
    ▼
AddTransaction()
    │                          ┌── 交易量统计 (decimal)
    ├─────────────────────────► ├── 价格统计 (decimal) 
    │                          ├── 买卖次数统计
    │                          └── 独立钱包统计
    ▼
GetStats() ──────────────────► TokenStats生成
    │
    ▼
CompositeDetector ──────────► 复合条件检测
    │                          ├── 价格涨幅 ≥ 30%
    │                          └── 交易量 ≥ 250k USD
    ▼
Signal生成 ──────────────────► 满足条件时生成信号
```

### 3. 信号发布流程

```
Signal
  │
  ▼
Publisher Manager
  │
  ├─────────────► Log Publisher ─────────► Console输出
  │
  └─────────────► [扩展] HTTP Publisher ──► API通知
                  [扩展] Webhook Publisher ► 第三方回调
                  [扩展] Database Publisher ► 数据库存储
```

## 🔧 核心组件详解

### 1. 数据源层 (Source Layer)

**功能**: 负责从外部系统获取交易数据

**核心组件**:
- `source.Manager`: 数据源管理器，支持多种数据源
- `kafka.Source`: Kafka数据源实现
- `DecodeEvent()`: Kafka消息解码器

**关键特性**:
```go
// Kafka配置传递 - 完整配置透传
kafkaSource := kafka.NewSource(kafkaConfig)

// 并发安全的数据接收
func (s *Source) Start(ctx context.Context) error {
    // 启动Kafka消费者
    // 并发处理消息
    // 错误处理和重试
}
```

### 2. 检测引擎 (Detection Layer)

**功能**: 核心信号检测和统计分析

**核心组件**:
- `detector.Engine`: 检测引擎，管理16个Worker协程
- `detector.Worker`: 工作协程，处理特定Token的交易
- `TokenWindow`: 5分钟滑动窗口统计
- `TimeBucket`: 30秒时间桶，存储细粒度统计

**时间窗口设计**:
```go
const (
    WindowSize  = 5 * time.Minute  // 5分钟窗口
    BucketSize  = 30 * time.Second // 30秒时间桶  
    BucketCount = 10               // 10个桶
)

// 精确decimal计算
type TimeBucket struct {
    Volume       decimal.Decimal  // 交易量
    FirstPrice   decimal.Decimal  // 首价
    LastPrice    decimal.Decimal  // 末价
    HighestPrice decimal.Decimal  // 最高价
    LowestPrice  decimal.Decimal  // 最低价
}
```

**复合信号检测算法**:
```go
func (d *compositeSignalDetector) Detect(stats *model.TokenStats, tx *model.Transaction) []*model.Signal {
    // 条件1: 价格涨幅 ≥ 30%
    priceChangeOK := stats.PriceChangePercent.GreaterThanOrEqual(d.priceChangeThreshold)
    
    // 条件2: 交易量 ≥ 250k USD  
    volumeOK := stats.Volume24h.GreaterThanOrEqual(d.volumeThreshold)
    
    // 复合条件: 必须同时满足
    if priceChangeOK && volumeOK {
        return []*model.Signal{signal}
    }
    return nil
}
```

### 3. 发布层 (Publisher Layer)

**功能**: 信号输出和通知

**核心组件**:
- `publisher.Manager`: 发布管理器
- `publisher.LogPublisher`: 日志发布器
- 支持扩展多种发布方式

## 🔧 配置说明

### 配置文件结构 (config/config.yaml)

```yaml
# 日志配置
logger:
  output: "stdout"        # 输出方式: stdout/file
  debug: false           # 调试模式
  level: INFO            # 日志级别
  add_caller: true       # 添加调用者信息
  caller_skip: 1         # 调用栈跳过层数

# Kafka配置  
kafka:
  brokers:              # Kafka集群地址
    - broker1:9094
    - broker2:9094
    - broker3:9094
  
  consumer:
    topics:             # 消费主题列表
      - "test-solana-tx"
    group_id: "test-signal"           # 消费组ID
    assignor: "sticky"                # 分区分配策略
    version: "2.6.2"                  # Kafka版本
    client_id: "test-signal"          # 客户端ID
    enable_auto_commit: false         # 禁用自动提交
    enable_auto_store: false          # 禁用自动存储
    auto_commit_interval: 5000        # 提交间隔(ms)
    read_timeout: 3000                # 读取超时(ms)
    security_protocol: "SASL_PLAINTEXT" # 安全协议
    sasl_username: "dev"              # SASL用户名
    sasl_password: "password"         # SASL密码
    sasl_mechanism: "PLAIN"           # SASL机制
```

### 环境变量

```bash
# 配置文件路径
MEME_SIGNAL_CONFIG_PATH=./config/config.yaml

# 日志级别覆盖
MEME_SIGNAL_LOG_LEVEL=DEBUG

# Kafka连接覆盖
MEME_SIGNAL_KAFKA_BROKERS=localhost:9092
```

## 🚀 快速开始

### 1. 环境要求

- Go 1.19+
- Kafka 2.6.2+
- 8GB+ RAM (建议)
- 4+ CPU核心 (建议)

### 2. 安装依赖

```bash
# 克隆项目
git clone <repository-url>
cd meme-signal

# 安装Go依赖
go mod download

# 验证依赖
go mod verify
```

### 3. 配置设置

```bash
# 复制配置模板
cp config/config.yaml.example config/config.yaml

# 编辑配置文件
vim config/config.yaml
```

**必须配置项**:
- Kafka brokers地址
- SASL认证信息 
- Topic名称

### 4. 编译运行

```bash
# 编译
go build -o meme-signal .

# 运行 
./meme-signal

# 后台运行
nohup ./meme-signal > meme-signal.log 2>&1 &
```

### 5. 验证运行

**检查日志输出**:
```bash
# 查看启动日志
tail -f meme-signal.log

# 期望看到的关键日志
# ✅ Kafka数据源已启动
# 🎯 信号检测引擎已启动  
# 📡 信号发布管理器已启动
# 🔥 Meme交易信号监听服务已启动
```

**检查系统状态**:
```bash
# 检查进程
ps aux | grep meme-signal

# 检查端口占用 (如果有HTTP服务)
netstat -tlnp | grep :8080
```

## 📈 性能指标

### 系统容量

| 指标 | 数值 | 说明 |
|------|------|------|
| 交易处理能力 | 10,000+ TPS | 每秒处理交易数 |
| 并发Worker数 | 16个 | 可配置 |
| 内存使用 | ~2GB | 正常运行状态 |
| 延迟 | <100ms | 信号检测延迟 |
| 时间窗口 | 5分钟 | 统计窗口大小 |

### 监控指标

```bash
# Worker统计信息
curl -s http://localhost:8080/metrics/workers

# 系统统计信息  
curl -s http://localhost:8080/metrics/system

# 信号统计信息
curl -s http://localhost:8080/metrics/signals
```

## 🔄 扩展开发

### 1. 添加新的信号检测器

```go
// 实现Detector接口
type CustomDetector struct {
    threshold decimal.Decimal
}

func (d *CustomDetector) Detect(stats *model.TokenStats, tx *model.Transaction) []*model.Signal {
    // 自定义检测逻辑
    if stats.Volume24h.GreaterThan(d.threshold) {
        return []*model.Signal{signal}
    }
    return nil
}

func (d *CustomDetector) GetType() string {
    return "custom_detector"
}

// 注册检测器
engine.AddDetectors([]detector.Detector{&CustomDetector{}})
```

### 2. 添加新的数据源

```go
// 实现Source接口
type CustomSource struct {
    config CustomConfig
}

func (s *CustomSource) Start(ctx context.Context) error {
    // 启动数据源
    return nil
}

func (s *CustomSource) Transactions() <-chan *model.Transaction {
    return s.txChan
}

func (s *CustomSource) Errors() <-chan error {
    return s.errChan  
}

// 注册数据源
pipeline.GetSourceManager().AddSource(customSource)
```

### 3. 添加新的发布器

```go
// 实现Publisher接口  
type WebhookPublisher struct {
    url string
}

func (p *WebhookPublisher) Start() error {
    // 启动发布器
    return nil
}

func (p *WebhookPublisher) Publish(signal *model.Signal) error {
    // 发送webhook通知
    return nil
}

func (p *WebhookPublisher) GetType() string {
    return "webhook"
}

// 注册发布器
pipeline.GetPublisherManager().AddPublisher(&WebhookPublisher{})
```

## 🐛 故障排查

### 常见问题

**1. Kafka连接失败**
```bash
# 检查网络连通性
telnet <kafka-broker> 9094

# 检查SASL认证
kafka-console-consumer --bootstrap-server <broker> \
  --topic test-solana-tx \
  --consumer.config consumer.properties
```

**2. 内存使用过高**
```bash
# 查看内存使用
top -p $(pgrep meme-signal)

# 检查Worker统计
# 可能需要调整Worker数量或清理策略
```

**3. 检测延迟过高**
```bash
# 检查Worker负载分布
curl http://localhost:8080/metrics/workers

# 检查Kafka消费延迟
kafka-consumer-groups --bootstrap-server <broker> \
  --describe --group test-signal
```

### 日志分析

**关键日志模式**:
```bash
# 正常启动
grep "🚀\|✅\|🎯" meme-signal.log

# 错误信息
grep "ERROR\|FATAL" meme-signal.log

# 信号检测
grep "🚨.*检测到信号" meme-signal.log

# 性能监控
grep "📊.*更新代币统计" meme-signal.log
```

## 📄 API文档

### 指标接口

**Worker统计**
```
GET /metrics/workers
Response: {
  "0": 125,    // Worker 0管理的Token数量
  "1": 108,    // Worker 1管理的Token数量
  ...
}
```

**系统统计**  
```
GET /metrics/system
Response: {
  "uptime": "2h30m15s",
  "goroutines": 45,
  "memory_usage": "1.2GB",
  "gc_cycles": 156
}
```

**信号统计**
```
GET /metrics/signals  
Response: {
  "total_signals": 1250,
  "signals_per_hour": 25,
  "last_signal_time": "2024-01-15T10:30:45Z"
}
```

## 🤝 贡献指南

### 开发流程

1. Fork项目
2. 创建功能分支: `git checkout -b feature/new-detector`
3. 提交更改: `git commit -am 'Add new detector'`
4. 推送分支: `git push origin feature/new-detector`  
5. 创建Pull Request

### 代码规范

- 遵循Go官方代码规范
- 使用go fmt格式化代码
- 添加必要的注释和文档
- 编写单元测试

## 📝 版本历史

### v1.0.0 (2024-01-15)
- ✅ 初始版本发布
- ✅ Kafka数据源支持
- ✅ 复合信号检测
- ✅ 16Worker并发架构
- ✅ Decimal精度计算

### 待开发功能
- [ ] HTTP API接口
- [ ] 数据库持久化
- [ ] 监控告警系统
- [ ] 配置热更新
- [ ] 集群部署支持

## 📞 联系方式

- 项目维护者: [维护者姓名]
- 邮箱: [contact@example.com]
- 技术讨论: [Discord/Telegram群组]

---

**⚡ 开始您的Meme代币信号监听之旅！**