
# Backup and Restore Your tfvenv Environment

Managing Terraform and Terragrunt with tfvenv requires effective backup and restoration to ensure your configurations are safe and recoverable. This guide provides simple steps for backing up and restoring tfvenv environments on Linux/macOS and Windows.

## Contents
- [Understanding tfvenv](#understanding-tfvenv)
- [Backup Methods](#backup-methods)
- [Backup Steps](#backup-steps)
- [Restore Steps](#restore-steps)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## 1. Understanding tfvenv
A tfvenv environment includes:
- **Binaries**: Terraform and Terragrunt versions.
- **Configuration Files**: `.tfvenvrc`, `.tfvars`, `terragrunt.hcl`.
- **Plugin Cache**: Terraform providers.
- **Data Directory**: Terraform state files.
- **Activation Scripts**: `activate.sh` / `activate.ps1`.

## 2. Backup Methods
- **Manual Backup**: Copy essential files to a secure location.
- **Automated Scripts**: Schedule scripts to perform backups regularly.
- **Version Control (Git)**: Track and back up configuration files.

## 3. Backup Steps

### Manual Backup

#### Linux/macOS:
1. Open Terminal.
2. Navigate to tfvenv directory:
```bash
cd $TFVENV_PATH
```
3. Compress environment:
```bash
tar -czvf <env-name>_backup_YYYY-MM-DD.tgz environments/<env-name>
```
4. Backup config files:
```bash
cp configurations/<env-name>/.tfvenvrc backups/
```
5. Move to secure location:
```bash
mv <env-name>_backup_*.tgz /path/to/backup/
```

#### Windows:
1. Open PowerShell.
2. Navigate to tfvenv directory:
```powershell
cd $env:TFVENV_PATH
```
3. Compress environment:
```powershell
Compress-Archive -Path environments\<env-name>\* -DestinationPath backups\<env-name>_backup_YYYY-MM-DD.zip
```
4. Backup config files:
```powershell
Copy-Item configurations\<env-name>\.tfvenvrc backups\
```
5. Move to secure location:
```powershell
Move-Item backups\*.zip D:\Backup\tfvenv\
```

### Automated Backup Scripts

#### Linux/macOS (`tfvenv_backup.sh`):
```bash
#!/bin/bash
TFVENV_PATH="$HOME/.tfvenv"
BACKUP_DIR="$HOME/tfvenv_backups"
DATE=$(date +%F)
ENV_NAME="dev"

mkdir -p "$BACKUP_DIR"
tar -czvf "$BACKUP_DIR/${ENV_NAME}_backup_${DATE}.tgz" "$TFVENV_PATH/environments/$ENV_NAME"
cp "$TFVENV_PATH/configurations/$ENV_NAME/.tfvenvrc" "$BACKUP_DIR/"
find "$BACKUP_DIR" -type f -mtime +30 -delete
```

Schedule with Cron:
```bash
crontab -e
# Add: 0 0 * * * /bin/bash ~/tfvenv_backup.sh
```

#### Windows (`tfvenv_backup.ps1`):
```powershell
$TFVENV_PATH = "$env:USERPROFILE\.tfvenv"
$BACKUP_DIR = "$env:USERPROFILE\tfvenv_backups"
$DATE = Get-Date -Format "yyyy-MM-dd"
$ENV_NAME = "dev"

New-Item -ItemType Directory -Path $BACKUP_DIR -Force
Compress-Archive -Path "$TFVENV_PATH\environments\$ENV_NAME\*" -DestinationPath "$BACKUP_DIR\$ENV_NAME`_backup_$DATE.zip" -Force
Copy-Item "$TFVENV_PATH\configurations\$ENV_NAME\.tfvenvrc" "$BACKUP_DIR\" -Force
Remove-Item "$BACKUP_DIR\*" -Recurse -Include *backup* -OlderThan 30
```

Schedule with Task Scheduler:  
Create a task to run the script daily.

### Version Control (Git)
1. Initialize Git:
```bash
cd $TFVENV_PATH/configurations
git init
```
2. Add files:
```bash
git add <env-name>/.tfvenvrc <env-name>/terragrunt.hcl <env-name>/*.tfvars
```
3. Commit and push:
```bash
git commit -m "Backup tfvenv configurations"
git remote add origin https://github.com/yourusername/tfvenv-configurations.git
git push -u origin master
```

## 4. Restore Steps

### Manual Restore

#### Linux/macOS:
1. Open Terminal.
2. Navigate to tfvenv:
```bash
cd $TFVENV_PATH
```
3. Extract backup:
```bash
tar -xzvf /path/to/backup/<env-name>_backup_YYYY-MM-DD.tgz -C environments/
```
4. Restore config:
```bash
cp /path/to/backup/.tfvenvrc configurations/<env-name>/
```
5. Activate environment:
```bash
tfvenv activate <env-name>
source environments/<env-name>/bin/activate.sh
```
6. Initialize Terraform:
```bash
terraform init
```

#### Windows:
1. Open PowerShell.
2. Navigate to tfvenv:
```powershell
cd $env:TFVENV_PATH
```
3. Extract backup:
```powershell
Expand-Archive -Path "D:\Backup\tfvenv\<env-name>_backup_YYYY-MM-DD.zip" -DestinationPath environments\
```
4. Restore config:
```powershell
Copy-Item "D:\Backup\tfvenv\.tfvenvrc" "configurations\<env-name>\" -Force
```
5. Activate environment:
```powershell
tfvenv activate <env-name>
.\environments\<env-name>\bin\activate.ps1
```
6. Initialize Terraform:
```powershell
terraform init
```

### Automated Restore Scripts

#### Linux/macOS (`tfvenv_restore.sh`):
```bash
#!/bin/bash
TFVENV_PATH="$HOME/.tfvenv"
BACKUP_DIR="$HOME/tfvenv_backups"
DATE="YYYY-MM-DD"
ENV_NAME="dev"

tar -xzvf "$BACKUP_DIR/${ENV_NAME}_backup_${DATE}.tgz" -C "$TFVENV_PATH/environments/"
cp "$BACKUP_DIR/.tfvenvrc" "$TFVENV_PATH/configurations/$ENV_NAME/"
echo "Restoration completed."
```

#### Windows (`tfvenv_restore.ps1`):
```powershell
$TFVENV_PATH = "$env:USERPROFILE\.tfvenv"
$BACKUP_DIR = "$env:USERPROFILE\tfvenv_backups"
$DATE = "YYYY-MM-DD"
$ENV_NAME = "dev"

Expand-Archive -Path "$BACKUP_DIR\$ENV_NAME`_backup_$DATE.zip" -DestinationPath "$TFVENV_PATH\environments\" -Force
Copy-Item "$BACKUP_DIR\.tfvenvrc" "$TFVENV_PATH\configurations\$ENV_NAME\" -Force
Write-Output "Restoration completed."
```

### Restore from Git
1. Clone repository:
```bash
cd $TFVENV_PATH/configurations
git clone https://github.com/yourusername/tfvenv-configurations.git <env-name>
```
2. Checkout branch:
```bash
cd <env-name>
git checkout main
```
3. Activate environment and verify:
```bash
tfvenv activate <env-name>
source environments/<env-name>/bin/activate.sh
terraform --version
```

## 5. Best Practices
- **Regular Backups**: Schedule daily or weekly backups.
- **Automate**: Use scripts and schedulers to minimize manual tasks.
- **Version Control**: Track changes with Git for easy recovery and collaboration.
- **Secure Storage**: Store backups in multiple locations (e.g., cloud, external drives).
- **Encrypt Sensitive Data**: Protect backups containing sensitive information.
- **Test Restores**: Periodically verify that backups can be successfully restored.
- **Document Processes**: Keep clear records of backup and restore procedures.

## 6. Troubleshooting

### Incomplete Backups:
- Check backup contents.
- Ensure scripts include all necessary files.
- Verify file permissions.

### Restoration Failures:
- Confirm backup integrity.
- Ensure enough disk space.
- Check write permissions.

### Version Conflicts:
- Reinitialize Terraform:
```bash
terraform init -upgrade
```
- Install correct versions with tfvenv:
```bash
tfvenv install-terraform --version x.x.x
tfvenv install-terragrunt --version x.x.x
```

### Environment Variables Issues:
- Review `.tfvenvrc` for correct settings.
- Re-activate the environment.

### Corrupted Config Files:
- Restore from backup or Git.
- Validate configurations:
```bash
tfvenv validate --env <env-name>
```

For more details, refer to the tfvenv User Manual or join the tfvenv Community for support.
