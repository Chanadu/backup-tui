# backup-tui

A beautiful, interactive terminal user interface (TUI) for creating and uploading encrypted backups to remote servers. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [lipgloss](https://github.com/charmbracelet/lipgloss).

A TUI version of [backup-cli](https://github.com/Chanadu/backup-cli)

## Features

### 🎯 Stage-Based Workflow

The application guides you through a linear workflow with clear stages:

1. **Parameters** - Configure connection and backup settings
2. **Check Server** - Validate SSH connection and backup destination
3. **Select Files** - Choose which files to backup with search functionality
4. **Create Backups** - Compress files locally with 7-Zip
5. **Upload Backups** - Transfer backups to remote server via SFTP

### ⚙️ Parameters Stage

- **User**: SSH username for remote server connection
- **Server**: Remote server IP address or hostname
- **Password**: SSH password (never saved to disk)
- **Backup Path**: Destination path on remote server
- **Print Commands**: Toggle to show executed shell commands during backup creation
- **Show Progress**: Toggle to display real-time progress bars, spinners, and timing information during backup and upload operations

**Smart Defaults**: Previously entered parameters (except password) are automatically loaded from local storage on app start, saving you time on repeated backups.

### 🔐 Server Validation

- Establishes SSH connection to verify credentials
- Automatically creates the backup path on the remote server if it doesn't exist
- Provides immediate feedback on connection success or failure
- Option to retry if validation fails

### 📁 File Selection

- **Interactive File Browser**: Navigate your file system with arrow keys or vim-style keybindings
- **Search Functionality**: Press `/` to filter files in real-time with instant preview
- **Symlink Handling**: Intelligently displays symlinks while filtering prevents duplicate selections
- **File Actions**:
  - Press `Enter` to select/deselect files
  - Press `x` to remove a file from the selection
  - Search results use filtered views to avoid duplicate selections via different paths
- **Duplicate Prevention**: Uses OS-level file identity checks to prevent accidentally selecting the same file twice
- **Fixed-Height Display**: Optimized viewport shows multiple files while keeping UI stable
- **Help Footer**: Context-aware keyboard hints for all available actions

### 🗜️ Backup Creation

- **7-Zip Compression**: Uses industry-standard 7z format for compression
- **Real-Time Progress**: Live progress bar showing compression percentage with cyan-to-pink gradient (when Show Progress is enabled)
- **Elapsed Timer**: Displays time spent on current file backup (when Show Progress is enabled)
- **Command Display**: Optional printing of underlying shell commands (if Print Commands is enabled in parameters)
- **Per-File Tracking**: Always shows which file is currently being backed up `[N/Total]`
- **Resilient Processing**: Continues backup creation even if individual file compression completes with warnings
- **Flexible Display**: When Show Progress is disabled, displays only the filename without progress details for a minimal UI

### 📤 Upload to Remote Server

- **SFTP Upload**: Secure file transfer via SSH File Transfer Protocol
- **Live Progress Tracking**: Real-time progress bar for each file upload with gradient coloring (when Show Progress is enabled)
- **Speed Calculation**: Displays upload speed in MB/s for current operation (when Show Progress is enabled)
- **Elapsed Time**: Shows duration of current file upload (when Show Progress is enabled)
- **Per-File Monitoring**: Always tracks `[N/Total]` file uploads with current filename display
- **Summary Statistics**: Final timing table showing:
  - Files uploaded with individual timing
  - Backup creation durations
  - Total session time
- **Flexible Display**: When Show Progress is disabled, displays only the filename without progress details for a minimal UI

### 🎨 User Interface

- **Full-Width Responsive Container**: App UI adapts to terminal width with rounded borders
- **Color Scheme**:
  - **Primary (Purple)**: Focused elements and navigation indicators
  - **Secondary (Light Purple)**: Labels and prompts
  - **Info (Cyan)**: Current filenames and important information
  - **Success (Cyan)**: Toggled switches and successful states
  - **Muted (Gray)**: Disabled elements and secondary text
- **Gradient Progress Bars**: Cyan-to-pink animated progress visualization
- **Spinner Animation**: Loading indicators during long operations
- **Clean Typography**: Bold titles, properly spaced sections, and context-aware help text

### 💾 Local Configuration Storage

- **Platform-Aware Paths**:
  - **Linux/macOS**: `~/.config/backup-tui/`
  - **Windows**: `%APPDATA%\backup-tui\`
- **Persistent Settings**: Automatically saves:
  - SSH username
  - Server address
  - Backup path
  - Toggle preferences (Print Commands)
- **Password Security**: Password is never saved to disk for security
- **Auto-Loading**: On startup, previous parameters populate the input form for faster repeated backups
- **JSON Format**: Human-readable configuration file for manual editing if needed

### 📋 Logging

- **Platform-Aware Log Directory**:
  - **Linux/macOS**: `~/.config/backup-tui/log/`
  - **Windows**: `%APPDATA%\backup-tui\log\`
- **Timestamped Logs**: Each session creates a log file named `YYYY-MM-DD_HH-MM-SS.log` for tracking
- **Debug Mode**: Set `DEBUG` environment variable to log to `debug.log` for detailed troubleshooting
- **Automatic Directory Creation**: Log directories are created automatically on first run

### 🛡️ Error Handling

- **Graceful Degradation**: Application continues processing even if individual files fail
- **Detailed Error Reporting**: Error messages displayed with context about what failed
- **Retry Mechanisms**: Server validation stage allows retry on connection failure
- **Backup Path Creation**: Automatically creates remote backup directory if missing

### 🎮 Keyboard Controls

**Navigation**:
- `Tab` / `Shift+Tab`: Move between input fields
- `Up` / `Ctrl+K`: Navigate up
- `Down` / `Ctrl+J`: Navigate down
- `Enter`: Submit parameters or select files

**File Selection** (when in Files stage):
- `Enter`: Select/deselect current file
- `x`: Remove selected file
- `/`: Enter search mode to filter files
- Arrow keys: Navigate file list

**General**:
- `Space`: Toggle switches (e.g., Print Commands)
- `Ctrl+C`: Quit application at any time

### 📊 Performance Features

- **Background Commands**: 7z and SSH operations run asynchronously without blocking UI
- **Streaming Output**: Real-time progress updates from both compression and upload operations
- **Process Management**: Internal process tracking and cleanup on exit
- **Efficient File Handling**: Smart symlink and hard-link management during file selection filtering
- **Bounded Memory**: Fixed-height display prevents memory issues with large file lists

## Installation

### From Source

```bash
git clone https://github.com/Chanadu/backup-tui.git
cd backup-tui
go build -o bin/backup-tui
```

### Run

```bash
./bin/backup-tui
```

### Debug Mode

```bash
DEBUG=1 ./bin/backup-tui
```

Logs will be available in:
- **Linux/macOS**: `~/.config/backup-tui/log/`
- **Windows**: `%APPDATA%\backup-tui\log\`

## Requirements

- Go 1.25+
- 7-Zip (`7z` binary) for compression
- SSH access to remote server
- SFTP support on remote server

## Configuration

Configuration is stored in JSON format:

**Linux/macOS**: `~/.config/backup-tui/config.json`
**Windows**: `%APPDATA%\backup-tui\config.json`

Example:
```json
{
  "user": "pi",
  "server": "192.168.1.100",
  "backupPath": "/mnt/backups/",
  "debug": false,
  "commands": true,
  "progress": true
}
```

Note: Password is intentionally not stored in configuration for security.

## Architecture

The application uses Bubble Tea's Model-Update-View (MVU) pattern with separate models for each stage:

- `Stage`: Enum representing current workflow position
- `InputModel`: Handles parameter collection with text inputs and switches
- `CheckServerModel`: Validates SSH connection and backup destination
- `FileSelectorModel`: Interactive file browser with search
- `CreateBackupsModel`: Manages 7z compression process
- `UploadBackupsModel`: Handles SFTP upload with progress tracking
- `Config`: OS-aware storage and logging setup

Each model maintains its own state and handles UI rendering for its stage.

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/charmbracelet/bubbles` - Reusable UI components (progress, spinner, help, textinput)
- `github.com/pkg/sftp` - SFTP protocol implementation
- `golang.org/x/crypto/ssh` - SSH client
- `github.com/dustin/go-humanize` - Human-readable file sizes

## Troubleshooting

**Connection Issues**: Check your server address, username, and password. Use debug mode (`DEBUG=1`) for detailed connection logs.

**7z Not Found**: Ensure 7-Zip is installed and `7z` binary is in your PATH.

**Permission Errors**: Verify you have write permissions to the backup destination on the remote server.

**Log Location**: On first run, check the log directory (`.config/backup-tui/log/` or `%APPDATA%\backup-tui\log\`) for detailed error messages.

