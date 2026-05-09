# LLM Proxy 🚀

一个轻量级、高性能的 LLM API 代理网关，专为个人和团队设计。将 OpenAI、Anthropic (Claude) 和 Google Gemini 聚合为统一的 OpenAI 兼容接口。

## ✨ 主要特性

- **统一 API 接口**: 完全兼容 OpenAI 格式的 `/v1/chat/completions`。
- **多模型支持**: 内置 OpenAI, Anthropic, Gemini 适配器。
- **内置管理后台**: 现代化的 Vue 3 仪表盘，支持实时日志、渠道管理和令牌分发。
- **静态资源合一**: 前后端合并，单二进制文件部署，访问 8080 端口即可使用全部功能。
- **自动文档**: 内置 Swagger 交互式文档 (`/swagger/index.html`)。
- **安全性**: 
  - 数据库密钥 **静态加密** (AES-256-GCM)。
  - 管理后台 JWT 认证。
  - 客户端 `sk-xxx` 令牌校验。
- **多架构支持**: 提供 Docker 镜像，支持 x64 和 ARM64 (适用于树莓派等设备)。

## 🚀 快速开始

### 使用 Docker (推荐)

```bash
docker run -d \
  --name llm-proxy \
  -p 8080:8080 \
  -v ./data:/app/data \
  -e DB_ENCRYPTION_KEY="你的强加密随机密钥" \
  ghcr.io/你的用户名/llm-proxy:main
```

### 本地编译

1. **安装依赖**: 确保安装了 Go 1.26+ 和 Node.js 24+。
2. **一键构建**:
   ```bash
   make build
   ```
3. **运行**:
   ```bash
   ./bin/llm-proxy
   ```

## 🔐 安全建议 (重要!)

> [!CAUTION]
> **严禁直接将 8080 端口暴露在公网！**
> 
> 为了保障数据安全和 API 密钥不被泄露，请务必配合 **反向代理 (Nginx / Caddy / Traefik)** 并配置 **SSL 证书 (HTTPS)** 使用。

**示例 Nginx 配置:**
```nginx
server {
    listen 443 ssl;
    server_name proxy.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## ⚙️ 环境变量

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| `PORT` | 服务运行端口 | `8080` |
| `DB_PATH` | 数据库文件路径 | `/app/data/data.db` |
| `DB_ENCRYPTION_KEY` | 数据库密钥加密 Key (建议 32 位) | - |
| `LOG_LEVEL` | 日志级别 (debug/info/warn/error) | `info` |

## 🛠 开发

- `make dev`: 启动开发模式（支持前端热重载 HMR）。
- `make doc`: 重新生成 Swagger 文档。

## 📄 License
MIT
