
# tfvenv User Manual

Welcome to tfvenv, a comprehensive command-line tool designed to streamline the management of Terraform and Terragrunt virtual environments. Whether you're working on multiple projects, collaborating with teams, or ensuring consistent infrastructure deployments, tfvenv offers a robust solution to handle your environment needs efficiently.

## Table of Contents
- Introduction
- Feature Overview
- Getting Started
- Prerequisites
- Installation
- Creating a New Environment
- Activating an Environment
- Deactivating an Environment
- Command Reference
  - Environment Management Commands
    - Create
    - Delete
    - List
  - Activation Commands
    - Activate
    - Deactivate
    - Switch
  - Tool Management Commands
    - Install Terraform
    - Install Terragrunt
    - Upgrade
  - Validation and Formatting Commands
    - Validate
    - HCL Format
  - Configuration Management Commands
    - Merge
    - Lock
    - Unlock
  - Snap Management Commands
    - Save Snap
    - Get Snap
    - Update Snap
    - Remove Snap
    - Remote Snap Configuration
    - Remote Snap Operations
  - Utility Commands
    - Cleanup
    - Status
    - List Versions
    - Shell Completions
- Configuration Files
- Best Practices
- Troubleshooting
- Frequently Asked Questions (FAQ)
- Contributing
- License

## Introduction
tfvenv is a dedicated tool for managing virtual environments tailored for Terraform and Terragrunt. It simplifies the setup, maintenance, and activation of isolated environments, ensuring that each project or team operates with the precise tooling versions and configurations required. By encapsulating dependencies and configurations, tfvenv promotes consistency, reduces conflicts, and enhances productivity in infrastructure as code workflows.

## Feature Overview
tfvenv encompasses a variety of features designed to simplify and enhance the management of Terraform and Terragrunt environments:

- **Environment Isolation**: Create isolated environments with specific tool versions.
- **Activation & Deactivation**: Seamlessly switch between environments, ensuring correct configurations.
- **Tool Management**: Install, upgrade, and manage Terraform and Terragrunt binaries within environments.
- **Validation & Formatting**: Validate Terraform configurations and format HCL files for consistency.
- **Configuration Management**: Merge and manage configuration changes without overwriting user customizations.
- **Snap Management**: Save and restore environment states locally and remotely, facilitating collaboration and recovery.
- **Utility Functions**: Cleanup plugin caches, check environment statuses, and list available tool versions.
- **Shell Completions**: Enhance command-line efficiency with auto-completion support for various shells.
- **Comprehensive Logging**: Detailed logs for monitoring and troubleshooting.

## Getting Started
Begin your journey with tfvenv by setting up your environment and managing your first virtual environment. Below are the fundamental steps to get you up and running.

### Prerequisites
- **AWS Credentials (Optional)**: If you plan to use remote snap storage via S3, ensure your AWS credentials are configured.

### Installation
To install tfvenv, follow the documentation for your installation type. Installation documentation can be found at: github.com/rickcollette/tfvenv/documentation

### Creating a New Environment
To create a new virtual environment for Terraform and Terragrunt:

Run the Create Command:

```shell
tfvenv create <env-name> [tf-version] [tg-version]
```
- `<env-name>`: (Required) The name of the environment (e.g., dev, prod).
- `[tf-version]`: (Optional) Specify a Terraform version. Defaults to the latest available version.
- `[tg-version]`: (Optional) Specify a Terragrunt version. Use `none` to skip installing Terragrunt.

**Example**:

```shell
tfvenv create dev 1.0.0 0.35.0
```

This command creates a `dev` environment with Terraform version `1.0.0` and Terragrunt version `0.35.0`.

**NOTE:** If you want to see the last 5 versions of terraform and terragrunt so that you can better make your selection, run this tfvenv command:  
```shell
tfvenv list-versions  
```


### Activating an Environment
To activate a specific environment and configure your shell for its settings:

Run the Activate Command:

```shell
tfvenv activate <env-name>
```
- `<env-name>`: (Required) The name of the environment you wish to activate.

**Example**:

```shell
tfvenv activate dev
```

#### Follow On-Screen Instructions:

After running the command, you'll receive instructions to source the activation script in your current shell:

```shell
source /path/to/env/bin/activate.sh
```

Executing this script sets up the necessary environment variables and updates your PATH to include the tools specific to the activated environment.

### Deactivating an Environment
To revert your shell to its previous state and deactivate the active environment:

Run the Deactivate Command:

```shell
tfvenv deactivate <env-name>
```
- `<env-name>`: (Required) The name of the environment you wish to deactivate.

**Example**:

```shell
tfvenv deactivate dev
```

#### Follow On-Screen Instructions:

After running the command, you'll receive instructions to source the deactivation script:

```shell
source /path/to/env/bin/deactivate.sh
```

Executing this script restores your original environment settings, removing environment-specific variables and tools from your PATH.

## Command Reference
tfvenv offers a suite of commands to manage environments, tools, configurations, and more. Below is a comprehensive overview of each command, including its purpose and usage.

### Environment Management Commands

#### Create
**Description**:
Creates a new virtual environment with specified versions of Terraform and Terragrunt.

**Usage**:

```shell
tfvenv create <env-name> [tf-version] [tg-version]
```
- `<env-name>`: Name of the environment (e.g., dev, prod).
- `[tf-version]`: (Optional) Specific Terraform version to install. Defaults to the latest.
- `[tg-version]`: (Optional) Specific Terragrunt version to install. Use `none` to skip installing Terragrunt.

**Example**:

```shell
tfvenv create staging 1.0.0 0.35.0
```

#### Delete
**Description**:
Deletes an existing virtual environment, removing all associated configurations and tools.

**Usage**:

```shell
tfvenv delete --env <env-name>
```
- `--env <env-name>`: Specifies the name or directory of the environment to delete.

**Example**:

```shell
tfvenv delete --env staging
```

#### List
**Description**:
Lists all managed environments within the specified environment directory.

**Usage**:

```shell
tfvenv list --env <env-directory>
```
- `--env <env-directory>`: (Optional) Specifies the base directory where environments are managed. Defaults to the current directory.

**Example**:

```shell
tfvenv list --env ~/tfvenv/environments
```

### Activation Commands

#### Activate
**Description**:
Activates a virtual environment, configuring your shell with the environment's settings.

**Usage**:

```shell
tfvenv activate <env-name>
```
- `<env-name>`: Name of the environment to activate.

**Example**:

```shell
tfvenv activate dev
```

#### Deactivate
**Description**:
Deactivates the currently active environment, restoring your shell to its previous state.

**Usage**:

```shell
tfvenv deactivate <env-name>
```
- `<env-name>`: Name of the environment to deactivate.

**Example**:

```shell
tfvenv deactivate dev
```

#### Switch
**Description**:
Switches to a different Terraform environment by name or full path. Supports reverting to the previous environment.

**Usage**:

```shell
tfvenv switch <new-env-name|full-path|previous>
```
- `<new-env-name|full-path|previous>`: The target environment name, its absolute path, or `previous` to switch back to the last active environment.

**Example**:

```shell
tfvenv switch prod
```
```shell
tfvenv switch /absolute/path/to/env
```
```shell
tfvenv switch previous
```

## Tool Management Commands

### Install Terraform
**Description**:
Installs a specific version of Terraform into the environment's binary directory.

**Usage**:

```shell
tfvenv install-terraform --version <tf-version>
```
- `--version <tf-version>`: (Optional) Specifies the Terraform version to install. Defaults to the latest version.

**Example**:

```shell
tfvenv install-terraform --version 1.2.3
```

### Install Terragrunt
**Description**:
Installs a specific version of Terragrunt into the environment's binary directory.

**Usage**:

```shell
tfvenv install-terragrunt --version <tg-version>
```
- `--version <tg-version>`: (Optional) Specifies the Terragrunt version to install. Defaults to the latest version.

**Example**:

```shell
tfvenv install-terragrunt --version 0.35.0
```

### Upgrade
**Description**:
Upgrades Terraform and Terragrunt binaries to specified versions within the environment.

**Usage**:

```shell
tfvenv upgrade --tf-version <tf-version> --tg-version <tg-version> --env <env-directory>
```
- `--tf-version <tf-version>`: (Optional) Specifies the Terraform version to upgrade to. Defaults to the latest.
- `--tg-version <tg-version>`: (Optional) Specifies the Terragrunt version to upgrade to. Defaults to the latest.
- `--env <env-directory>`: (Optional) Specifies the base directory of environments. Defaults to the current directory.

**Example**:

```shell
tfvenv upgrade --tf-version 1.3.0 --tg-version 0.36.0 --env ~/tfvenv/environments
```

## Validation and Formatting Commands

### Validate
**Description**:
Validates `.tfvars` and `terragrunt.hcl` files within the environment using Terraform and Terragrunt validators.

**Usage**:

```shell
tfvenv validate --env <env-directory> --env-type <env-type>
```
- `--env <env-directory>`: (Required) Specifies the environment directory.
- `--env-type <env-type>`: (Optional) Specifies the environment type (e.g., dev, prod). Defaults to dev.

**Example**:

```shell
tfvenv validate --env ~/tfvenv/environments/dev --env-type dev
```

### HCL Format
**Description**:
Formats or checks `.hcl` files within the environment using Terragrunt's `hclfmt`.

**Usage**:

```shell
tfvenv hclfmt --env <env-directory> --env-type <env-type> [--check]
```
- `--env <env-directory>`: (Required) Specifies the environment directory.
- `--env-type <env-type>`: (Optional) Specifies the environment type (e.g., dev, prod). Defaults to dev.
- `--check`: (Optional) Checks formatting without making changes.

**Example**:

```shell
tfvenv hclfmt --env ~/tfvenv/environments/dev --env-type dev --check
```

## Configuration Management Commands

### Merge
**Description**:
Merges changes from template configurations into the environment without overwriting user modifications.

**Usage**:

```shell
tfvenv merge --env <env-directory> --env-type <env-type>
```
- `--env <env-directory>`: (Required) Specifies the environment directory.
- `--env-type <env-type>`: (Optional) Specifies the environment type (e.g., dev, prod). Defaults to dev.

**Example**:

```shell
tfvenv merge --env ~/tfvenv/environments/dev --env-type dev
```

### Lock
**Description**:
Locks the environment to prevent concurrent operations, ensuring safe modifications.

**Usage**:

```shell
tfvenv lock --env <env-directory>
```
- `--env <env-directory>`: (Required) Specifies the environment directory.

**Example**:

```shell
tfvenv lock --env ~/tfvenv/environments/dev
```

### Unlock
**Description**:
Unlocks the environment, allowing operations to proceed.

**Usage**:

```shell
tfvenv unlock --env <env-directory>
```
- `--env <env-directory>`: (Required) Specifies the environment directory.

**Example**:

```shell
tfvenv unlock --env ~/tfvenv/environments/dev
```

## Snap Management Commands
Snaps are snapshots of your environment's state, allowing you to save, retrieve, update, and manage environments both locally and remotely.

### Save Snap
**Description**:
Saves the current environment state to a snap file locally.

**Usage**:

```shell
tfvenv snap save <filename>
```
- `<filename>`: (Required) The name of the snap file to save.

**Example**:

```shell
tfvenv snap save dev.snap
```

### Get Snap
**Description**:
Retrieves a snap from local storage and decrypts it.

**Usage**:

```shell
tfvenv snap get <snap-name>
```
- `<snap-name>`: (Required) The name of the snap file to retrieve.

**Example**:

```shell
tfvenv snap get dev.snap
```

### Update Snap
**Description**:
Updates an existing snap file with the current environment state.

**Usage**:

```shell
tfvenv snap update <filename>
```
- `<filename>`: (Required) The name of the snap file to update.

**Example**:

```shell
tfvenv snap update dev.snap
```

### Remove Snap
**Description**:
Removes a snap file from local storage.

**Usage**:

```shell
tfvenv snap remove <filename>
```
- `<filename>`: (Required) The name of the snap file to remove.

**Example**:

```shell
tfvenv snap remove dev.snap
```

## Remote Snap Configuration
**Description**:
Configures remote snap settings, specifically for S3 storage.

**Usage**:

```shell
tfvenv snap config
```

**Example**:

```shell
tfvenv snap config
```

**Notes**:

Ensure that the following environment variables are set:
- `REMOTE_SNAP_ENDPOINT`
- `REMOTE_SNAP_AUTH`
- `REMOTE_SNAP_TYPE` (currently only S3 is supported)

## Remote Snap Operations
Manage snaps stored remotely in S3.

### Remote Snap Save
**Description**:
Saves a snap to remote S3 storage.

**Usage**:

```shell
tfvenv snap save <snap-name>
```
- `<snap-name>`: (Required) The name of the snap to save remotely.

**Example**:

```shell
tfvenv snap save dev.snap
```

### Remote Snap Get
**Description**:
Retrieves a snap from remote S3 storage, decrypts it, and saves it locally.

**Usage**:

```shell
tfvenv snap get <snap-name>
```
- `<snap-name>`: (Required) The name of the snap to retrieve from S3.

**Example**:

```shell
tfvenv snap get dev.snap
```

### Remote Snap List
**Description**:
Lists all snaps available in the remote S3 bucket.

**Usage**:

```shell
tfvenv snap list
```

**Example**:

```shell
tfvenv snap list
```

### Remote Snap Remove
**Description**:
Removes a snap from remote S3 storage.

**Usage**:

```shell
tfvenv snap remove <snap-name>
```
- `<snap-name>`: (Required) The name of the snap to remove from S3.

**Example**:

```shell
tfvenv snap remove dev.snap
```

## Utility Commands

### Cleanup
**Description**:
Cleans up duplicate or unused provider plugins from the plugin cache, optimizing storage and performance.

**Usage**:

```shell
tfvenv cleanup --env <env-directory> --env-dir <environment-directory>
```
- `--env <env-directory>`: (Optional) Specifies the environment directory. Defaults to the current directory.
- `--env-dir <environment-directory>`: (Optional) Specifies the base environment directory.

**Example**:

```shell
tfvenv cleanup --env ~/tfvenv/environments/dev
```

### Status
**Description**:
Displays the current status of the environment, including installed tools and active environment variables.

**Usage**:

```shell
tfvenv status --env <env-directory> --env-type <env-type>
```
- `--env <env-directory>`: (Required) Specifies the environment directory.
- `--env-type <env-type>`: (Optional) Specifies the environment type (e.g., dev, prod). Defaults to dev.

**Example**:

```shell
tfvenv status --env ~/tfvenv/environments/dev --env-type dev
```

### List Versions
**Description**:
Lists the last 5 versions of Terraform and Terragrunt available.

**Usage**:

```shell
tfvenv list-versions
```

**Example**:

```shell
tfvenv list-versions
```

## Shell Completions

### Completion
**Description**:
Generates shell completion scripts for bash, zsh, fish, and PowerShell.

**Usage**:

```shell
tfvenv completion <shell>
```
- `<shell>`: (Required) The type of shell (bash, zsh, fish, powershell).

**Example**:

```shell
tfvenv completion bash > /etc/bash_completion.d/tfvenv
```

**Notes**:

Follow on-screen instructions for integrating the completion scripts into your shell environment.

## Configuration Files
tfvenv uses configuration files to manage environment settings and tool versions. The primary configuration file is `.tfvenvrc`, typically located within the environment's configuration directory.

### .tfvenvrc
This file contains key-value pairs that define the environment's settings.

**Example**:

```makefile
TF_VERSION=1.0.0
TG_VERSION=0.35.0
S3_STATE_BUCKET=your_s3_state_bucket
S3_STATE_PATH=your_s3_state_path
REGION=your_aws_region
ACCESS_KEY=your_access_key
SECRET_KEY=your_secret_key
REMOTE_SNAP_ENDPOINT=your_remote_snap_endpoint
REMOTE_SNAP_AUTH=your_remote_snap_auth
REMOTE_SNAP_TYPE=S3
ENV_VARS=VAR1=value1,VAR2=value2
```

**Fields**:
- `TF_VERSION`: Specifies the Terraform version.
- `TG_VERSION`: Specifies the Terragrunt version.
- `S3_STATE_BUCKET`: The S3 bucket for Terraform state.
- `S3_STATE_PATH`: The path within the S3 bucket for state files.
- `REGION`: AWS region for S3.
- `ACCESS_KEY`: AWS access key for S3 operations.
- `SECRET_KEY`: AWS secret key for S3 operations.
- `REMOTE_SNAP_ENDPOINT`: Endpoint for remote snap storage.
- `REMOTE_SNAP_AUTH`: Authentication method for remote snaps.
- `REMOTE_SNAP_TYPE`: Type of remote storage (currently S3).
- `ENV_VARS`: Additional environment variables in `KEY=value` format, separated by commas.

## Best Practices
- **Consistent Naming**: Use descriptive and consistent names for environments to avoid confusion.
- **Version Pinning**: Specify exact tool versions to ensure reproducibility across teams and deployments.
- **Regular Cleanup**: Periodically run the cleanup command to manage plugin caches and storage efficiently.
- **Secure Credentials**: Protect AWS credentials and other sensitive information, especially when using remote snap storage.
- **Environment Locking**: Use lock and unlock commands to prevent concurrent modifications, ensuring environment integrity.
- **Backup Snaps**: Regularly save snaps, especially before making significant changes, to facilitate easy recovery.

## Troubleshooting
Encountering issues with tfvenv? Here are common problems and their solutions:

### Environment Not Activating
- **Issue**: Activation script not found or not executable.
- **Solution**: Ensure that the activation script exists in the environment's bin directory and has execute permissions.

```shell
chmod +x /path/to/env/bin/activate.sh
```

### Tool Installation Failures
- **Issue**: Failed to download or install Terraform/Terragrunt.
- **Solution**: Check your internet connection, ensure that the specified version exists, and verify AWS credentials if using remote storage.

### Snap Encryption Errors
- **Issue**: Errors related to snap encryption or decryption.
- **Solution**: Ensure that the `SNAP_KEY` environment variable is set correctly and is 32 bytes long for AES-256 encryption.

### Remote Snap Issues
- **Issue**: Unable to save or retrieve snaps from S3.
- **Solution**: Verify AWS credentials, check S3 bucket permissions, and ensure that the remote snap configuration is correctly set.

### HCL Formatting Errors
- **Issue**: `hclfmt` command fails or does not format files as expected.
- **Solution**: Ensure that Terragrunt is correctly installed and accessible in the environment. Check for syntax errors in HCL files.

## Frequently Asked Questions (FAQ)

### Q1: How do I list all available environments?
**A1**: Use the list command:

```shell
tfvenv list --env ~/tfvenv/environments
```

### Q2: Can I manage snaps remotely using providers other than AWS S3?
**A2**: Currently, tfvenv supports remote snap storage via AWS S3. Support for additional providers may be added in future releases.

### Q3: How do I specify custom environment variables during activation?
**A3**: Use the `--var` flag with the activate command to pass custom environment variables in `KEY=value` format, separated by commas:

```shell
tfvenv activate dev --var VAR1=value1,VAR2=value2
```

### Q4: What happens if I try to create an environment with a name that already exists?
**A4**: tfvenv will notify you that the environment already exists and prevent overwriting existing configurations unless explicitly specified.

### Q5: How can I revert to a previous environment?
**A5**: Use the switch command with the `previous` keyword:

```shell
tfvenv switch previous
```

## Contributing
We welcome contributions to tfvenv! Whether you're reporting bugs, suggesting features, or submitting pull requests, your involvement helps make tfvenv better for everyone.

### Steps to Contribute:

1. **Fork the Repository**: Click the "Fork" button on the tfvenv GitHub repository page.
2. **Clone Your Fork**:

```shell
git clone https://github.com/your-username/tfvenv.git
```

3. **Create a New Branch**:

```shell
git checkout -b feature/your-feature-name
```

4. **Make Your Changes**: Implement your feature or fix.
5. **Commit Your Changes**:

```shell
git commit -m "Add feature: your feature description"
```

6. **Push to Your Fork**:

```shell
git push origin feature/your-feature-name
```

7. **Open a Pull Request**: Navigate to the original tfvenv repository and open a pull request from your fork.

Please ensure that your code adheres to the project's coding standards and includes appropriate tests.

## License
tfvenv is released under the MIT License. You are free to use, modify, and distribute this software in accordance with the license terms.

For further assistance or to report issues, please visit the tfvenv GitHub Repository.

