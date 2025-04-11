# MediaScanner

MediaScanner 是一个由大型语言模型（LLM）驱动的高效媒体信息刮削器。它能自动处理媒体文件，提取信息，并根据用户偏好进行组织。

## 功能特点

- **LLM 驱动分析**：利用大型语言模型准确识别复杂或非标准命名的媒体文件。
- **多 API 集成**：集成 TMDB、TVDB 和 Bangumi API，获取全面的媒体信息。
- **批量处理**：高效处理包含多个相关文件的目录。
- **灵活组织**：可自定义目录结构和命名模板。
- **元数据生成**：为 Emby/Plex 等媒体服务器创建 NFO 文件并下载图片。
- **通知系统**：通过 Telegram 发送处理成功和错误通知。

## 系统要求

- Go 1.18 或更高版本
- PostgreSQL 数据库
- OpenAI API 密钥或兼容的 LLM API
- TMDB、TVDB 和 Bangumi API 密钥

## 安装

1. 克隆仓库：
   ```
   git clone https://github.com/sleepstars/mediascanner.git
   cd mediascanner
   ```

2. 安装依赖：
   ```
   go mod tidy
   ```

3. 构建应用：
   ```
   go build -o mediascanner cmd/mediascanner/main.go
   ```

## 配置

复制示例配置文件并根据需要修改：

```
cp config.example.yaml config.yaml
```

### 配置选项

- **通用设置**：日志级别、扫描间隔
- **LLM 设置**：API 密钥、模型等
- **API 设置**：TMDB、TVDB 和 Bangumi API 密钥
- **数据库设置**：PostgreSQL 连接详情
- **扫描器设置**：媒体目录、排除模式等
- **文件操作**：文件处理模式（复制/移动/软链接）、目标结构
- **通知设置**：Telegram 机器人令牌和频道/群组 ID

## 使用方法

使用配置文件运行 MediaScanner：

```
./mediascanner -config config.yaml
```

或使用环境变量（参见配置文件了解可用选项）：

```
export LLM_API_KEY="your-openai-api-key"
export TMDB_API_KEY="your-tmdb-api-key"
./mediascanner
```

## 工作原理

1. **扫描**：MediaScanner 定期扫描配置的目录，查找新的媒体文件。
2. **分析**：LLM 分析文件以识别媒体标题、类型和其他信息。
3. **API 集成**：LLM 使用函数调用查询 TMDB、TVDB 和 Bangumi API 获取准确信息。
4. **处理**：根据配置的目录结构和命名模板组织文件。
5. **元数据**：为媒体服务器生成 NFO 文件和图片。
6. **通知**：通过 Telegram 发送成功和错误通知。

## 目录结构

组织后的媒体库遵循以下结构：

### 电影
```
/电影/分类/标题 (年份)/标题 (年份).扩展名
```

### 电视剧
```
/电视剧/分类/标题 (年份)/Season X/标题 - SXXEXX - 剧集标题.扩展名
```

## 特别说明

- Bangumi API 使用遵循其 [User-Agent 要求](https://github.com/bangumi/api/blob/master/docs-raw/user%20agent.md)，默认使用 `sleepstars/MediaScanner (https://github.com/sleepstars/MediaScanner)` 作为 User-Agent。

## 鸣谢

本项目使用了以下数据源和开源库，在此表示感谢：

### 数据源

- [TMDB (The Movie Database)](https://www.themoviedb.org/) - 提供电影和电视剧信息的开放数据库
- [TVDB](https://thetvdb.com/) - 提供电视剧信息的社区驱动数据库
- [Bangumi](https://bgm.tv/) - 提供动画、漫画等ACG内容信息的数据库

### 开源库

- [go-openai](https://github.com/sashabaranov/go-openai) - OpenAI API的Go客户端库
- [golang-tmdb](https://github.com/cyruzin/golang-tmdb) - TMDB API的Go客户端库
- [gorm](https://github.com/go-gorm/gorm) - Go语言的ORM库
- [echo](https://github.com/labstack/echo) - Go语言的高性能Web框架
- [godotenv](https://github.com/joho/godotenv) - 从.env文件加载环境变量的Go库

## 许可证

本项目采用 MIT 许可证 - 详情请参阅 LICENSE 文件。
