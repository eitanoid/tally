package logger

import (
	"fmt"
	"log/slog"
)

func init() {
	fmt.Println(slog.Logger{})

}

type Logger struct {
}
