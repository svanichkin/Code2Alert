# Code2Alert

**Code2Alert** is a minimalist utility that monitors a specified folder and displays metadata (`xattr`) from new files from [email2folder](https://github.com/svanichkin/Email2Folder) in the macOS/Linux/Windows menu bar.

If a file has the `xattr` attribute `type=code`, its `summary` attribute is displayed in the menu bar and copied to the clipboard.

## 🛠 Features

- Monitoring files and subfolders in the specified directory
- Displaying short text in the macOS menu bar (via [systray](https://github.com/getlantern/systray))
- Automatically copying the `summary` content to the clipboard
- Auto-restart when the binary is updated

## 🔧 Installation

### Dependencies

Build the binary:

```bash
go build
```

## ⚙️ Configuration

Create `config.json` next to the binary:

```json
{
  "folder": "/folder/from/email2folder/emails/"
}
```

## 📋 How It Works

1. The program loads the config and begins monitoring the folder and all subfolders.
2. When a new file appears:
   - - Verifies it's a regular file
   - - Reads `xattr` attributes: `type` and `summary`
   - - If `type=code`, shows `summary` in the menu bar and copies it to clipboard
3. - The menu bar title clears after 10 seconds.
4. - If the binary is updated — the app restarts.

---

License: MIT
