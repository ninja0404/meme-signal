package publisher

import (
	"fmt"
	"strconv"

	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/ninja0404/meme-signal/internal/notifier"
	"github.com/ninja0404/meme-signal/internal/repo"
)

// FeishuPublisher 飞书发布器
type FeishuPublisher struct {
	webhookURL      string
	tokenInfoRepo   repo.TokenInfoRepo
	tokenHolderRepo repo.TokenHolderRepo
}

// NewFeishuPublisher 创建飞书发布器
func NewFeishuPublisher(webhookURL string, tokenInfoRepo repo.TokenInfoRepo, tokenHolderRepo repo.TokenHolderRepo) *FeishuPublisher {
	return &FeishuPublisher{
		webhookURL:      webhookURL,
		tokenInfoRepo:   tokenInfoRepo,
		tokenHolderRepo: tokenHolderRepo,
	}
}

func (p *FeishuPublisher) GetType() string {
	return "feishu"
}

func (p *FeishuPublisher) Publish(signal *model.Signal) error {
	// 格式化消息内容
	message := p.formatSignalMessage(signal)

	// 发送到飞书
	return notifier.SendToLark(message, p.webhookURL)
}

func (p *FeishuPublisher) Close() error {
	return nil
}

// formatVolume 格式化交易量，大于1000时显示为k格式
func (p *FeishuPublisher) formatVolume(volumeStr string) string {
	// 解析数值
	volume, err := strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return "$" + volumeStr // 解析失败时返回原值
	}

	if volume >= 1000000 {
		// 大于等于100万，显示为M格式
		return fmt.Sprintf("$%.1fM", volume/1000000)
	} else if volume >= 1000 {
		// 大于等于1000，显示为k格式
		return fmt.Sprintf("$%.1fk", volume/1000)
	} else {
		// 小于1000，保持原格式
		return fmt.Sprintf("$%.2f", volume)
	}
}

// formatSignalMessage 格式化信号消息
func (p *FeishuPublisher) formatSignalMessage(signal *model.Signal) string {
	// 从signal.Data中获取信息
	tokenAddr := signal.TokenAddress
	currentPrice := "N/A"
	priceChange5m := "N/A"
	uniqueWallets := "N/A"
	txCount5m := "N/A"
	volume5m := "N/A"

	// 查询代币symbol
	tokenSymbol := "UNKNOWN"
	if p.tokenInfoRepo != nil {
		if symbol, err := p.tokenInfoRepo.GetTokenSymbol(tokenAddr); err == nil && symbol != "" {
			tokenSymbol = symbol
		}
	}

	// 查询持仓人数
	holderCount := "N/A"
	if p.tokenHolderRepo != nil {
		if count, err := p.tokenHolderRepo.GetHolderCount(tokenAddr); err == nil {
			holderCount = fmt.Sprintf("%d个", count)
		}
	}

	// 从Data字段获取详细信息
	if signal.Data != nil {
		if price, ok := signal.Data["current_price"].(string); ok {
			currentPrice = "$" + price
		}
		if change, ok := signal.Data["price_change_5m"].(string); ok {
			priceChange5m = change + "%"
		}
		if wallets, ok := signal.Data["unique_wallets"].(int); ok {
			uniqueWallets = fmt.Sprintf("%d个", wallets)
		}
		if txCount, ok := signal.Data["tx_count_5m"].(int); ok {
			txCount5m = fmt.Sprintf("%d笔", txCount)
		}
		if volume, ok := signal.Data["volume_5m"].(string); ok {
			volume5m = p.formatVolume(volume)
		}
	}

	message := fmt.Sprintf(`🚨 Meme交易信号检测

🪙 代币符号: %s
📍 代币地址: %s
💰 当前价格: %s
📈 5分钟涨幅: %s
👥 独立地址数: %s
🏦 持仓人数: %s
📊 5分钟交易数: %s
💵 5分钟交易量: %s

⏰ 触发时间: %s`,
		tokenSymbol,
		tokenAddr,
		currentPrice,
		priceChange5m,
		uniqueWallets,
		holderCount,
		txCount5m,
		volume5m,
		signal.Timestamp.Format("2006-01-02 15:04:05"))

	return message
}
