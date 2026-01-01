<!-- TODO: filename is horrible, change it  -->
# Environment

First define a folder where to store the notes.

Then in your bashrc set :

```bash
export K_DIR="$HOME/notes"
```

## Configuration file

Create the file `~/.config/kl/config.json`.

Fill it like following and change the fields as you need:


| Field                  | Required   | Default                  | Description                                                                                              |
| -------                | ---------- | ---------                | -------------                                                                                            |
| **dirExport**          | Optional   | "/tmp/knowledge-export/" | Directory where exported notes will be saved                                                             |
| **joplinConfigFile**   | Optional   | -                        | Path to Joplin configuration file for cloud sync                                                         |
| **fzf**                | Optional   | false                    | Enable/disable fuzzy finder integration                                                                  |
| **fzfEnv**             | Optional   | "fzf"                    | Environment for fzf: `"fzf"` (standard) or `"fzf-tmux"` (tmux popup)                                     |
| **fzfOption**          | Optional   | FZF_DEFAULT_OPTS         | Options for fzf. Defaults to `FZF_DEFAULT_OPTS` environment variable                                     |
| **fzfTmuxOption**      | Optional   | FZF_TMUX_OPTS            | Options for fzf-tmux. Defaults to `FZF_TMUX_OPTS` environment variable                                   |
| **terminalParams**     | Optional   | ["--execute"]            | Parameters passed to the terminal emulator                                                               |
| **screenshotTool**     | Optional   | "flameshot"              | Tool used for taking screenshots                                                                         |
| **screenshotParams**   | Optional   | ["gui", "--path"]        | Parameters passed to the screenshot tool                                                                 |
| **screenrecordTool**   | Optional   | -                        | Tool used for screen recording. See [video recording guide](./video_recording.md)                        |
| **screenrecordParams** | Optional   | -                        | Parameters passed to the screen recording tool                                                           |
| **schemaTool**         | Optional   | "inkscape"               | Tool used for creating diagrams/schemas                                                                  |
| **schemaToolParams**   | Optional   | ["--export-filename"]    | Parameters passed to the schema tool                                                                     |
| **terminal**           | Optional   | "x-terminal-emulator"    | Terminal emulator to use                                                                                 |

NB: The configuration file is entirely optional. If not present, kl will use sensible defaults.

## Example of configuration file

```json
{
  "fzf": true,
  "fzfEnv": "fzf-tmux",
  "fzfOption": "--height 40% --reverse --border",
  "fzfTmuxOption": "--height 80% --border",
  "screenshotTool": "flameshot",
  "screenshotParams": ["gui", "--path"],
  "screenrecordTool": "screenrecord",
  "screenrecordParams": [],
  "schemaTool": "inkscape",
  "schemaToolParams": ["--export-filename"],
  "terminal": "x-terminal-emulator",
  "terminalParams": ["--execute"],
  "joplinConfigFile": "/home/user/.config/joplin-desktop/settings.json",
  "dirExport": "/tmp/knowledge-export/"
}
```
