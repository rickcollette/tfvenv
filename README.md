# tfvenv

`tfvenv` is a command-line tool designed to manage isolated virtual environments for Terraform and Terragrunt. Whether you're juggling multiple projects or ensuring consistent infrastructure deployments, `tfvenv` streamlines the setup, maintenance, and activation of your environments, keeping your Terraform and Terragrunt tooling versions organized and conflict-free.

This is an independant open source project. We are in no way affiliated with HashiCorp(tm) or Gruntwork.io.  

## Key Features

- **Environment Isolation**: Create and manage isolated environments with specific versions of Terraform and Terragrunt.
- **Tool Management**: Easily install, upgrade, and manage Terraform and Terragrunt binaries in your environments.
- **Snap Management**: Save and restore environment states both locally and remotely (via S3), facilitating collaboration and backups.
- **Validation and Formatting**: Validate Terraform configurations and ensure consistent formatting of HCL files.
- **Comprehensive Logging**: Keep track of all operations, troubleshooting, and debugging with detailed logs.

## Getting Started

To get started with `tfvenv`, follow these steps:

1. **Installation**: For installation instructions, refer to the [Installation Guide](documentation/2_installation_guide.md).
2. **Creating Environments**: Learn how to create and manage environments in the [User Manual](documentation/1_user_manual.md).
3. **Building and Installing**: See the [Build and Install Guide](documentation/3_build_install_guide.md) for building `tfvenv` from source and setting it up in your system.
4. **Quickstart with Binary**: If you prefer a quick setup, the [Binary TGZ Quickstart](documentation/9_binary_tgz_quickstart.md) will guide you through using a pre-built binary.

## Documentation

All documentation for `tfvenv` can be found in the [documentation](https://github.com/rickcollette/tfvenv/tree/main/documentation) directory of this repository, including:

- [User Manual](documentation/1_user_manual.md): Complete guide to using `tfvenv`.
- [Installation Guide](documentation/2_installation_guide.md): Instructions on how to install `tfvenv`.
- [Build and Install Guide](documentation/3_build_install_guide.md): Step-by-step guide for building from source.
- [Upgrade Guide](documentation/4_upgrade_guide.md): Learn how to upgrade `tfvenv` or the environments.
- [Backup & Restore Guide](documentation/5_backup_restore_guide.md): Keep your environments safe with proper backups and restoration processes.
- [Moving Your Environment](documentation/6_moving_your_environment.md): How to move your `tfvenv` environment across systems.
- [Migrating Between Terraform Versions](documentation/7_migrate_from_one_tf_version_to_another.md): Transitioning between Terraform versions in your environments.
- [Using Docker](documentation/8_using_docker.md): Learn how to use `tfvenv` with Docker.
- [Environment Variables](documentation/10_environment_variables.md): A very deep dive into the use of environment variables.  This also includes a primer of using secrets managment (Vault, Google, AWS).  

## Contributing

We are actively seeking contributions from developers passionate about Terraform, Terragrunt, and infrastructure management! Whether you're looking to contribute code, submit feedback, or help with documentation, we'd love to have you on board. This project is a labor of love, and any help is appreciated.

### How to Contribute:

1. **Fork this Repository**: Click the "Fork" button at the top right of this page.
2. **Clone Your Fork**:
   ```bash
   git clone https://github.com/your-username/tfvenv.git
   cd tfvenv
   ```
3. **Create a Branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```
4. **Make Your Changes**.
5. **Commit Your Changes**:
   ```bash
   git commit -m "Add feature: your-feature-description"
   ```
6. **Push to Your Fork**:
   ```bash
   git push origin feature/your-feature-name
   ```
7. **Open a Pull Request**: From your forked repository, open a pull request into this repo's main branch.

Please follow our contribution guidelines and ensure all changes are thoroughly tested before submitting your pull request.

## Support

If you encounter any issues or have questions, please use the [GitHub Issues](https://github.com/rickcollette/tfvenv/issues) page to report bugs or ask for help. 

We aim to provide excellent community support, and your feedback will help us improve `tfvenv`!

## License

This project is licensed under the MIT License. For more details, see the [LICENSE](LICENSE) file.