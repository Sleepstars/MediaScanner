# MediaScanner

MediaScanner is a high-efficiency media information scraper powered by Large Language Models (LLMs). It automatically processes media files, extracts information, and organizes them according to user preferences.

## Features

- **LLM-Powered Analysis**: Uses LLMs to accurately identify media from filenames, even with complex or non-standard naming.
- **Multiple API Integration**: Integrates with TMDB, TVDB, and Bangumi APIs for comprehensive media information.
- **Batch Processing**: Efficiently processes directories with multiple related files.
- **Flexible Organization**: Customizable directory structure and naming templates.
- **Metadata Generation**: Creates NFO files and downloads images for media servers like Emby/Plex.
- **Notification System**: Sends notifications via Telegram for successful processing and errors.

## Requirements

- Go 1.18 or higher
- PostgreSQL database
- OpenAI API key or compatible LLM API
- TMDB, TVDB, and Bangumi API keys

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/sleepstars/mediascanner.git
   cd mediascanner
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Build the application:
   ```
   go build -o mediascanner cmd/mediascanner/main.go
   ```

## Configuration

Copy the example configuration file and modify it according to your needs:

```
cp config.example.yaml config.yaml
```

### Configuration Options

- **General Settings**: Log level, scan interval
- **LLM Settings**: API key, model, etc.
- **API Settings**: TMDB, TVDB, and Bangumi API keys
- **Database Settings**: PostgreSQL connection details
- **Scanner Settings**: Media directories, exclusion patterns, etc.
- **File Operations**: File handling mode (copy/move/symlink), destination structure
- **Notification Settings**: Telegram bot token and channel/group IDs

## Usage

Run MediaScanner with a configuration file:

```
./mediascanner -config config.yaml
```

Or use environment variables (see the configuration file for available options):

```
export LLM_API_KEY="your-openai-api-key"
export TMDB_API_KEY="your-tmdb-api-key"
./mediascanner
```

## How It Works

1. **Scanning**: MediaScanner periodically scans configured directories for new media files.
2. **Analysis**: Files are analyzed by the LLM to identify the media title, type, and other information.
3. **API Integration**: The LLM uses Function Calling to query TMDB, TVDB, and Bangumi APIs for accurate information.
4. **Processing**: Files are organized according to the configured directory structure and naming templates.
5. **Metadata**: NFO files and images are generated for media servers.
6. **Notification**: Success and error notifications are sent via Telegram.

## Directory Structure

The organized media library follows this structure:

### Movies
```
/Movies/Category/Title (Year)/Title (Year).ext
```

### TV Shows
```
/TV Shows/Category/Title (Year)/Season X/Title - SXXEXX - Episode Title.ext
```

## Acknowledgements

This project makes use of the following data sources and open-source libraries:

### Data Sources

- [TMDB (The Movie Database)](https://www.themoviedb.org/) - An open database for movie and TV show information
- [TVDB](https://thetvdb.com/) - A community-driven database of television shows
- [Bangumi](https://bgm.tv/) - A database for anime, manga, and other ACG content

### Libraries

- [go-openai](https://github.com/sashabaranov/go-openai) - Go client library for OpenAI API
- [golang-tmdb](https://github.com/cyruzin/golang-tmdb) - Go client library for TMDB API
- [gorm](https://github.com/go-gorm/gorm) - The fantastic ORM library for Go
- [echo](https://github.com/labstack/echo) - High performance web framework for Go
- [godotenv](https://github.com/joho/godotenv) - Go library for loading environment variables from .env files

## Docker Support

MediaScanner provides full Docker support with multi-architecture images (amd64 and arm64).

### Using Docker

```bash
# Pull the image
docker pull ghcr.io/sleepstars/mediascanner:latest

# Run with environment variables
docker run -d \
  -v /path/to/config.yaml:/root/config.yaml \
  -v /path/to/media/input:/media/input \
  -v /path/to/media/output:/media/output \
  -e LLM_API_KEY=your-openai-api-key \
  -e TMDB_API_KEY=your-tmdb-api-key \
  -e TVDB_API_KEY=your-tvdb-api-key \
  -e BANGUMI_API_KEY=your-bangumi-api-key \
  ghcr.io/sleepstars/mediascanner:latest
```

### Using Docker Compose

1. Edit the `docker-compose.yml` file to configure your media paths
2. Create a `.env` file with your API keys
3. Start the services:

```bash
docker-compose up -d
```

## CI/CD

This project uses GitHub Actions for continuous integration and delivery:

- **CI Workflow**: Automatically tests and builds the application on each push and pull request
- **Docker Workflow**: Builds and publishes multi-architecture Docker images (amd64 and arm64) to GitHub Container Registry (ghcr.io) using native runners and Noelware's docker-manifest-action
- **Release Workflow**: Automatically creates releases with binaries for multiple platforms when a tag is pushed

## Special Notes

- Bangumi API usage follows their [User-Agent requirements](https://github.com/bangumi/api/blob/master/docs-raw/user%20agent.md), using `sleepstars/MediaScanner (https://github.com/sleepstars/MediaScanner)` as the default User-Agent.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
