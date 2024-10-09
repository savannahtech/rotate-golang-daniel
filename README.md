# Project Setup and Usage

go project track file changes on host machine

## Prerequisites

- **[Docker](https://www.docker.com/)**: Ensure that Docker is installed and running on your machine.
- **[Make](https://www.gnu.org/software/make/)**: Ensure that `make` is installed on your system.

## Dependencies

- **[Node.js](https://nodejs.org/en)**
- **[Wails](https://wails.io/docs/introduction)**: Wails is a project that enables you to write desktop apps using Go and web technologies.
- **[Osquery](https://osquery.io/)**: Osquery uses basic SQL commands to leverage a relational data-model to describe a device

### 1. Configuration & Setup

- setup config and osquery on mac

```bash
make setup/osquery/mac
```

- setup mongo db in docker

```bash
make logsdb
```

### 2. start ui

- start osquery on mac

```bash
make start/osqueryd/mac
```

- Build package

```bash
make build/package
```

- Run build

```bash
make run/build/mac
```

OR

- Run in dev mode

```bash
make run/dev/mac
```

### 3. Check heath of of workers

`curl -s -X GET http://localhost:9000/v1/health`

### 4. Add new command to queue

```bash
curl -s -X POST http://localhost:9000/v1/commands \
-H "Content-Type: application/json" \
-d "{\"commands\":[\"touch $HOME/Downloads/test1.txt\"]}"
```

OR create new file manually

```bash
touch $HOME/Downloads/test2.txt
```

### 5. Get logs

- wait 5 seconds and click "Fetch logs" on UI

OR run command below

`curl -s -X GET http://localhost:9000/v1/logs\?limit=2`

---

NOTES

osqueryd flags

- macOS:

  ```bash
  sudo osqueryd --verbose --disable_events=false --disable_audit=false --disable_endpointsecurity=false --disable_endpointsecurity_fim=false --enable_file_events=true
  ```

- Windows: (_as administrator_)

  ```bash
  osqueryd --verbose --disable_events=false --enable_ntfs_event_publisher=true --enable_powershell_events_subscriber=true --enable_windows_events_publisher=true --enable_windows_events_subscriber=true

  ```
