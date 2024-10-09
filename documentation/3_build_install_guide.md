
# tfvenv Build and Install Guide

Follow this guide to build and install tfvenv from source using Go.

## Contents
- [Prerequisites](#prerequisites)
- [Build and Install](#build-and-install)
  - [Clone Repository](#clone-repository)
  - [Build Binary](#build-binary)
  - [Install Binary](#install-binary)
- [Configure Environment Variables](#configure-environment-variables)
- [Verify Installation](#verify-installation)
- [Post-Installation Steps](#post-installation-steps)
- [Troubleshooting](#troubleshooting)

## 1. Prerequisites

Ensure your system meets the following:
- **Operating System**: Linux, macOS, or Windows.
- **Go**: Version 1.16 or higher.
    ```bash
    go version
    # Example Output:
    # go version go1.20.3 linux/amd64
    ```
- **Git**: Installed on your system.

## 2. Build and Install

### Clone Repository
- Open Terminal or Command Prompt.
- Navigate to your projects directory:
    ```bash
    cd ~/projects
    ```
- Clone the tfvenv repository:
    ```bash
    git clone https://github.com/rickcollette/tfvenv.git
    ```

### Build Binary
- Navigate to the tfvenv directory:
    ```bash
    cd tfvenv
    ```
- Build the tfvenv binary:
    ```bash
    go build -o tfvenv main.go
    ```
  - `-o tfvenv`: Names the output binary as tfvenv.
  - `main.go`: Main source file.

### Install Binary

- **Linux/macOS**:
  - Move the binary to a directory in your PATH:
    ```bash
    sudo mv tfvenv /usr/local/bin/
    ```
  - Make it executable:
    ```bash
    sudo chmod +x /usr/local/bin/tfvenv
    ```

- **Windows**:
  - Move `tfvenv.exe` to a directory in your PATH, e.g., `C:\\Program Files\\tfvenv\\`.
  - Ensure the directory is added to the PATH environment variable:
    - Go to System Properties > Environment Variables.
    - Edit the PATH variable to include `C:\\Program Files\\tfvenv\\`.

## 3. Configure Environment Variables

Set necessary environment variables for tfvenv.

### Set `TFVENV_PATH`
- **Linux/macOS**:
  - Open shell config file:
    ```bash
    nano ~/.bashrc
    # or for Zsh
    nano ~/.zshrc
    ```
  - Add:
    ```bash
    export TFVENV_PATH="$HOME/.tfvenv"
    ```
  - Apply changes:
    ```bash
    source ~/.bashrc
    # or
    source ~/.zshrc
    ```

- **Windows**:
  - Open System Properties > Environment Variables.
  - Add a new user variable:
    - **Name**: `TFVENV_PATH`
    - **Value**: `C:\\Users\\YourUsername\\.tfvenv`

### Set `TFVENV_LOG_FILE` (Optional)
- **Linux/macOS**:
  - Add to shell config:
    ```bash
    export TFVENV_LOG_FILE="$HOME/.tfvenv/tfvenv.log"
    ```
  - Apply changes:
    ```bash
    source ~/.bashrc
    # or
    source ~/.zshrc
    ```

- **Windows**:
  - Add a new user variable:
    - **Name**: `TFVENV_LOG_FILE`
    - **Value**: `C:\\Users\\YourUsername\\.tfvenv\\tfvenv.log`

## 4. Verify Installation
- Open Terminal or Command Prompt.
- Check tfvenv version:
    ```bash
    tfvenv --version
    ```

**Expected Output**:
```
tfvenv version 1.0.0
```

If "command not found," ensure the binary is in your PATH and has execute permissions.

## 5. Post-Installation Steps

### Add tfvenv to PATH (If Not Done)

- **Linux/macOS**:
  - Edit shell config:
    ```bash
    nano ~/.bashrc
    # or
    nano ~/.zshrc
    ```
  - Add:
    ```bash
    export PATH="$PATH:$TFVENV_PATH/bin"
    ```
  - Apply changes:
    ```bash
    source ~/.bashrc
    # or
    source ~/.zshrc
    ```

- **Windows**:
  - The installer typically adds tfvenv to PATH. If not, manually add the installation directory via System Properties > Environment Variables.

### Set Additional Environment Variables

- **Linux/macOS**:
  - Edit shell config:
    ```bash
    nano ~/.bashrc
    # or
    nano ~/.zshrc
    ```
  - Add:
    ```bash
    export TF_PLUGIN_CACHE_DIR="$TFVENV_PATH/plugin-cache"
    export TF_DATA_DIR="$TFVENV_PATH/terraform-data"
    ```
  - Apply changes:
    ```bash
    source ~/.bashrc
    # or
    source ~/.zshrc
    ```

- **Windows**:
  - Open System Properties > Environment Variables.
  - Add or edit user variables:
    - **Name**: `TF_PLUGIN_CACHE_DIR`  
      **Value**: `C:\\Users\\YourUsername\\.tfvenv\\plugin-cache`
    - **Name**: `TF_DATA_DIR`  
      **Value**: `C:\\Users\\YourUsername\\.tfvenv\\terraform-data`

## 6. Troubleshooting

### Command Not Found:
- **Cause**: tfvenv not in PATH.
- **Solution**: Ensure tfvenv binary is in a PATH directory and has execute permissions.
    ```bash
    sudo chmod +x /usr/local/bin/tfvenv
    ```
- Restart Terminal or Command Prompt.

### Permission Denied:
- **Cause**: Insufficient permissions to execute tfvenv.
- **Solution**: Set executable permissions.
    ```bash
    sudo chmod +x /usr/local/bin/tfvenv
    ```

### Incorrect Version Displayed:
- **Cause**: Old tfvenv binary in PATH.
- **Solution**: Verify binary location.
    ```bash
    which tfvenv
    # or on Windows
    where tfvenv
    ```
- Remove or update older versions.

### Environment Variables Issues:
- **Cause**: Missing or incorrect variables.
- **Solution**: Check and update `.tfvenvrc` and shell config files.
    ```bash
    source ~/.bashrc
    # or
    source ~/.zshrc
    ```

### Issues with Go Build:
- **Cause**: Go not installed or outdated.
- **Solution**: Install or update Go from the official website.

### Network Issues:
- **Cause**: Problems accessing GitHub.
- **Solution**: Check internet connection and retry cloning:
    ```bash
    git clone https://github.com/rickcollette/tfvenv.git
    ```

### Check Logs:
- **Location**:
  - **Linux/macOS**: `~/.tfvenv/logs/`
  - **Windows**: `%USERPROFILE%\\.tfvenv\\logs\\`

- **View Logs**:
    ```bash
    cat ~/.tfvenv/logs/tfvenv.log
    ```

