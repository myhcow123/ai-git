#!/bin/bash

echo "Setting up Go module proxy for China..."

# 设置 Go 模块代理（中国区镜像）
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=off

# 持久化配置（写入到 shell 配置文件）
SHELL_CONFIG=""
if [ -f "$HOME/.bashrc" ]; then
    SHELL_CONFIG="$HOME/.bashrc"
elif [ -f "$HOME/.zshrc" ]; then
    SHELL_CONFIG="$HOME/.zshrc"
fi

if [ -n "$SHELL_CONFIG" ]; then
    echo "" >> "$SHELL_CONFIG"
    echo "# Go module proxy for China" >> "$SHELL_CONFIG"
    echo "export GOPROXY=https://goproxy.cn,direct" >> "$SHELL_CONFIG"
    echo "export GOSUMDB=off" >> "$SHELL_CONFIG"
    echo "Configuration added to $SHELL_CONFIG"
    echo "Please run: source $SHELL_CONFIG"
fi

echo ""
echo "Current Go environment:"
echo "GOPROXY: $GOPROXY"
echo "GOSUMDB: $GOSUMDB"
echo ""
echo "Setup complete! You can now run: go mod download"
