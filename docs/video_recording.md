# Video Recording

Unlike screenshots, there are no CLI tools available on all platforms that allow you to:
- record your screen with area selection
- set the path of the file before starting the command line: `recordingtools [path]`

To make it work on my computer I've created the following simple script that works on Ubuntu/Debian.

Don't hesitate to raise an issue if you find the right tool.

## Script

Install the dependencies:

```bash
sudo apt install python-tk   # Needed to create dial box
```

Create the file `screenrecord` and fill it with the following:

```python
#!/bin/python3
import tkinter as tk
import subprocess
import os
import sys
from datetime import datetime

class SimpleRecorderGUI:
    def __init__(self, output_path=None):
        if output_path is None:
            timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
            home_dir = os.path.expanduser("~")  # Python handles ~ expansion
            output_path = os.path.join(home_dir, f"{timestamp}.mp4")

        self.output_path = output_path
        self.root = tk.Tk()
        self.root.title("Screen Recorder")
        self.root.resizable(False, False)
        self.root.geometry("250x120")
        self.setup_ui()

    def setup_ui(self):
        label = tk.Label(self.root, text="Screen Recording Control")
        label.pack(padx=20, pady=10)

        frame = tk.Frame(self.root)
        frame.pack(pady=10)

        start_btn = tk.Button(
            frame, text="Launch Script", command=self.launch_script, width=12
        )
        start_btn.pack(side=tk.LEFT, padx=10)

        stop_btn = tk.Button(
            frame, text="Stop Script", command=self.stop_script, width=12
        )
        stop_btn.pack(side=tk.LEFT, padx=10)

    def is_ffmpeg_running(self):
        try:
            result = subprocess.run(["pidof", "ffmpeg"], capture_output=True)
            return result.returncode == 0
        except:
            return False

    def kill_ffmpeg(self):
        try:
            subprocess.run(["killall", "ffmpeg"], check=False)
            return True
        except:
            return False

    def notify(self, message):
        try:
            subprocess.run(
                ["notify-send", message, "--icon=dialog-information"], check=False
            )
        except:
            pass

    def get_screen_area(self):
        try:
            result = subprocess.run(
                ["slop", "-f", "%x %y %w %h"],
                capture_output=True,
                text=True,
                check=True,
            )
            coords = result.stdout.strip().split()
            if len(coords) == 4:
                return {
                    "x": int(coords[0]),
                    "y": int(coords[1]),
                    "w": int(coords[2]),
                    "h": int(coords[3]),
                }
        except:
            return None

    def start_recording(self, area):
        display = os.environ.get("DISPLAY", ":0")
        cmd = [
            "ffmpeg",
            "-f",
            "x11grab",
            "-s",
            f"{area['w']}x{area['h']}",
            "-framerate",
            "60",
            "-thread_queue_size",
            "512",
            "-i",
            f"{display}.0+{area['x']},{area['y']}",
            "-vcodec",
            "libx264",
            "-qp",
            "18",
            "-preset",
            "ultrafast",
            self.output_path,
        ]
        try:
            subprocess.Popen(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            return True
        except:
            return False

    def launch_script(self):
        if self.is_ffmpeg_running():
            self.stop_script()
        else:
            area = self.get_screen_area()
            if area and area["w"] > 0:
                self.start_recording(area)

    def stop_script(self):
        self.kill_ffmpeg()
        self.notify("Stopped Recording!")

    def run(self):
        self.root.mainloop()

if __name__ == "__main__":
    if len(sys.argv) > 1:
        path = sys.argv[1]
        if not os.path.isabs(path):
            path = os.path.abspath(path)
        if os.path.isdir(path):
            timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
            path = os.path.join(path, f"{timestamp}.mp4")
        app = SimpleRecorderGUI(path)
    else:
        print("No arguments provided. Using timestamp-based filename in home directory.")
        app = SimpleRecorderGUI()

    app.run()
```

To make sure it works, `screenrecord` needs to have the execution right and be located in one of the `$PATH` directories.

Then call the script using the following command:

```bash
screenrecord .

``````
