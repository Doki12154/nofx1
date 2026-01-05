# AI 历史记忆功能说明

## 概述

项目现已支持 AI 历史记忆功能，可以在调用 AI 时自动带入最近 10 轮针对同一币种的历史决策，让 AI 能够基于之前的决策和推理进行连续性思考。

## 实现机制

### 1. 历史存储

所有 AI 决策都会保存到 `decision_records` 表中，包括：

- 系统提示词 (system_prompt)
- 用户提示词 (input_prompt)
- AI 推理过程 (cot_trace)
- 决策列表 (decisions JSON)
- 原始响应 (raw_response)

### 2. 历史检索

新增了 `GetLatestRecordsBySymbol` 方法，可以按币种检索最近 N 轮决策：

```go
// 检索指定 trader 和币种的最近 10 轮决策
records, err := decisionStore.GetLatestRecordsBySymbol(traderID, "BTCUSDT", 10)
```

### 3. 历史回传

修改了 `kernel/engine.go` 中的 `GetFullDecisionWithStrategy` 函数：

- 从 `CallWithMessages` 切换到使用 `RequestBuilder` + `CallWithRequest`
- 自动为每个相关币种（持仓+候选）检索最近 10 轮历史
- 将历史以对话形式添加到请求中：
  - 用户消息：历史上下文摘要
  - 助手消息：AI 之前的决策和推理

## 使用方式

### AutoTrader 中使用

AutoTrader 会自动启用历史记忆功能（如果有 store）：

```go
// 在 buildTradingContext 中自动添加
if at.store != nil {
    ctx.DecisionStore = kernel.NewDecisionStoreAdapter(at.store.Decision())
    ctx.TraderID = at.id
}
```

### 手动使用示例

```go
import (
    "nofx/kernel"
    "nofx/mcp"
    "nofx/store"
)

// 1. 创建 context
ctx := &kernel.Context{
    // ... 其他字段
    DecisionStore: kernel.NewDecisionStoreAdapter(store.Decision()),
    TraderID: "your-trader-id",
}

// 2. 调用 AI（会自动带入历史）
decision, err := kernel.GetFullDecisionWithStrategy(ctx, mcpClient, engine, "balanced")
```

## 历史格式示例

AI 看到的历史消息格式如下：

```
User: [Historical context for BTCUSDT at 14:30:25]
Assistant: Previous decision for BTCUSDT: action=open_long, reasoning=突破关键阻力位，持仓量增加，建议开多...

User: [Historical context for BTCUSDT at 14:35:30]
Assistant: Previous decision for BTCUSDT: action=close_long, reasoning=达到止盈目标，盈利3.5%，建议平仓...

User: [当前市场数据和分析要求]
```

## 优势

1. **连续决策**：AI 能够基于之前的决策进行连续判断
2. **避免重复**：AI 能看到之前刚刚做过的决策，避免短时间内重复操作
3. **改进推理**：AI 能够参考之前的推理过程，优化当前决策
4. **趋势跟踪**：AI 能够基于历史形成对币种的整体判断

## 注意事项

1. **存储要求**：需要启用 store 才能使用历史记忆功能
2. **性能影响**：每个币种会额外查询数据库，建议限制候选币种数量
3. **Token 消耗**：历史消息会增加 AI 请求的 token 数量
4. **回测模式**：回测暂时不支持历史记忆（因为回测时没有真实的决策历史）

## 配置

历史轮数硬编码为 10 轮，如需修改可编辑 `kernel/engine.go`：

```go
historyRecords, err := ctx.DecisionStore.GetLatestRecordsBySymbol(ctx.TraderID, symbol, 10)
//                                                                              ^^ 修改这里
```

## 数据库表结构

`decision_records` 表已自动包含所需字段：

- `trader_id`: 交易者 ID（索引）
- `timestamp`: 时间戳（索引）
- `decisions`: JSON 格式的决策列表（用于币种过滤）
- `input_prompt`: 用户提示词
- `raw_response`: AI 原始响应

查询通过 `LIKE` 匹配 JSON 中的币种，性能可能较慢，如有需要可添加专门的币种字段和索引。
