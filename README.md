# GoNotify – Multi-Platform Trading Notifications

GoNotify is a standalone Go package that sends rich trading notifications to multiple messaging platforms (Element/Matrix and Telegram). It ships with a lightweight event bus, JSON configuration helpers, and a high-level service that wires everything together.

## Features

- **Multi-platform support**: Send notifications to Element (Matrix) and/or Telegram
- **Configurable notifications** for important trading events:
  - Trade execution
  - Order filled (including stop losses and take profits)
  - Position opened/closed
  - Profit and loss updates
  - System errors
  - Strategy errors
- **Flexible configuration**: Enable/disable specific messengers and event types
- **Event-driven architecture**: Lightweight pub/sub event bus

## Installation

```bash
go get github.com/evdnx/gonotify
```

## Configuration

The notification system is configured using a JSON file. An example configuration file is provided at `configs/notification.json`.

```json
{
  "element": {
    "homeserver_url": "https://matrix.org",
    "access_token": "YOUR_ELEMENT_ACCESS_TOKEN",
    "room_id": "!cryptobot:matrix.org",
    "enabled": true
  },
  "telegram": {
    "bot_token": "YOUR_TELEGRAM_BOT_TOKEN",
    "chat_id": "YOUR_CHAT_ID",
    "enabled": true
  },
  "events": {
    "trade_execution": true,
    "order_filled": true,
    "position_change": true,
    "pnl_update": true,
    "stop_loss": true,
    "take_profit": true,
    "system_errors": true,
    "strategy_errors": true,
    "profit_threshold": 1.0
  }
}
```

### Element Configuration

- `homeserver_url`: URL of the Element homeserver (e.g., "https://matrix.org")
- `access_token`: Your Element access token (required for authentication)
- `room_id`: ID of the chat room to send notifications to (e.g., "!cryptobot:matrix.org")
- `enabled`: Enable or disable Element notifications

### Telegram Configuration

- `bot_token`: Your Telegram bot token (obtained from @BotFather)
- `chat_id`: The chat ID where notifications will be sent
- `enabled`: Enable or disable Telegram notifications

### Event Configuration

- `trade_execution`: Send notifications for trade executions
- `order_filled`: Send notifications for filled orders
- `position_change`: Send notifications for position changes (open/close)
- `pnl_update`: Send notifications for profit and loss updates
- `stop_loss`: Send notifications for stop loss orders
- `take_profit`: Send notifications for take profit orders
- `system_errors`: Send notifications for system errors
- `strategy_errors`: Send notifications for strategy errors
- `profit_threshold`: Minimum profit/loss percentage to trigger a notification (e.g., 1.0 for 1%)

## Getting Credentials

### Element Access Token

1. Create an account on an Element homeserver (e.g., https://app.element.io/)
2. Create a new room or use an existing one
3. Get your access token:
   - Open Element in a web browser
   - Open Developer Tools (F12)
   - Go to Application > Local Storage
   - Find the `mx_access_token` value
4. Update the configuration file with your access token and room ID

### Telegram Bot Token and Chat ID

1. Create a bot by messaging [@BotFather](https://t.me/botfather) on Telegram
2. Use `/newbot` command and follow the instructions
3. Save the bot token provided by BotFather
4. Get your chat ID:
   - Start a chat with your bot
   - Send a message to your bot
   - Visit `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
   - Find your chat ID in the response (it's a number, may be negative for groups)
5. Update the configuration file with your bot token and chat ID

## Usage

### Basic Usage

```go
import (
    "github.com/evdnx/gonotify/eventbus"
    "github.com/evdnx/gonotify/config"
    "github.com/evdnx/gonotify/service"
)

eventBus := eventbus.NewEventBus()

cfg, err := config.LoadConfig("configs/notification.json")
if err != nil {
    panic(err)
}

svc, err := service.NewNotificationService(cfg, eventBus)
if err != nil {
    panic(err)
}

if err := svc.Start(); err != nil {
    panic(err)
}

eventBus.PublishData(eventbus.EventTradeExecuted, map[string]interface{}{
    "id":         "trade-123",
    "symbol":     "BTCUSDT",
    "side":       "buy",
    "quantity":   0.1,
    "price":      68000.0,
    "base_asset": "BTC",
    "quote_asset": "USDT",
})
```

### Environment Variables

You can also set credentials using environment variables:

```bash
# Element
export ELEMENT_ACCESS_TOKEN="your_access_token"

# Telegram
export TELEGRAM_BOT_TOKEN="your_bot_token"
export TELEGRAM_CHAT_ID="your_chat_id"
```

### Using Multiple Messengers

The service supports sending notifications to multiple messengers simultaneously. Simply enable both in your configuration:

```json
{
  "element": {
    "enabled": true,
    ...
  },
  "telegram": {
    "enabled": true,
    ...
  }
}
```

### Custom Messenger Implementation

You can also provide custom messenger implementations:

```go
import "github.com/evdnx/gonotify/messenger"

type CustomMessenger struct {
    // your fields
}

func (m *CustomMessenger) SendMessage(message string) error {
    // your implementation
    return nil
}

func (m *CustomMessenger) Name() string {
    return "Custom"
}

// Use it
messengers := []messenger.Messenger{
    &CustomMessenger{},
}
svc, err := service.NewNotificationServiceWithMessengers(cfg, eventBus, messengers)
```

### Event Types

The built-in event bus ships with predefined event identifiers:

- `eventbus.EventTradeExecuted`
- `eventbus.EventOrderFilled`
- `eventbus.EventPositionOpened`
- `eventbus.EventPositionClosed`
- `eventbus.EventPnLUpdate`
- `eventbus.EventSystemError`
- `eventbus.EventStrategyError`

Publish any of these events (or your own custom ones) to the bus and the service will deliver the corresponding message to all enabled messengers.

## Integration

Use `InitializeNotificationSystem` to bootstrap the service from a config file:

```go
import (
    "github.com/evdnx/gonotify"
    "github.com/evdnx/gonotify/eventbus"
)

eventBus := eventbus.NewEventBus()
service, err := gonotify.InitializeNotificationSystem(eventBus, "configs/notification.json", "", "")
if err != nil {
    panic(err)
}
```

You can also provide Telegram credentials as function parameters if they're not in the config file:

```go
// If JSON config doesn't have telegram configured, use function parameters
service, err := gonotify.InitializeNotificationSystem(
    eventBus, 
    "configs/notification.json",
    "your_telegram_bot_token",
    "your_telegram_chat_id",
)
```

The helper ensures the configuration file exists, loads it, optionally reads environment variables, and starts the notification service. Function parameters for Telegram credentials are used only if Telegram is not already enabled in the config file or environment variables.

## Package Structure

The library is organized into the following packages:

- `eventbus`: Event bus for pub/sub messaging
- `messenger`: Messenger interface and implementations
  - `messenger/element`: Element (Matrix) messenger client
  - `messenger/telegram`: Telegram messenger client
- `config`: Configuration loading and management
- `service`: Notification service that handles events and sends messages
- `types`: Shared type definitions

## License

See LICENSE file for details.
