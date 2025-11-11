# GoNotify – Element Notifications

GoNotify is a standalone Go package that sends rich trading notifications to an Element (Matrix) chat room. It ships with a lightweight event bus, JSON configuration helpers, and a high-level service that wires everything together.

## Features

- Send notifications for important trading events:
  - Trade execution
  - Order filled (including stop losses and take profits)
  - Position opened/closed
  - Profit and loss updates
  - System errors
  - Strategy errors
- Configurable notification settings
- Integration with Element decentralized messenger

## Configuration

The notification system is configured using a JSON file. An example configuration file is provided at `configs/notification.json`.

```json
{
  "element": {
    "homeserver_url": "https://matrix.org",
    "access_token": "YOUR_ELEMENT_ACCESS_TOKEN",
    "room_id": "!cryptobot:matrix.org"
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

## Getting an Element Access Token

To use the notification system, you need an Element access token:

1. Create an account on an Element homeserver (e.g., https://app.element.io/)
2. Create a new room or use an existing one
3. Get your access token:
   - Open Element in a web browser
   - Open Developer Tools (F12)
   - Go to Application > Local Storage
   - Find the `mx_access_token` value
4. Update the configuration file with your access token and room ID

## Usage

### Basic Usage

```go
import notification "github.com/evdnx/gonotify"

eventBus := notification.NewEventBus()

config, err := notification.LoadConfig("configs/notification.json")
if err != nil {
    panic(err)
}

service, err := notification.NewNotificationService(config, eventBus)
if err != nil {
    panic(err)
}

if err := service.Start(); err != nil {
    panic(err)
}

eventBus.PublishData(notification.EventTradeExecuted, map[string]interface{}{
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

You can also set the Element access token using an environment variable:

```bash
export ELEMENT_ACCESS_TOKEN="your_access_token"
```

### Event Types

The built-in event bus ships with predefined event identifiers that drive the notification service:

- `EventTradeExecuted`
- `EventOrderFilled`
- `EventPositionOpened`
- `EventPositionClosed`
- `EventPnLUpdate`
- `EventSystemError`
- `EventStrategyError`

Publish any of these events (or your own custom ones) to the bus and the service will deliver the corresponding Element message.

## Integration

Use `InitializeNotificationSystem` to bootstrap the service from a config file:

```go
eventBus := notification.NewEventBus()
service, err := notification.InitializeNotificationSystem(eventBus, "configs/notification.json")
if err != nil {
    panic(err)
}
```

The helper ensures the configuration file exists, loads it, optionally reads `ELEMENT_ACCESS_TOKEN`, and starts the notification service.
