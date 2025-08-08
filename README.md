# cSQL

cSQL is a terminal-based user interface (TUI) for working with several types of database instances. It provides a convenient way to browse, connect to, and manage your database instances from the terminal.

## Features

*   **Terminal-based UI:** A user-friendly, terminal-based interface for managing database instances.
*   **Cloud Integration:** Discover and connect to cloud-based database instances (e.g., Google Cloud SQL).
*   **Password Manager Integration:** Supports multiple password managers for authentication.
*   **Persistent Configuration:** Automatically saves and loads your database instance configurations, ensuring your setup is preserved across sessions. 
*   **Extensible:** The application is designed to be extensible, allowing for the addition of new database types, discovery and authentication methods.

### Supported Databases

*   PostgreSQL

### Cloud Provider Auto-Discovery

*   Google Cloud Platform (GCP) for instance discovery

### Password Manager Integration

*   Local (password stored in configuration file)
*   HashiCorp Vault

## Getting Started

### Prerequisites

*   Go 1.23 or later
*   A configured Go environment

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/rhariady/csql.git
    ```
2.  Navigate to the project directory:
    ```bash
    cd csql
    ```
3.  Build the application:
    ```bash
    go build
    ```

### Configuration

cSQL automatically manages its configuration file, which stores your database instance details. The configuration file is named `.csql` and is located in your user's configuration directory:

*   **Linux:** `~/.config/.csql`
*   **macOS:** `~/Library/Application Support/.csql`
*   **Windows:** `%AppData%/.csql`

You can manually edit this file to add, modify, or remove database instances. Here is an example of the `.csql` file content:

```toml
[instances]
  [instances.my-postgres-instance]
    type = "postgresql"
    host = "localhost"
    port = 5432
    source = "manual"
```

### Usage

To run the application, execute the following command:

```bash
./csql
```

The application will display a list of your configured database instances. You can use the arrow keys to navigate the list and press `Enter` to connect to an instance.

### Keybindings

*   `a`: Add a new database instance.
*   `d`: Remove the selected database instance.
*   `<Enter>`: Connect to the selected database instance.
*   `q`: Quit the application.

## Contributing

Contributions are welcome! If you would like to contribute to the project, please fork the repository and submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
