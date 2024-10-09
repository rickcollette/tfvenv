
# Upgrading tfvenv

Keep tfvenv updated to access the latest features, improvements, and security fixes. This guide outlines the steps to upgrade tfvenv efficiently on various operating systems and within existing environments.

## Contents
- [Pre-Upgrade Steps](#pre-upgrade-steps)
  - [Backup Configurations](#backup-configurations)
  - [Check Current Version](#check-current-version)
  - [Review Release Notes](#review-release-notes)
- [Upgrade Methods](#upgrade-methods)
  - [Using Package Managers](#using-package-managers)
    - [Debian/Ubuntu (.deb)](#debianubuntu-deb)
    - [Red Hat/CentOS/Fedora (.rpm)](#red-hatcentosfedora-rpm)
    - [macOS (.pkg)](#macos-pkg)
    - [Windows (.exe via InstallForge)](#windows-exe-via-installforge)
  - [Using Binary Archives](#using-binary-archives)
    - [Linux/macOS (.tgz)](#linuxmacos-tgz)
    - [Windows (.exe)](#windows-exe)
  - [Manual Upgrade](#manual-upgrade)
- [Upgrade in Existing Environments](#upgrade-in-existing-environments)
  - [Assess Compatibility](#assess-compatibility)
  - [Upgrade tfvenv Binary](#upgrade-tfvenv-binary)
  - [Update Configuration](#update-configuration)
  - [Verify Functionality](#verify-functionality)
- [Post-Upgrade Verification](#post-upgrade-verification)
  - [Check Version](#check-version)
  - [Test Operations](#test-operations)
- [Troubleshooting](#troubleshooting)

## 2. Pre-Upgrade Steps

### Backup Configurations
- **Backup .tfvenvrc**:
    ```bash
    cp ~/.tfvenvrc ~/.tfvenvrc.backup
    ```

- **Backup Environments**:
    ```bash
    cp -r ~/tfvenv/environments ~/tfvenv/environments.backup
    ```

- **Backup Logs (Optional)**:
    ```bash
    cp ~/tfvenv/logs/tfvenv.log ~/tfvenv/logs/tfvenv.log.backup
    ```

### Check Current Version
```bash
tfvenv --version
# Example Output:
# tfvenv version 1.2.3
```

### Review Release Notes
Visit the tfvenv Releases Page to understand changes and potential impacts.

## 3. Upgrade Methods

### Using Package Managers

#### Debian/Ubuntu (.deb)
- **Download .deb Package**:  
  From Releases Page, download `tfvenv_1.3.0_amd64.deb`.

- **Install Package**:
    ```bash
    sudo dpkg -i tfvenv_1.3.0_amd64.deb
    ```

- **Fix Dependencies**:
    ```bash
    sudo apt-get install -f
    ```

#### Red Hat/CentOS/Fedora (.rpm)
- **Download .rpm Package**:  
  From Releases Page, download `tfvenv-1.3.0-1.x86_64.rpm`.

- **Install Package**:
    ```bash
    sudo rpm -Uvh tfvenv-1.3.0-1.x86_64.rpm
    ```

- Or using yum/dnf:
    ```bash
    sudo yum update tfvenv-1.3.0-1.x86_64.rpm
    # or
    sudo dnf upgrade tfvenv-1.3.0-1.x86_64.rpm
    ```

#### macOS (.pkg)
- **Download .pkg Installer**:  
  From Releases Page, download `tfvenv-1.3.0.pkg`.

- **Run Installer**:  
  Double-click the `.pkg` file and follow the prompts.

#### Windows (.exe via InstallForge)
- **Download .exe Installer**:  
  From Releases Page, download `tfvenv-1.3.0.exe`.

- **Run Installer**:  
  Double-click the `.exe` file and follow InstallForge prompts.

### Using Binary Archives

#### Linux/macOS (.tgz)
- **Download .tgz Archive**:  
  From Releases Page, download `tfvenv-1.3.0-linux-amd64.tgz`.

- **Extract Archive**:
    ```bash
    tar -xzvf tfvenv-1.3.0-linux-amd64.tgz
    ```

- **Move Binary to PATH**:
  - **System-wide**:
    ```bash
    sudo mv tfvenv /usr/local/bin/
    ```
  - **User-specific**:
    ```bash
    mv tfvenv ~/bin/
    ```

- **Make Executable**:
    ```bash
    sudo chmod +x /usr/local/bin/tfvenv
    ```

#### Windows (.exe)
- **Download .exe Binary**:  
  From Releases Page, download `tfvenv-1.3.0.exe`.

- **Replace Existing Binary**:  
  Navigate to installation directory (e.g., `C:\Program Files	fvenv\`) and replace `tfvenv.exe`.

- **Verify Permissions**:  
  Ensure `tfvenv.exe` has execution permissions.

### Manual Upgrade
- **Download Latest Binary**:  
  From Releases Page, download the appropriate binary.

- **Deactivate Environments (Optional)**:
    ```bash
    tfvenv deactivate <env-name>
    ```

- **Replace Binary**:
    ```bash
    sudo mv ~/Downloads/tfvenv /usr/local/bin/
    sudo chmod +x /usr/local/bin/tfvenv
    ```

- **Verify Upgrade**:
    ```bash
    tfvenv --version
    # Should display the new version
    ```

## 4. Upgrade in Existing Environments

### Assess Compatibility
- **Review Release Notes**: Check for breaking changes.
- **Check Configurations**: Ensure `.tfvenvrc` and other configs are compatible.

### Upgrade tfvenv Binary
Follow the appropriate Upgrade Methods based on your installation.

### Update Configuration
- **Review .tfvenvrc**:  
  Update any new configuration options.

- **Set Environment Variables**:
    ```bash
    export TFVENV_NEW_VAR="value"
    ```

### Verify Functionality
- **Activate Environment**:
    ```bash
    tfvenv activate <env-name>
    ```

- **Check Tool Versions**:
    ```bash
    terraform --version
    terragrunt --version
    ```

- **Run Validation**:
    ```bash
    tfvenv validate --env <env-name> --env-type <type>
    ```

## 5. Post-Upgrade Verification

### Check Version
```bash
tfvenv --version
# Expected: tfvenv version 1.3.0
```

### Test Operations
- **List Environments**:
    ```bash
    tfvenv list
    ```

- **Activate/Deactivate**:
    ```bash
    tfvenv activate <env-name>
    tfvenv deactivate <env-name>
    ```

- **Run Terraform Commands**:
    ```bash
    terraform plan
    ```

## 6. Troubleshooting

### Upgrade Fails Due to Package Conflicts
- **Cause**: Architecture mismatch.
- **Solution**: Download the correct package for your system.

### tfvenv Command Not Found After Upgrade
- **Cause**: Binary not in PATH.
- **Solution**: Add tfvenv directory to PATH.
    ```bash
    export PATH="$PATH:/usr/local/bin"
    ```

- **Reload shell**:
    ```bash
    source ~/.bashrc
    ```

### Environment Activation Issues
- **Cause**: Misconfigured `.tfvenvrc` or permissions.
- **Solution**:
    ```bash
    chmod +x /path/to/env/bin/activate.sh
    ```

- **Check `.tfvenvrc` for errors**.

### Tool Versions Not Reflecting Upgrade
- **Cause**: Multiple installations or environment not activated.
- **Solution**:
    ```bash
    which terraform
    # Ensure it points to tfvenv-managed binary
    ```

### Configuration Validation Errors
- **Cause**: Incompatible configurations.
- **Solution**:
    ```bash
    tfvenv validate --env <env-name> --env-type <type>
    ```

### Consult Logs for Detailed Information
- **Location**:
  - **Linux/macOS**: `~/.tfvenv/logs/`
  - **Windows**: `%USERPROFILE%\.tfvenv\logs\`

- **View Logs**:
    ```bash
    cat ~/.tfvenv/logs/tfvenv.log
    ```
