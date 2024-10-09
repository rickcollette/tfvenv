
# tfvenv Installation Guide

Welcome to tfvenv! Follow this guide to install tfvenv on your operating system using your preferred method.

## Contents
- [Installation Methods](#installation-methods)
  - [Debian/Ubuntu (.deb)](#debianubuntu-deb)
  - [Red Hat/CentOS/Fedora (.rpm)](#red-hatcentosfedora-rpm)
  - [macOS (.pkg)](#macos-pkg)
  - [Windows (.exe via InstallForge)](#windows-exe-via-installforge)
  - [Linux/macOS (.tgz)](#linuxmacos-tgz)
- [Verify Installation](#verify-installation)
- [Post-Installation](#post-installation)
  - [Add to PATH](#add-to-path)
  - [Set Environment Variables](#set-environment-variables)
- [Troubleshooting](#troubleshooting)
- [Conclusion](#conclusion)

## 1. Installation Methods

### Debian/Ubuntu (.deb)
**Download .deb Package**:

Visit the tfvenv Releases Page and download the latest .deb file (e.g., `tfvenv_1.0.0_amd64.deb`).

**Install Package**:
```bash
sudo dpkg -i tfvenv_1.0.0_amd64.deb
```

**Fix Dependencies**:
```bash
sudo apt-get install -f
```

### Red Hat/CentOS/Fedora (.rpm)
**Download .rpm Package**:

Get the latest .rpm file from the tfvenv Releases Page (e.g., `tfvenv-1.0.0-1.x86_64.rpm`).

**Install Package**:
```bash
sudo rpm -ivh tfvenv-1.0.0-1.x86_64.rpm
```

Or using yum/dnf:
```bash
sudo yum install tfvenv-1.0.0-1.x86_64.rpm
# or
sudo dnf install tfvenv-1.0.0-1.x86_64.rpm
```

### macOS (.pkg)
**Download .pkg Installer**:

Download the latest .pkg from the tfvenv Releases Page (e.g., `tfvenv-1.0.0.pkg`).

**Run Installer**:

Double-click the `.pkg` file and follow the prompts.

### Windows (.exe via InstallForge)
**Download .exe Installer**:

Get the latest .exe from the tfvenv Releases Page (e.g., `tfvenv-1.0.0.exe`).

**Run Installer**:

Double-click the `.exe` file and follow the InstallForge prompts.

### Linux/macOS (.tgz)
**Download .tgz Archive**:

Download the appropriate `.tgz` from the tfvenv Releases Page (e.g., `tfvenv-1.0.0-linux-amd64.tgz`).

**Extract Archive**:
```bash
tar -xzvf tfvenv-1.0.0-linux-amd64.tgz
```

**Move Binary to PATH**:
- **System-wide**:
    ```bash
    sudo mv tfvenv /usr/local/bin/
    ```
- **User-specific**:
    ```bash
    mv tfvenv ~/bin/
    ```

**Make Executable**:
```bash
sudo chmod +x /usr/local/bin/tfvenv
```

## 2. Verify Installation

1. Open Terminal or Command Prompt.
2. Check tfvenv Version:
    ```bash
    tfvenv --version
    ```

**Expected Output**:
```
tfvenv version 1.0.0
```

If "command not found" appears, ensure tfvenv is in your PATH.

## 3. Post-Installation

### Add to PATH
If using `.tgz` and binary isn't in PATH:

- **Bash**:
    ```bash
    echo 'export PATH="$PATH:/path/to/tfvenv-directory"' >> ~/.bashrc
    source ~/.bashrc
    ```

- **Zsh**:
    ```bash
    echo 'export PATH="$PATH:/path/to/tfvenv-directory"' >> ~/.zshrc
    source ~/.zshrc
    ```

- **Windows**:

    The installer typically adds tfvenv to PATH. If not, add the installation directory to the PATH variable via System Properties.

### Set Environment Variables

**Locate Configuration File**:
- **Linux/macOS**: `~/.tfvenvrc`
- **Windows**: `%USERPROFILE%\.tfvenvrc`

**Edit Configuration**:
```makefile
TF_VERSION=1.0.0
TG_VERSION=0.35.0
S3_STATE_BUCKET=your_s3_state_bucket
S3_STATE_PATH=your_s3_state_path
REGION=your_aws_region
```

**Apply Changes**:

Reload shell or restart Terminal.

## 4. Troubleshooting

### Command Not Found:
- **Cause**: tfvenv not in PATH.
- **Solution**: Add tfvenv directory to PATH.

### Permission Denied:
- **Cause**: Executable permissions missing.
- **Solution**:
    ```bash
    sudo chmod +x /usr/local/bin/tfvenv
    ```

### Incorrect Version:
- **Cause**: Wrong version downloaded or installed.
- **Solution**: Re-download the correct version from Releases Page.

### Configuration Errors:
- **Cause**: Misconfigured `.tfvenvrc`.
- **Solution**: Review and correct settings in the configuration file.

### Dependency Issues:
- **Cause**: Missing dependencies.
- **Solution**: Ensure all dependencies are installed. Package managers typically handle this automatically.

### Check Logs:
- **Location**:
  - **Linux/macOS**: `~/.tfvenv/logs/`
  - **Windows**: `%USERPROFILE%\.tfvenv\logs\`

### Seek Help:
Open an issue on the tfvenv GitHub Repository with detailed information.
