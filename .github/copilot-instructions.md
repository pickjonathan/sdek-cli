# sdek-cli Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-11

## Active Technologies
- Go 1.23+ (latest stable) (001-create-sdek)
- Go 1.23+ (latest stable, per existing project) (002-ai-evidence-analysis)
- Go 1.23+ (per existing project standard) (003-ai-context-injection)
- Go 1.23+ (per existing project standard) + Cobra (CLI), Viper (config), Bubble Tea + Lip Gloss (TUI), fsnotify (file watching), JSON Schema validator library (004-mcp-native-agent)
- File system (JSON configs in `~/.sdek/mcp/`, `./.sdek/mcp/`, and `$SDEK_MCP_PATH`); existing state management via internal/store (004-mcp-native-agent)

## Project Structure
```
src/
tests/
```

## Commands
# Add commands for Go 1.23+ (latest stable)

## Code Style
Go 1.23+ (latest stable): Follow standard conventions

## Recent Changes
- 004-mcp-native-agent: Added Go 1.23+ (per existing project standard) + Cobra (CLI), Viper (config), Bubble Tea + Lip Gloss (TUI), fsnotify (file watching), JSON Schema validator library
- 003-ai-context-injection: Added Go 1.23+ (per existing project standard)
- 002-ai-evidence-analysis: Added Go 1.23+ (latest stable, per existing project)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
