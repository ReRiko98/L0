package check

import (
	"fmt"
	"log"
	"myapp/internal/main"
	"net/http"
	"contex"
)

func check(ctx context.Context, msg <-chan[]byte, conn *pgx.Conn) {
	logger := log.LoggerFromContext(ctx)
	logger.Info("checking messages")
}

go func() {
    for {
      select {
      case msg := <-msg:
        msgStruct := &validator.Order{}
        if err := json.Unmarshal(msg, msgStruct); err != nil {
          logger.Error("JSON bad data:", zap.Error(err))
          break
        }

        if err := msgStruct.Validate(ctx); err != nil {
          logger.Error("Validation error:", zap.Error(err))
          break
        }

        if err := postgres.Create(ctx, conn, msgStruct); err != nil {
          logger.Error("postgres.Create() error")
          break
        }
        cache.CACHE[msgStruct.OrderUID] = msgStruct
      case <-ctx.Done():
        logger.Info("Stop listening messages...")
        return
      }
    }
  }()