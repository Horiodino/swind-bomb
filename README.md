---
## Overview

Swinds is a Go-based program that replicates the *Swinds Attack* functionality. Use the `install.sh` script to build, install, and manage the *swinds* binary as a systemd service.

## Usage

Run the `install.sh` script with one of the following commands:

- **`build`**: Build the binary.
  ```bash
  ./install.sh build
  ```

- **`install`**: Build and install the binary.
  ```bash
  ./install.sh install
  ```

- **`create-service`**: Create the systemd service.
  ```bash
  ./install.sh create-service
  ```

- **`setup-service`**: Set up and enable the service.
  ```bash
  ./install.sh setup-service
  ```

- **`status`**: Check the service status and logs.
  ```bash
  ./install.sh status
  ```

- **`clean`**: Clean up the build directory.
  ```bash
  ./install.sh clean
  ```

- **`check`**: Check if the binary is installed.
  ```bash
  ./install.sh check
  ```

## License

Open-source under the [MIT License](LICENSE).

---
