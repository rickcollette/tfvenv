
                                TFVENV Docker
---

This guide provides instructions on how to build and run the Docker image for `tfvenv`, along with various usage scenarios for both Linux/macOS and Windows.

By using the Docker image, you can easily manage isolated `tfvenv` environments with consistent settings and tooling versions. You can also customize the environment through configuration files and environment variables.

                                Prerequisites
---
- Ensure you have Docker installed on your system.
- You need a local directory to mount as `/tfvenvroot/` inside the container.

                                Do the things
---

## 1. Build the Docker Image

Use the following command to build the Docker image from the provided Dockerfile:

```bash
docker build -t tfvenv:latest -f Dockerfile .
```

This command creates a Docker image named `tfvenv` with the `latest` tag.

## 2. Run the Docker Container

To start a container using the built image, use the `docker run` command. You can mount a local directory to `/tfvenvroot/` inside the container.

### For Linux/macOS

```bash
docker run -it --rm -v $(pwd)/your-local-tfvenv-dir:/tfvenvroot tfvenv:latest
```

### For Windows

On Windows, you can use the full path to your local directory and use forward slashes (`/`) or double backslashes (`\`) for the file path. Make sure to specify the drive letter in uppercase (e.g., `C:`).

```powershell
docker run -it --rm -v C:/path/to/your-local-tfvenv-dir:/tfvenvroot tfvenv:latest
```

Or using PowerShell-style paths:

```powershell
docker run -it --rm -v ${PWD}/your-local-tfvenv-dir:/tfvenvroot tfvenv:latest
```

### Explanation
- `-it`: Runs the container in interactive mode with a terminal attached.
- `--rm`: Automatically removes the container when it exits.
- `-v <local-path>:/tfvenvroot`: Mounts the local directory `your-local-tfvenv-dir` to `/tfvenvroot` inside the container.
- `tfvenv:latest`: Specifies the Docker image to use.

## Mounting Local Directories

You can use any local directory for `your-local-tfvenv-dir`, which will be mounted to `/tfvenvroot` inside the container:

- **A Git Cloned Repository**: If you have a Terraform or Terragrunt repository, you can clone it into a local folder and mount that folder when running the Docker container. This allows all your Terraform/Terragrunt configurations to be accessed within the container.

    #### Linux/macOS Example:
    ```bash
    git clone https://github.com/your-repo/terraform-config.git
    docker run -it --rm -v $(pwd)/terraform-config:/tfvenvroot tfvenv:latest
    ```

    #### Windows Example:
    ```powershell
    git clone https://github.com/your-repo/terraform-config.git
    docker run -it --rm -v C:/path/to/terraform-config:/tfvenvroot tfvenv:latest
    ```

    Here, the `terraform-config` directory contains all your Terraform/Terragrunt files and is mounted to `/tfvenvroot` in the Docker container.

- **Any Local Folder**: You can also use any existing folder on your system to store configurations, `.tfvenvrc` files, state files, or other necessary `tfvenv` resources.

    #### Linux/macOS Example:
    ```bash
    docker run -it --rm -v $(pwd)/your-local-tfvenv-dir:/tfvenvroot tfvenv:latest
    ```

    #### Windows Example:
    ```powershell
    docker run -it --rm -v C:/path/to/your-local-tfvenv-dir:/tfvenvroot tfvenv:latest
    ```

    Inside the container, `/tfvenvroot` will mirror the contents of `your-local-tfvenv-dir` on your host system. You can interact with your Terraform and Terragrunt files, run `tfvenv` commands, and manage your environment as if you were working locally, but isolated within the Docker environment.

## 3. Running TFVENV Commands Inside Docker

You can pass any `tfvenv` command to the container.

#### Linux/macOS Example:
```bash
docker run -it --rm -v $(pwd)/your-local-tfvenv-dir:/tfvenvroot tfvenv:latest tfvenv status
```

#### Windows Example:
```powershell
docker run -it --rm -v C:/path/to/your-local-tfvenv-dir:/tfvenvroot tfvenv:latest tfvenv status
```

In this example, the `tfvenv status` command is executed inside the container.

## 4. Connecting to a Running Container

If you want to connect to a running container to manually execute commands or explore the environment, you can use `docker exec`:

1. Start the container (without `--rm` to keep it running in the background):

    #### Linux/macOS:
    ```bash
    docker run -d -v $(pwd)/your-local-tfvenv-dir:/tfvenvroot --name tfvenv_container tfvenv:latest
    ```

    #### Windows:
    ```powershell
    docker run -d -v C:/path/to/your-local-tfvenv-dir:/tfvenvroot --name tfvenv_container tfvenv:latest
    ```

    - `-d`: Runs the container in detached mode (in the background).
    - `--name tfvenv_container`: Names the container `tfvenv_container`.

2. Connect to the running container:

    ```bash
    docker exec -it tfvenv_container /bin/bash
    ```

    This command opens a Bash session inside the container.

3. Upon connecting, the `tfvenv` environment should be sourced if `/tfvenvroot/.tfvenvrc` and `/tfvenvroot/bin/activate.sh` are present.

## 5. Docker Environment Variables for TFVENV

You can customize the behavior of `tfvenv` within the container using environment variables.

### Examples

- **Set Logging File**:
    ```bash
    docker run -it --rm -v $(pwd)/your-local-tfvenv-dir:/tfvenvroot -e TFVENV_LOG_FILE=/tfvenvroot/tfvenv.log tfvenv:latest
    ```

    **Windows**:
    ```powershell
    docker run -it --rm -v C:/path/to/your-local-tfvenv-dir:/tfvenvroot -e TFVENV_LOG_FILE=/tfvenvroot/tfvenv.log tfvenv:latest
    ```

- **Enable Verbose Logging**:
    ```bash
    docker run -it --rm -v $(pwd)/your-local-tfvenv-dir:/tfvenvroot -e TFVENV_VERBOSE=true tfvenv:latest
    ```

    **Windows**:
    ```powershell
    docker run -it --rm -v C:/path/to/your-local-tfvenv-dir:/tfvenvroot -e TFVENV_VERBOSE=true tfvenv:latest
    ```

## 6. Stopping the Container

If you started the container in detached mode (`-d`), you can stop it using:

```bash
docker stop tfvenv_container
```

This stops the container named `tfvenv_container`.
