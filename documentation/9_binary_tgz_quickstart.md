
# If you unzipped the binary tar.gz, follow these steps to complete the setup:

---

## Install the Man Pages

Make sure the `install-man.sh` script is executable and run it:

```bash
chmod +x install-man.sh
./install-man.sh
```

## Set Your PATH

You need to add the binary directory to your PATH so that you can run `tfvenv` from anywhere.

### Bash

```bash
export PATH=$PATH:/path/to/tfvenv/usr/local/bin
```

You can add this line to your `.bashrc` or `.bash_profile` for it to persist across sessions:

```bash
echo 'export PATH=$PATH:/path/to/tfvenv/usr/local/bin' >> ~/.bashrc
source ~/.bashrc
```

### Zsh

```zsh
export PATH=$PATH:/path/to/tfvenv/usr/local/bin
```

To make this change persistent, add it to your `.zshrc`:

```zsh
echo 'export PATH=$PATH:/path/to/tfvenv/usr/local/bin' >> ~/.zshrc
source ~/.zshrc
```

### Fish

```fish
set -Ux fish_user_paths /path/to/tfvenv/usr/local/bin $fish_user_paths
```

This adds the directory to the `fish_user_paths`, which Fish uses to store custom paths.

## Read the README.md File

The `README.md` file contains detailed instructions and information about `tfvenv`. It's highly recommended to go through it to understand all features, configurations, and usage.
