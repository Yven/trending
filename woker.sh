#!/bin/bash

# 获取脚本所在的目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# 设置默认的输出路径
OUTPUT_PATH="$SCRIPT_DIR/today.html"
# 设置默认日志输出路径
LOG_PATH="$SCRIPT_DIR/log/cron.log"

# 解析参数，获取输出路径
while [ $# -gt 0 ]; do
  case "$1" in
    -o|--output)
      OUTPUT_PATH="$2"
      shift 2
      ;;
    -l|--log)
      LOG_PATH="$2"
      shift 2
      ;;
    *)
      echo "Unknown Parameter: $1"
      exit 1
      ;;
  esac
done

# 检查路径是否存在以及是否可写
OUTPUT_DIR=$(dirname "$OUTPUT_PATH")
if [ ! -d "$OUTPUT_DIR" ]; then
  echo "The path does not exist: $OUTPUT_DIR"
  exit 1
elif [ ! -w "$OUTPUT_DIR" ]; then
  echo "Insufficient permissions: $OUTPUT_DIR"
  exit 1
fi

# 检查日志文件的父目录是否存在以及是否可写
LOG_DIR=$(dirname "$LOG_PATH")
if [ ! -d "$LOG_DIR" ]; then
  echo "The path does not exist: $LOG_DIR"
  exit 1
elif [ ! -w "$LOG_DIR" ]; then
  echo "Insufficient permissions: $LOG_DIR"
  exit 1
fi

# 检查日志文件是否存在，不存在则创建
if [ ! -f "$LOG_PATH" ]; then
    touch "$LOG_PATH"
fi

# 定义执行的命令
CMD="$SCRIPT_DIR/main -o $OUTPUT_PATH -i $SCRIPT_DIR/template.html"

# 最大重试次数
MAX_RETRIES=3
# 当前重试次数
retry_count=0
# 获取当前时间
current_time=$(date)

# 编译
go build $SCRIPT_DIR/main.go

# 循环执行命令直到返回值为 0 或达到最大重试次数
until $CMD >> $LOG_PATH 2>&1; do
    retry_count=$((retry_count + 1))
    if [ $retry_count -ge $MAX_RETRIES ]; then
        echo "[$current_time]Command failed after $retry_count attempts. Giving up." >> $LOG_PATH
        exit 1
    fi
    echo "[$current_time]Command failed. Retrying ($retry_count/$MAX_RETRIES)..." >> $LOG_PATH
    sleep 2 # 等待2秒后重试
done

echo "Command succeeded." >> $LOG_PATH
