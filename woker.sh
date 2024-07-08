#!/bin/bash

# 获取脚本所在的目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 定义执行的命令
CMD="$SCRIPT_DIR/main -o $SCRIPT_DIR/today.php -i $SCRIPT_DIR/template.html"

# 定义日志文件
LOG_FILE="$SCRIPT_DIR/log/cron.log"

# 检查日志文件所在的目录是否存在，不存在则创建
LOG_DIR="$(dirname "$LOG_FILE")"
if [ ! -d "$LOG_DIR" ]; then
    mkdir -p "$LOG_DIR"
fi

# 检查日志文件是否存在，不存在则创建
if [ ! -f "$LOG_FILE" ]; then
    touch "$LOG_FILE"
fi

# 最大重试次数
MAX_RETRIES=3

# 当前重试次数
retry_count=0

# 循环执行命令直到返回值为 0 或达到最大重试次数
until $CMD >> $LOG_FILE 2>&1; do
    retry_count=$((retry_count + 1))
    if [ $retry_count -ge $MAX_RETRIES ]; then
        echo "Command failed after $retry_count attempts. Giving up." >> $LOG_FILE
        exit 1
    fi
    echo "Command failed. Retrying ($retry_count/$MAX_RETRIES)..." >> $LOG_FILE
    sleep 5 # 等待5秒后重试
done

echo "Command succeeded." >> $LOG_FILE
