# 📂 Go Drop

**Go Drop** is a fast and simple file-sharing server you can run on your local machine. Once running, anyone connected to the same Wi-Fi or local network can upload, view, and download files through their browser.

> ⚠️ Internet is **not required** — everything happens locally.

## Features

- Upload files up to 100MB
- Shareable URLs for every file (e.g. `http://192.168.x.x:8080/shared/myfile.png`)
- List of all uploaded files with metadata (name, size, timestamp)
- Delete files
- Download via direct links
- CORS enabled
- Auto-sanitized filenames
- Automatically shows your local IP address on startup

## Getting Started

1. **Clone the repo:**

   ```bash
   git clone https://github.com/davidumoru/go-drop
   cd go-drop
   ```

2. **Build the binary:**

   ```bash
   go build -o go-drop
   ```

3. **Run the server:**

   ```bash
   ./go-drop
   ```

4. **Access on local machine:**

   ```text
   http://localhost:8080
   ```

   **Access on other devices (same Wi-Fi):**

   ```text
   http://192.168.X.X:8080
   ```

## Folder Structure

```tree
go-drop/
├── main.go              // Core server logic
├── public/              // Frontend assets (index.html)
├── shared/              // Uploaded files are stored here
```

## Example Use Case

- A teacher runs Go Drop on their laptop and students connect to `http://192.168.0.105:8080` to download assignments.
- You want to send a file from your PC to your phone with no cables or apps.
- Quickly pass files between laptops during a hackathon or office collab.

## Security Notes

- Only accessible within your LAN (local area network)
- No authentication — anyone on the same network can access
- You can run it behind a firewall or add basic auth if needed (future feature idea)

## Future Improvements (TODOs)

- [ ] Password protection or token-based access
- [ ] File size configuration via flags
- [ ] Auto-deletion after X hours
- [ ] File previews (image/audio/video)

## License

[MIT License](LICENSE)
